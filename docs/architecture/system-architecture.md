# System Architecture

This document provides details on how components interact to create a distributed, auto-scaling environment.

## High-Level Architecture

The system consists of three main layers:

1. **Host Layer** - Manages the virtualization infrastructure
2. **Control Plane** - Orchestrates load balancing and scaling
3. **Worker Plane** - Runs the distributed application

```mermaid
graph TB
    subgraph Host["Host Machine"]
        VBox[VirtualBox]
        ServerMgr[Server Manager API<br/>Port 3000]
    end

    subgraph Control["Control Node VM"]
        LB[Nginx Load Balancer<br/>Port 80]
        Scaler[Scaler Service]
    end

    subgraph Agents["Agent Node VMs"]
        Agent1["Agent Node 1<br/>• Metrics API (5100)<br/>• App (5000)"]
        Agent2["Agent Node 2<br/>• Metrics API (5100)<br/>• App (5000)"]
        AgentN["Agent Node N<br/>• Metrics API (5100)<br/>• App (5000)"]
    end

    ServerMgr -->|VM Control| VBox
    VBox -->|Creates/Manages| Agents
    VBox -->|Manages| Control

    Scaler -->|Scale Commands| ServerMgr
    Scaler -->|Updates Config| LB

    Scaler -->|Collects Metrics| Agent1
    Scaler -->|Collects Metrics| Agent2
    Scaler -->|Collects Metrics| AgentN

    LB -->|Distributes Traffic| Agent1
    LB -->|Distributes Traffic| Agent2
    LB -->|Distributes Traffic| AgentN

    User[User/Client] -->|HTTP Requests| LB
```


## System Topology

### Network Architecture

All VMs are connected through VirtualBox's network configuration:

- **Host Machine** - Runs VirtualBox and Server Manager API
- **Control Node** - Static VM, always running
  - Port 80 (forwarded to host 8080) - Load Balancer
  - Port 22 (forwarded to host 2223) - SSH
- **Agent Nodes** - Dynamic VMs, scaled on demand
  - Port 5000 (forwarded to host 5001, 5002, ...) - Application API
  - Port 5100 (forwarded to host 5101, 5102, ...) - Metrics API
  - Port 22 (forwarded to host 2224, 2225, ...) - SSH

### Component Communication

```mermaid
sequenceDiagram
    participant User
    participant LB as Load Balancer
    participant Scaler
    participant Metrics as Metrics API
    participant App as Application
    participant ServerMgr as Server Manager

    Note over Scaler: Every 10 seconds
    Scaler->>Metrics: GET /metrics
    Metrics-->>Scaler: CPU, Memory stats

    Scaler->>Scaler: Evaluate scaling rules

    alt High Load Detected
        Scaler->>ServerMgr: POST /api/v1/servers/power<br/>{action: "on", server: "agent-N"}
        ServerMgr-->>Scaler: VM started
        Scaler->>LB: Update upstream config
        LB-->>Scaler: Config reloaded
    end

    User->>LB: HTTP Request
    LB->>App: Forward to least-loaded agent
    App-->>LB: Response
    LB-->>User: Response
```

## Request Flow

### Normal Request Handling

1. **User sends HTTP request** to the load balancer (port 8080 on host, forwarded to port 80 on control node)
2. **Nginx Load Balancer** receives the request and selects an agent node using least-connections algorithm
3. **Agent Node** processes the request through the distributed application
4. **Response flows back** through the load balancer to the user

### Auto-Scaling Flow

#### Scale-Up Scenario

```mermaid
sequenceDiagram
    participant Scaler
    participant Metrics as Agent Metrics
    participant ServerMgr as Server Manager
    participant VBox as VirtualBox
    participant LB as Load Balancer
    participant NewAgent as New Agent Node

    loop Every 10s
        Scaler->>Metrics: Collect metrics from all agents
        Metrics-->>Scaler: CPU: 85%, Memory: 78%
    end

    Scaler->>Scaler: CPU > 70% threshold<br/>Decision: Scale Up

    Scaler->>ServerMgr: POST /api/v1/servers/power<br/>{action: "on", server: "agent-3"}
    ServerMgr->>VBox: VBoxManage startvm agent-3
    VBox-->>ServerMgr: VM started
    ServerMgr-->>Scaler: Success

    Scaler->>NewAgent: Wait for health check
    NewAgent-->>Scaler: Metrics API ready

    Scaler->>LB: Update nginx upstream config
    LB->>LB: Reload configuration
    LB-->>Scaler: Ready to route traffic

    Note over Scaler,NewAgent: New agent now receiving traffic
```

#### Scale-Down Scenario

```mermaid
sequenceDiagram
    participant Scaler
    participant Metrics as Agent Metrics
    participant LB as Load Balancer
    participant Agent as Agent to Remove
    participant ServerMgr as Server Manager
    participant VBox as VirtualBox

    loop Every 10s
        Scaler->>Metrics: Collect metrics from all agents
        Metrics-->>Scaler: CPU: 15%, Memory: 25%
    end

    Scaler->>Scaler: CPU < 30% threshold<br/>Above minimum instances<br/>Decision: Scale Down

    Scaler->>LB: Remove agent from upstream config
    LB->>LB: Reload configuration
    LB-->>Scaler: No longer routing to agent

    Scaler->>Scaler: Wait 30s for connection drain

    Scaler->>ServerMgr: POST /api/v1/servers/power<br/>{action: "off", server: "agent-3"}
    ServerMgr->>VBox: VBoxManage controlvm agent-3 poweroff
    VBox-->>ServerMgr: VM stopped
    ServerMgr-->>Scaler: Success

    Note over Scaler,Agent: Agent removed from cluster
```

## Data Flow

### Metrics Collection

```
┌─────────────┐
│   Scaler    │
│ (Control)   │
└──────┬──────┘
       │
       │ HTTP GET /metrics (every 30s)
       │
       ├──────────────┬──────────────┬──────────────┐
       │              │              │              │
       ▼              ▼              ▼              ▼
┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐
│ Metrics API│ │ Metrics API│ │ Metrics API│ │ Metrics API│
│  Agent 1   │ │  Agent 2   │ │  Agent 3   │ │  Agent N   │
└────────────┘ └────────────┘ └────────────┘ └────────────┘
       │              │              │              │
       │              │              │              │
       └──────────────┴──────────────┴──────────────┘
                      │
                      ▼
            {
              "cpu": 45.2,
              "memory": {
                "total": 8589934592,
                "available": 4294967296,
                "percent": 50.0
              },
              "status": "active"
            }
```

### Application Traffic

```
┌──────────┐
│   User   │
└────┬─────┘
     │
     │ HTTP Request
     ▼
┌─────────────────┐
│ Nginx Load      │
│ Balancer        │
│ (Control Node)  │
└────┬────────────┘
     │
     │ Least Connections Algorithm
     │
     ├──────────────┬──────────────┬──────────────┐
     │              │              │              │
     ▼              ▼              ▼              ▼
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│ App:5000 │  │ App:5000 │  │ App:5000 │  │ App:5000 │
│ Agent 1  │  │ Agent 2  │  │ Agent 3  │  │ Agent N  │
└──────────┘  └──────────┘  └──────────┘  └──────────┘
     │              │              │              │
     └──────────────┴──────────────┴──────────────┘
                    │
                    ▼
          External Services
          (PostgreSQL, Redis)
```

## Component Responsibilities

### Host Layer

**VirtualBox**
- Provides virtualization platform
- Manages VM lifecycle

**Server Manager API**
- Exposes API for VM control
- Executes VBoxManage commands
- Provides VM status information

### Control Plane

**Nginx Load Balancer**
- Receives all incoming HTTP traffic
- Distributes requests using least-connections algo
- Dynamically reloads configuration without downtime

**Scaler Service**
- Monitors agent metrics every 10 seconds
- Evaluates scaling rules based on CPU/memory thresholds
- Triggers scale-up/down actions via Server Manager API
- Updates load balancer configuration dynamically
- Maintains minimum and maximum instance counts

### Worker Plane

**Metrics API**
- Exposes system metrics (CPU, memory)

**Distributed Application**
- Runs in containers on each agent

## Deployment Architecture

### VM Configuration

| VM Type | Count | Always Running | Purpose |
|---------|-------|----------------|---------|
| Control Node | 1 | Yes | Load balancing and orchestration |
| Agent Node | 2-N | Dynamic | Application workload |

### Port Mapping

**Control Node:**
```
Host:8080 -> VM:80   (Load Balancer)
Host:2223 -> VM:22   (SSH)
```

**Agent Nodes:**
```
Host:5001 -> Agent1:5000  (Application)
Host:5101 -> Agent1:5100  (Metrics)
Host:2224 -> Agent1:22    (SSH)

Host:5002 -> Agent2:5000  (Application)
Host:5102 -> Agent2:5100  (Metrics)
Host:2225 -> Agent2:22    (SSH)

... and so on for additional agents
```

**Visual Representation:**
```mermaid
flowchart TB
    U[User]
    CN[Control Node<br/>LB :80]

    subgraph AgentN["Agent N"]
        AN_APP[App :5000]
        AN_MET[Metrics :5100]
    end

    subgraph Agent2["Agent 2"]
        A2_APP[App :5000]
        A2_MET[Metrics :5100]
    end

    subgraph Agent1["Agent 1"]
        A1_APP[App :5000]
        A1_MET[Metrics :5100]
    end

    U -->|Host:8080 to VM:80| CN

    CN -->|Host:5001 to :5000| A1_APP
    CN -->|Host:5002 to :5000| A2_APP
    CN -->|Host:50NN to :5000| AN_APP

    CN -.->|Host:5101 to :5100| A1_MET
    CN -.->|Host:5102 to :5100| A2_MET
    CN -.->|Host:51NN to :5100| AN_MET
```

## Security Considerations

### SSH Key Authentication
- Password authentication disabled
- All inter-node communication uses SSH keys
- Control node has SSH access to all agent nodes

### Network Isolation
- VMs communicate through VirtualBox internal network
- Only necessary ports exposed to host machine
- Firewall rules configured on each VM

### Service Security
- All services run as non-root users
- Docker containers use non-root user
- Minimal attack surface with Alpine-based images

## Scalability Characteristics

### Horizontal Scaling
- Agent nodes can be added/removed dynamically
- No theoretical limit on agent count

## Next Steps

- [Components](./components.md) - Learn about the components and their roles in the architecture.
