package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type HealthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

type MetricsResponse struct {
	CpuUtilizationPercent    float64 `json:"cpu_utilization_percent"`
	MemoryUtilizationPercent float64 `json:"memory_utilization_percent"`
	Status                   string  `json:"status"`
	Error                    string  `json:"error,omitempty"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Service: "metrics-api",
		Status:  "healthy",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cpuPercent, err := cpu.Percent(100*time.Millisecond, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MetricsResponse{Error: err.Error()})
		return
	}

	vMem, err := mem.VirtualMemory()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(MetricsResponse{Error: err.Error()})
		return
	}

	response := MetricsResponse{
		CpuUtilizationPercent:    cpuPercent[0],
		MemoryUtilizationPercent: vMem.UsedPercent,
		Status:                   "active",
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "5100"
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/metrics", metricsHandler)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
