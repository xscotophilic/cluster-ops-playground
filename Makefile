.PHONY: build-all clean

# Paths
SERVER_MANAGER_DIR := infrastructure/host/server-manager-api
SCALER_DIR := infrastructure/nodes/2.control-node/scaler
METRICS_API_DIR := infrastructure/nodes/3.agent-nodes/metrics-api

# Build targets
build-all: build-server-manager build-scaler build-metrics
	@echo "All services built successfully."

build-server-manager:
	@echo "Building Server Manager API..."
	cd $(SERVER_MANAGER_DIR) && go build -o server-manager-api .

build-scaler:
	@echo "Building Scaler Service..."
	cd $(SCALER_DIR) && go build -o scaler .

build-metrics:
	@echo "Building Metrics API..."
	cd $(METRICS_API_DIR) && go build -o metrics-api .

# Clean target
clean:
	@echo "Cleaning up binaries..."
	rm -f $(SERVER_MANAGER_DIR)/server-manager-api
	rm -f $(SCALER_DIR)/scaler
	rm -f $(METRICS_API_DIR)/metrics-api
	@echo "Clean complete."
