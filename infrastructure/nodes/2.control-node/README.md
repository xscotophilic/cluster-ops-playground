# Control Node

The brain of the cluster, responsible for traffic distribution and automated scaling.

## Contents

- **[Control Node Setup Guide](control-node-setup.md)**: Instructions for configuring SSH access to agents, deploying the load balancer, and setting up the scaler.
- **`scaler/`**: Go-based service that monitors agent health and manages the load balancer configuration.
- **`setup-load-balancer.sh`**: Script to deploy and configure the Nginx load balancer.
- **`setup-scaler.sh`**: Script to install and start the Scaler as a systemd service.

## Responsibilities

- **Traffic Management**: The Load Balancer (Nginx) distributes incoming requests to available Agent Nodes.
- **Orchestration**: The Scaler monitors metrics and automatically adjusts the pool of active agents.
