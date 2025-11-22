package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"scaler/pkg/config"
)

type MetricsResponse struct {
	CpuUtilizationPercent    float64 `json:"cpu_utilization_percent"`
	MemoryUtilizationPercent float64 `json:"memory_utilization_percent"`
	Status                   string  `json:"status"`
	Error                    string  `json:"error,omitempty"`
}

func IsActive(agent config.AgentConfig) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return false
	}

	commandArgs := []string{
		"-p", agent.SSH.Port,
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-i", fmt.Sprintf("%s/.ssh/id_ed25519", homeDir),
		fmt.Sprintf("%s@%s", agent.SSH.User, agent.SSH.IP),
		"exit 0",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ssh", commandArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if os.Getenv("DEBUG") == "true" {
			fmt.Printf("Error checking agent status: %v. Output: %s\n", err, string(output))
		}
		return false
	}

	return true
}

func ManagePower(serverManagerAPI, serverName, action string) error {
	payload := map[string]string{
		"action": action,
		"server": serverName,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverManagerAPI+"/api/v1/servers/power", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server manager returned status: %d", resp.StatusCode)
	}
	return nil
}

func GetMetrics(agent config.AgentConfig) (float64, float64, error) {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get(agent.TelemetryURL)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("metrics api returned status: %d", resp.StatusCode)
	}

	var metrics MetricsResponse
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return 0, 0, err
	}

	if metrics.Error != "" {
		return 0, 0, fmt.Errorf("metrics api error: %s", metrics.Error)
	}

	return metrics.CpuUtilizationPercent, metrics.MemoryUtilizationPercent, nil
}

func UpdateUpstreamConfig(activeAgents []config.AgentConfig) error {
	var upstreamServers []string
	for _, agent := range activeAgents {
		url := agent.UpstreamURL
		url = strings.TrimPrefix(url, "http://")
		url = strings.TrimPrefix(url, "https://")
		upstreamServers = append(upstreamServers, fmt.Sprintf("    server %s;", url))
	}

	upstreamBlock := fmt.Sprintf("upstream backend {\n%s\n}", strings.Join(upstreamServers, "\n"))
	cmd := exec.Command("sudo", "tee", "/etc/nginx/conf.d/upstream.conf")
	cmd.Stdin = strings.NewReader(upstreamBlock)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to write upstream config: %v", err)
	}

	fmt.Println("Upstream config:")
	fmt.Println(upstreamBlock)

	if err := exec.Command("sudo", "systemctl", "reload", "nginx").Run(); err != nil {
		return fmt.Errorf("failed to reload nginx: %v", err)
	}

	return nil
}
