package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/probe-system/agent/internal/collector"
	"github.com/probe-system/agent/internal/executor"
	"github.com/probe-system/agent/internal/ws"
	"github.com/probe-system/agent/pkg/protocol"
)

const Version = "1.0.0"

type Config struct {
	ServerURL      string `json:"server_url"`
	Token          string `json:"token"`
	MetricInterval int    `json:"metric_interval"`
}

func main() {
	configPath := flag.String("config", "agent.json", "Path to config file")
	serverURL := flag.String("server", "", "Server WebSocket URL")
	token := flag.String("token", "", "Authentication token")
	interval := flag.Int("interval", 10, "Metric collection interval in seconds")
	flag.Parse()

	// Load config
	config := &Config{
		MetricInterval: 10,
	}

	if data, err := os.ReadFile(*configPath); err == nil {
		json.Unmarshal(data, config)
	}

	// Override with flags
	if *serverURL != "" {
		config.ServerURL = *serverURL
	}
	if *token != "" {
		config.Token = *token
	}
	if *interval > 0 {
		config.MetricInterval = *interval
	}

	// Validate
	if config.ServerURL == "" {
		log.Fatal("Server URL is required")
	}

	log.Printf("Probe Agent v%s starting...", Version)
	log.Printf("Server: %s", config.ServerURL)

	// Create components
	client := ws.NewClient(config.ServerURL, config.Token, Version)
	coll := collector.NewCollector(time.Duration(config.MetricInterval) * time.Second)

	// Get script directory
	execPath, _ := os.Executable()
	scriptDir := filepath.Join(filepath.Dir(execPath), "scripts")
	taskMgr := executor.NewTaskManager(config.ServerURL, scriptDir)

	// Handle incoming messages
	client.SetMessageHandler(func(msg *protocol.Message) {
		switch msg.Type {
		case protocol.MsgTypeTaskAssign:
			var payload protocol.TaskAssignPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("Invalid task payload: %v", err)
				return
			}
			log.Printf("Received task: %s (%s)", payload.TaskID, payload.Type)
			taskMgr.HandleTask(&payload)

		case protocol.MsgTypeConfig:
			// Handle config updates
			log.Printf("Received config update")
		}
	})

	// Connect
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	client.Run()

	// Start metric collection
	go func() {
		ticker := time.NewTicker(coll.GetInterval())
		defer ticker.Stop()

		for range ticker.C {
			metrics, err := coll.Collect()
			if err != nil {
				log.Printf("Failed to collect metrics: %v", err)
				continue
			}

			if err := client.SendMetrics(metrics); err != nil {
				log.Printf("Failed to send metrics: %v", err)
			}
		}
	}()

	// Forward task results
	go func() {
		for result := range taskMgr.GetResultChan() {
			if err := client.SendTaskResult(result); err != nil {
				log.Printf("Failed to send task result: %v", err)
			}
		}
	}()

	// Wait for shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	taskMgr.Stop()
	client.Stop()
}
