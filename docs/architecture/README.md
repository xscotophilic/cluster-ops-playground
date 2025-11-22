# Architecture Overview

This section provides complete documentation of the Cluster Operations Playground architecture.

## Sections

### [System Architecture](./system-architecture.md)

Includes:
- System topology and network design
- Request handling lifecycle
- Communication flows
- Data flow diagrams

### [Components](./components.md)

Includes:
- Description of Host System (VirtualBox + Server Manager API)
- Description of Control Node (Load Balancer + Scaler)
- Description of Agent Nodes (Metrics API + Distributed API)

### [Scaling Concepts](./scaling-concepts.md)

Includes:
- Metrics-based scaling decisions
- Design pros and cons

## Architecture Principles

### Separation of Concerns
- **Infrastructure Layer** - VM management and cluster harmony
- **Control Plane** - Load balancing and scaling decisions
- **Application Layer** - Business logic and data processing

### Stateless Design
- Agent nodes are stateless and can be added/removed dynamically
- All persistent state is external (databases, caches)
- Horizontal scaling without coordination

### Metrics-Driven Operations
- Automated scaling based on near real-time CPU and Memory thresholds

### Developer Experience
- Simple setup with clear documentation
- Pluggable application architecture

## Next Steps

- [In-depth System Architecture](./system-architecture.md) - Understand how components interact
