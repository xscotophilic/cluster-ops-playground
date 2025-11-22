# Agent Metrics API (Go Version)

A lightweight Go-based API that runs on **Agent Nodes**. It exposes real-time system metrics (CPU, Memory) so the **Control Node** can make scaling decisions.

## Features

- Exposes CPU usage percentage.
- Exposes Memory usage statistics.
- Health check endpoint.
- JSON response format.
- High performance and low footprint (written in Go).

## Prerequisites

- Go 1.25.4 or higher

## Setup Instructions

### 1. Navigate to Directory

```bash
cd infrastructure/nodes/3.agent-nodes/metrics-api
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Environment Configuration

Copy the example environment file:

```bash
cp .env.example .env
```

The default port is `5100`. You can change it in `.env` if needed.

## Running the Application

```bash
# Run directly
go run main.go

# Or build and run
go build -o metrics-api
./metrics-api
```

The API will be available at `http://0.0.0.0:5100`.

## API Endpoints

### Health Check

```http
GET /health
```

**Response:**
```json
{
  "service": "metrics-api",
  "status": "healthy"
}
```

### Get Metrics

```http
GET /metrics
```

**Response:**
```json
{
  "cpu": 15.2,
  "memory": {
    "total": 8589934592,
    "available": 4294967296,
    "percent": 50.0
  },
  "status": "active"
}
```
