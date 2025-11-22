# Docker Setup

This document provides a comprehensive guide to the Docker configuration for deploying stateless applications in a distributed environment. It covers Dockerfiles, multi-stage builds, container optimization, and best practices.

## Overview

The distributed setup uses **multi-stage Docker builds** to create small, secure, production-ready images. This approach is applicable to any stateless application you want to deploy in the cluster.

> [!NOTE]
> The example application (async-fibonacci-lab) is used for demonstration purposes. You should replace it with your own stateless application following the same Docker patterns described here.

## Dockerfile Architecture

### Multi-Stage Build Strategy

```
Stage 1: Clone/Source  → Fetch application code
Stage 2: Build         → Install dependencies and prepare application
Stage 3: Runtime       → Minimal production image
```

**Benefits:**
- **Smaller images** - Only runtime dependencies included
- **Faster builds** - Cached layers reused
- **Security** - No build tools in production image
- **Separation of concerns** - Build and runtime environments isolated

## Application Components

A typical distributed stateless application consists of:

1. **Server/API** - Handles HTTP requests, serves API endpoints
2. **Worker** - Processes background jobs from a queue
3. **Database** - Persistent data storage (PostgreSQL, MySQL, etc.)
4. **Cache/Queue** - In-memory data store (Redis, RabbitMQ, etc.)

Each component runs in its own container with specific configurations.

## API Dockerfile Pattern

**Location:** `distributed-pluggable-api/docker/server/Dockerfile`

### Recommended Structure

```dockerfile
# syntax = docker/dockerfile:1

# Stage 1: Clone/Source
FROM alpine/git AS clone
WORKDIR /repo
RUN git clone <your-repository-url> .

# Stage 2: Runtime
FROM node:20-alpine AS base

ENV NODE_ENV=production
WORKDIR /app

# Security: Run as non-root user
RUN addgroup -S app && adduser -S app -G app \
  && chown -R app:app /app
USER app

# Install dependencies
ARG SERVER_PATH=/repo/server
COPY --from=clone ${SERVER_PATH}/package*.json ./
RUN npm ci --omit=dev && npm cache clean --force

# Copy application code
COPY --from=clone ${SERVER_PATH}/src ./src
COPY --from=clone ${SERVER_PATH}/migrations ./migrations

# Expose application port
EXPOSE 5000

# Start command
CMD ["npm", "run", "start"]
```

## Worker Dockerfile Pattern

**Location:** `distributed-pluggable-api/docker/worker/Dockerfile`

### Recommended Structure

```dockerfile
# syntax = docker/dockerfile:1

FROM alpine/git AS clone
WORKDIR /repo
RUN git clone <your-repository-url> .

FROM node:20-alpine AS base

ENV NODE_ENV=production
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app \
  && chown -R app:app /app
USER app

ARG WORKER_PATH=/repo/worker

COPY --from=clone ${WORKER_PATH}/package*.json ./
RUN npm ci --omit=dev && npm cache clean --force

COPY --from=clone ${WORKER_PATH}/src ./src

CMD ["npm", "run", "start"]
```

### Differences from Server

1. **No EXPOSE** - Workers don't listen on HTTP ports
2. **Different source path** - Worker-specific code location
3. **Same base patterns** - Security, optimization, and caching strategies apply

## Building Images

### Manual Build Commands

```bash
# Build server image
docker build \
  -t your-app-server:latest \
  -f docker/server/Dockerfile \
  .

# Build worker image
docker build \
  -t your-app-worker:latest \
  -f docker/worker/Dockerfile \
  .
```

### Build Context Structure

```
distributed-pluggable-api/
├── docker/
│   ├── server/
│   │   └── Dockerfile
│   └── worker/
│       └── Dockerfile
└── compose/
    └── docker-compose.yml
```

## Image Optimization

### Size Comparison

| Image Type | Size | Notes |
|------------|------|-------|
| Full Ubuntu + Runtime | ~900 MB | Includes full OS |
| Debian Slim + Runtime | ~200 MB | Minimal Debian |
| Alpine + Runtime | ~50 MB | Minimal Alpine |
| **Optimized Multi-Stage** | **~60 MB** | Alpine + app code |

### Optimization Techniques

1. Multi-Stage Builds
2. Alpine Base Images
3. Production Dependencies Only
4. Cache Cleaning

## Security Best Practices

1. Non-Root User
2. Minimal Base Image
3. No Secrets in Image
4. Pin Specific Versions
5. Scan for Vulnerabilities

## Docker Compose Integration

### Service Definition

```yaml
your-app-server:
  build:
    context: .
    dockerfile: docker/server/Dockerfile
  ports:
    - '5000:5000'
  env_file:
    - ./.env
  restart: unless-stopped
```

### Build and Run

```bash
# Build images
docker compose build

# Build and start services
docker compose up -d

# Rebuild and restart
docker compose up -d --build

# View logs
docker compose logs -f

# Stop services
docker compose down
```

## Advanced Patterns

### 1. Custom Entrypoint for Migrations

```yaml
your-app-db-migrate:
  build:
    context: .
    dockerfile: docker/server/Dockerfile
  command: ["npm", "run", "migrate:up"]
  env_file:
    - ./.env
  restart: "no"
```

**Use case:** Run database migrations before starting server

### 2. Health Checks

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:5000/health || exit 1
```

**Benefits:**
- Docker knows if container is healthy
- Can restart unhealthy containers automatically

## Adapting for Your Application

### For Node.js Applications

```dockerfile
FROM node:20-alpine
COPY package*.json ./
RUN npm ci --omit=dev
COPY . .
CMD ["node", "server.js"]
```

### For Python Applications

```dockerfile
FROM python:3.11-alpine
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
CMD ["python", "app.py"]
```

### For Go Applications

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o server

FROM alpine:latest
COPY --from=builder /app/server /server
CMD ["/server"]
```

### For Java Applications

```dockerfile
FROM maven:3.9-eclipse-temurin-17 AS builder
WORKDIR /app
COPY pom.xml ./
RUN mvn dependency:go-offline
COPY src ./src
RUN mvn package -DskipTests

FROM eclipse-temurin:17-jre-alpine
COPY --from=builder /app/target/*.jar app.jar
CMD ["java", "-jar", "app.jar"]
```

## Troubleshooting

### Build Failures

**Git clone fails:**
- Check network connectivity
- Verify repository URL and access permissions
- Use SSH keys or access tokens for private repos

**Dependency installation fails:**
- Check network connectivity
- Verify package registry availability, use private registry if needed

### Runtime Issues

**Permission denied:**
- Ensure proper file ownership with `chown`
- Verify `USER` directive is set correctly
- Check volume mount permissions

**Module/dependency not found:**
- Ensure dependency installation completed successfully
- Check that dependencies are copied to final stage

**Container exits immediately:**
- Check logs with `docker logs <container-id>`
- Ensure application doesn't crash on startup
- Check environment variables are set

## Next Steps

- [Infrastructure Setup](../../infrastructure/README.md) - Learn how to set up your own infrastructure.
