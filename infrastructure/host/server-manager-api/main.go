package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Servers []string
	Port    string
}

var ErrVMAlreadyRunning = errors.New("vm already running")

type VBoxManager struct{}

type Virtualizer interface {
	StartVM(name string) error
	StopVM(name string) error
}

type PowerRequest struct {
	Action string `json:"action"`
	Server string `json:"server"`
}

type Response struct {
	Status  string `json:"status,omitempty"`
	Error   string `json:"error,omitempty"`
	Service string `json:"service,omitempty"`
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	serversEnv := os.Getenv("SERVERS")
	var servers []string
	if serversEnv != "" {
		parts := strings.Split(serversEnv, ",")
		for _, s := range parts {
			servers = append(servers, strings.TrimSpace(s))
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	return &Config{
		Servers: servers,
		Port:    port,
	}
}

func (v *VBoxManager) StartVM(name string) error {
	cmd := exec.Command("VBoxManage", "startvm", name, "--type", "headless")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "already locked by a session") {
			return ErrVMAlreadyRunning
		}
		return fmt.Errorf("failed to start vm: %v, output: %s", err, string(output))
	}
	return nil
}

func (v *VBoxManager) StopVM(name string) error {
	cmd := exec.Command("VBoxManage", "controlvm", name, "poweroff")
	return cmd.Run()
}

func main() {
	config := LoadConfig()
	virtualizer := &VBoxManager{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		jsonResponse(w, http.StatusOK, Response{
			Service: "server-manager-api",
			Status:  "running",
		})
	})

	http.HandleFunc("/api/v1/servers/power", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req PowerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonResponse(w, http.StatusBadRequest, Response{Error: "Invalid JSON"})
			return
		}

		if req.Action != "on" && req.Action != "off" {
			jsonResponse(w, http.StatusBadRequest, Response{Error: "Invalid action. Use 'on' or 'off'."})
			return
		}

		if req.Server == "" {
			jsonResponse(w, http.StatusBadRequest, Response{Error: "Missing 'server' field."})
			return
		}

		allowed := false
		for _, s := range config.Servers {
			if s == req.Server {
				allowed = true
				break
			}
		}

		if !allowed {
			jsonResponse(w, http.StatusNotFound, Response{Error: fmt.Sprintf("Unknown server '%s'.", req.Server)})
			return
		}

		var err error
		if req.Action == "on" {
			err = virtualizer.StartVM(req.Server)
		} else {
			err = virtualizer.StopVM(req.Server)
		}

		if err != nil {
			if err == ErrVMAlreadyRunning {
				jsonResponse(w, http.StatusOK, Response{Status: fmt.Sprintf("Server '%s' was already on.", req.Server)})
				return
			}
			errMsg := fmt.Sprintf("Failed to perform %s on '%s': %v", req.Action, req.Server, err)
			jsonResponse(w, http.StatusInternalServerError, Response{Error: errMsg})
			return
		}

		jsonResponse(w, http.StatusOK, Response{Status: fmt.Sprintf("Server '%s' turned %s successfully.", req.Server, req.Action)})
	})

	log.Printf("Server starting on port %s...", config.Port)
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatal(err)
	}
}

func jsonResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
