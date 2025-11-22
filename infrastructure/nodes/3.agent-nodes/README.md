# Agent Nodes

The worker nodes that host the actual applications and provide system metrics for scaling decisions.

## Contents

- **[Agent Node Setup Guide](agent-node-setup.md)**: Steps for configuring firewalls, port forwarding, and the Metrics API.
- **`metrics-api/`**: Go-based service that exposes real-time system metrics (CPU, Memory).
- **`setup-metrics-api.sh`**: Script to install and start the Metrics API as a systemd service.

## Responsibilities

- **Workload Execution**: Hosting the stateless applications managed by the cluster.
- **Observability**: Providing the telemetry data used by the Control Node to make scaling decisions.
