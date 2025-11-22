package config

import (
	"encoding/json"
	"log"
	"os"
)

type SSHConfig struct {
	Port string `json:"port"`
	User string `json:"user"`
	IP   string `json:"ip"`
}

type AgentConfig struct {
	ServerName   string    `json:"server_name"`
	UpstreamURL  string    `json:"upstream_url"`
	TelemetryURL string    `json:"telemetry_url"`
	SSH          SSHConfig `json:"ssh"`
}

type ScalerConfig struct {
	ServerManagerAPI string
	AvailableAgents  []AgentConfig
}

func LoadConfig() ScalerConfig {
	config := ScalerConfig{
		ServerManagerAPI: os.Getenv("SERVER_MANAGER_API"),
	}

	agentsJSON := os.Getenv("AGENTS")
	if agentsJSON != "" {
		err := json.Unmarshal([]byte(agentsJSON), &config.AvailableAgents)
		if err != nil {
			log.Printf("Error parsing AGENTS environment variable: %v\n", err)
		}
	}

	return config
}
