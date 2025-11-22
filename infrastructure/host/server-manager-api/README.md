# Server Manager API (Go)

A Go-based REST API for managing VirtualBox virtual machines. This API allows you to power on/off VirtualBox VMs through HTTP requests.

## Features

- Power on/off VirtualBox virtual machines
- RESTful API endpoints
- Configurable server list via environment variables
- JSON response format
- Error handling and validation
- Lightweight and fast (written in Go)

## Prerequisites

- Go 1.25.4 or higher
- VirtualBox installed and `VBoxManage` command available in PATH
- Virtual machines configured in VirtualBox

## Setup Instructions

### 1. Navigate to Project Directory

```bash
cd infrastructure/host/server-manager-api
```

### 2. Environment Configuration

Create a `.env` file to configure your servers:

```bash
# Example .env content
SERVERS=YourVM1,YourVM2
PORT=3000
```

**Note:** Replace the server names with your actual VirtualBox VM names.

### 3. Build and Run

```bash
# Run directly
go run main.go

# Or build and run
go build -o server-manager-api
./server-manager-api
```

The API will be available at `http://<host_ip>:3000`

## API Endpoints

### Health Check

```http
GET /
```

**Response:**
```json
{
  "service": "server-manager-api",
  "status": "running"
}
```

### Power Control

```http
POST /api/v1/servers/power
Content-Type: application/json

{
  "action": "on|off",
  "server": "ServerName"
}
```

**Parameters:**
- `action`: Either "on" or "off"
- `server`: Name of the server (must match VM name in VirtualBox)

**Success Response (200):**
```json
{
  "status": "Server 'gandalf' turned on successfully."
}
```

**Error Responses:**
- `400`: Invalid action or missing server field
- `404`: Unknown server name
- `500`: VirtualBox command failed

## Troubleshooting

1. **VBoxManage not found**
   - Ensure VirtualBox is installed and in your PATH.

2. **VM not found**
   - Check VM names with `VBoxManage list vms`.
   - Ensure VM names in `.env` match exactly.

3. **Permission denied**
   - Ensure the user running the API has permission to control VirtualBox VMs.
