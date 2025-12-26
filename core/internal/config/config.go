package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server    ServerConfig    `json:"server"`
	Database  DatabaseConfig  `json:"database"`
	Auth      AuthConfig      `json:"auth"`
	Agent     AgentConfig     `json:"agent"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type DatabaseConfig struct {
	Path string `json:"path"`
}

type AuthConfig struct {
	JWTSecret string `json:"jwt_secret"`
}

type AgentConfig struct {
	Token string `json:"token"`
}

func Load(path string) (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Path: "probe.db",
		},
		Auth: AuthConfig{
			JWTSecret: "change-me-in-production",
		},
		Agent: AgentConfig{
			Token: "",
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	// Override with environment variables
	if v := os.Getenv("PROBE_HOST"); v != "" {
		config.Server.Host = v
	}
	if v := os.Getenv("PROBE_PORT"); v != "" {
		var port int
		if err := json.Unmarshal([]byte(v), &port); err == nil {
			config.Server.Port = port
		}
	}
	if v := os.Getenv("PROBE_DB_PATH"); v != "" {
		config.Database.Path = v
	}
	if v := os.Getenv("PROBE_JWT_SECRET"); v != "" {
		config.Auth.JWTSecret = v
	}
	if v := os.Getenv("PROBE_AGENT_TOKEN"); v != "" {
		config.Agent.Token = v
	}

	return config, nil
}
