package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/probe-system/core/internal/config"
	"github.com/probe-system/core/internal/handler"
	"github.com/probe-system/core/internal/notify"
	"github.com/probe-system/core/internal/repository"
	"github.com/probe-system/core/internal/service"
	"github.com/probe-system/core/internal/ws"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to config file")
	flag.Parse()

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("Probe System Core starting...")

	// Initialize database
	db, err := repository.NewDB(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	agentRepo := repository.NewAgentRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	metricsRepo := repository.NewMetricsRepository(db)
	trafficRepo := repository.NewTrafficRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	scriptRepo := repository.NewScriptRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	geoSvc := service.NewGeoService()
	agentSvc := service.NewAgentService(agentRepo, geoSvc)
	groupSvc := service.NewGroupService(groupRepo)
	metricSvc := service.NewMetricService(metricsRepo)
	trafficSvc := service.NewTrafficService(trafficRepo)
	taskSvc := service.NewTaskService(taskRepo)
	scriptSvc := service.NewScriptService(scriptRepo)
	alertSvc := service.NewAlertService(alertRepo)
	settingsSvc := service.NewSettingsService(settingsRepo)
	authSvc := service.NewAuthService(userRepo, cfg.Auth.JWTSecret)

	// Setup notifiers
	settings, _ := settingsSvc.Get(context.Background())
	if settings != nil {
		if settings.TelegramBotToken != "" {
			alertSvc.AddNotifier(notify.NewTelegramNotifier(settings.TelegramBotToken, settings.TelegramChatID))
		}
		if settings.SMTPHost != "" {
			alertSvc.AddNotifier(notify.NewEmailNotifier(
				settings.SMTPHost, settings.SMTPPort,
				settings.SMTPUsername, settings.SMTPPassword,
				settings.SMTPFrom, settings.AlertEmailTo,
			))
		}
	}

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	wsHandler := ws.NewHandler(hub, cfg.Agent.Token)
	wsHandler.SetServices(agentSvc, metricSvc, trafficSvc, taskSvc, alertSvc)

	// Initialize HTTP handlers
	adminHandler := handler.NewAdminHandler(
		agentSvc, groupSvc, metricSvc, trafficSvc,
		taskSvc, scriptSvc, alertSvc, settingsSvc, authSvc, wsHandler,
	)
	publicHandler := handler.NewPublicHandler(agentSvc, metricSvc, trafficSvc)
	dashboardWSHandler := handler.NewDashboardWSHandler(agentSvc, metricSvc, trafficSvc)
	scriptHandler := handler.NewScriptHandler(scriptSvc)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(handler.CORSMiddleware())

	// WebSocket endpoints
	r.GET("/ws/agent", func(c *gin.Context) {
		wsHandler.ServeWS(c.Writer, c.Request)
	})
	r.GET("/ws/dashboard", dashboardWSHandler.ServeWS)

	// Public API
	public := r.Group("/api/public")
	{
		public.GET("/agents", publicHandler.ListAgents)
		public.GET("/agents/:id", publicHandler.GetAgent)
	}

	// Auth API
	r.POST("/api/auth/login", adminHandler.Login)

	// Script download API (for agents)
	r.GET("/api/scripts/:id/content", scriptHandler.GetScriptContent)

	// Admin API (protected)
	admin := r.Group("/api/admin")
	admin.Use(handler.AuthMiddleware(authSvc))
	{
		// Agents
		admin.GET("/agents", adminHandler.ListAgents)
		admin.GET("/agents/:id", adminHandler.GetAgent)
		admin.PATCH("/agents/:id/remark", adminHandler.UpdateAgentRemark)
		admin.PATCH("/agents/:id/group", adminHandler.AssignAgentGroup)
		admin.PATCH("/agents/:id/visibility", adminHandler.SetAgentVisibility)
		admin.DELETE("/agents/:id", adminHandler.DeleteAgent)
		admin.GET("/agents/:id/metrics", adminHandler.GetAgentMetrics)
		admin.GET("/agents/:id/metrics/history", adminHandler.GetAgentMetricsHistory)
		admin.GET("/agents/:id/traffic", adminHandler.GetAgentTraffic)
		admin.POST("/agents/:id/traffic/cycle", adminHandler.ConfigureTrafficCycle)

		// Groups
		admin.GET("/groups", adminHandler.ListGroups)
		admin.POST("/groups", adminHandler.CreateGroup)
		admin.PUT("/groups/:id", adminHandler.UpdateGroup)
		admin.DELETE("/groups/:id", adminHandler.DeleteGroup)

		// Tasks
		admin.GET("/tasks", adminHandler.ListTasks)
		admin.POST("/tasks", adminHandler.CreateTask)
		admin.GET("/tasks/:id/results", adminHandler.GetTaskResults)
		admin.POST("/tasks/:id/cancel", adminHandler.CancelTask)

		// Scripts
		admin.GET("/scripts", adminHandler.ListScripts)
		admin.POST("/scripts", adminHandler.CreateScript)
		admin.GET("/scripts/:id", adminHandler.GetScript)
		admin.DELETE("/scripts/:id", adminHandler.DeleteScript)

		// Alerts
		admin.GET("/alerts/rules", adminHandler.ListAlertRules)
		admin.POST("/alerts/rules", adminHandler.CreateAlertRule)
		admin.PUT("/alerts/rules/:id", adminHandler.UpdateAlertRule)
		admin.DELETE("/alerts/rules/:id", adminHandler.DeleteAlertRule)
		admin.GET("/alerts/active", adminHandler.GetActiveAlerts)
		admin.GET("/alerts/history", adminHandler.GetAlertHistory)

		// Settings
		admin.GET("/settings", adminHandler.GetSettings)
		admin.PUT("/settings", adminHandler.UpdateSettings)
		admin.POST("/settings/password", adminHandler.ChangePassword)
	}

	// Serve static files (frontend)
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})
	r.Static("/assets", "./web/dist/assets")

	// Start background tasks
	go runCleanupTask(metricSvc, settingsSvc)
	go runTrafficCycleCheck(trafficSvc)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}

func runCleanupTask(metricSvc service.MetricService, settingsSvc service.SettingsService) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		settings, err := settingsSvc.Get(ctx)
		if err != nil {
			continue
		}

		deleted, err := metricSvc.Cleanup(ctx, settings.DataRetentionDays)
		if err != nil {
			log.Printf("Cleanup error: %v", err)
		} else if deleted > 0 {
			log.Printf("Cleaned up %d old metric records", deleted)
		}
	}
}

func runTrafficCycleCheck(trafficSvc service.TrafficService) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		if err := trafficSvc.CheckAndResetCycles(ctx); err != nil {
			log.Printf("Traffic cycle check error: %v", err)
		}
	}
}
