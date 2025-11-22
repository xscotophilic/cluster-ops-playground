# Project Overview

## What is Cluster Operations Playground?

A hands-on learning environment that simulates a moderate level production-grade distributed system.

## Learning Objectives

This project helps you understand:

- **Distributed System Architecture** - How multiple servers work together to handle requests
- **Load Balancing** - Distributing traffic across multiple backend servers
- **Auto-Scaling** - Automatically adjusting capacity based on demand
- **Infrastructure as Code** - Managing infrastructure through scripts and configuration
- **Container Orchestration** - Deploying and managing containerized applications
- **Metrics and Monitoring** - Collecting and using system metrics for scaling decisions

## Use Cases

### Learning and Understanding
- Understand distributed systems concepts
- Practice infrastructure management

### Development and Testing
- Test applications in a distributed environment
- Validate load balancing configurations
- Develop infrastructure automation

### Demonstrations
- To Demonstrate DevOps knowledge
- Build a foundation for cloud migration

## Architecture at a Glance

```
┌─────────────────────────────────────────────────────────┐
│                    Host Machine                         │
│  ┌────────────────────────────────────────────────────┐ │
│  │  VirtualBox + Server Manager API                   │ │
│  └────────────────────────────────────────────────────┘ │
└────────────────────────────┬────────────────────────────┘
                             │
       ┌─────────────────────┴───────────────────┐
       │                                         │
┌──────▼──────┐                        ┌─────────▼────────┐
│ Control Node│                        │   Agent Nodes    │
│             │                        │   (Dynamic)      │
│ • Nginx LB  │───────────────────────▶│                  │
│ • Scaler    │                        │ • Metrics API    │
│             │                        │ • Distributed API│
└─────────────┘                        └──────────────────┘
```

## Project Structure

```
cluster-ops-playground/
├── distributed-pluggable-api/
│   ├── docker/
│   └── compose/
├── infrastructure/
│   ├── host/
│   │   └── server-manager-api/
│   ├── nodes/
│   │   ├── 1.base-setup/
│   │   ├── 2.control-node/
│   │   └── 3.agent-nodes/
└── docs/
    ├── architecture/
    └── distributed-api/
```

## Relationship to async-fibonacci-lab

This project uses [async-fibonacci-lab](https://github.com/xscotophilic/async-fibonacci-lab) as the example distributed application. The async-fibonacci-lab is a learning project that demonstrates:

- Asynchronous job processing
- API and worker separation
- Database migrations
- Redis-based job queues

**How it's integrated:**
- The Dockerfiles in `distributed-pluggable-api/` clone async-fibonacci-lab at build time
- This keeps the infrastructure code separate from the application code
- You can easily replace it with your own application
- **Note on Environment Configuration**: The scaler's deployment scripts include setup for environment variables to automate the replication of the pluggable-api. While this simplifies the default setup, you may need to adjust the `deploy.go` logic when integrating a custom application to ensure the correct variables are transferred to agent nodes.

## Next Steps

- [System Architecture](./architecture/README.md) - Understand how components interact
