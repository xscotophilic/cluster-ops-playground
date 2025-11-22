# Cluster Operations Playground

This project is a hands-on learning environment for understanding distributed systems, load balancing, and auto-scaling using virtualized infrastructure.

## Overview

This project simulates a moderate level production-grade distributed cluster using VirtualBox VMs. Learn about:

- **Distributed System Architecture** - Multiple servers working together
- **Infrastructure as Code** - Automated setup and management
- **Load Balancing** - Traffic distribution with Nginx
- **Auto-Scaling** - Metrics-based capacity adjustment

## Features

- ⚖️ **Load Balancer** - Distributes traffic across agent nodes
- 📈 **Auto-Scaling** - Dynamically adjusts capacity based on CPU/memory metrics
- 🔧 **Pluggable Architecture** - Easy to swap applications

## Technology Stack

- **VirtualBox** - VM virtualization
- **Ubuntu Server** - Operating system
- **Nginx** - Load balancer
- **Go** - Infrastructure services
- **Docker** - Container runtime
- **[Will change depending on the application you wanna run in the cluster]**: **Node.js** - Pluggable application (async-fibonacci-lab)

## Getting Started

At the end of each document, there is a "Next Steps" section that will guide you to the next document.

### Quick Start

1. **Read the Documentation**

   Start with project [documentation](./docs/README.md), to get a better understanding of the project.

2. **Set Up Infrastructure (Running your own application in the cluster)**

   Dive into [infrastructure setup](./infrastructure/README.md), follow the instructions to set up the infrastructure (Make sure you replace the pluggable application with your own and modify stuff as per your requirements).

3. **Deploy and Test Application**

   Run services and verify the intended behavior.

## Contributing

This is a learning project. Feel free to:

- Experiment with different configurations
- Try deploying your own applications
- Improve the documentation
- Share your learnings

## Next Steps

- [Getting Started](./docs/README.md) - Outlines the learning path for this project.
