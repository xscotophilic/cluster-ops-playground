# Distributed API Overview

This document provides details on the distributed application that runs on agent nodes in the Cluster Operations Playground.

## What is the Distributed API?

The "distributed API" refers to the containerized application deployed across multiple agent nodes. The architecture is **pluggable** - you can deploy any containerized application, but this project uses [async-fibonacci-lab](https://github.com/xscotophilic/async-fibonacci-lab) as the example application.

## Why "Pluggable"?

The infrastructure is designed to be application-agnostic:

- **Separation of Concerns** - Infrastructure code is separate from application code
- **Easy Replacement** - Swap out the example app with your own

## Example Application: async-fibonacci-lab

The default application demonstrates:

- **API Server** - REST API for handling requests
- **Background Workers** - Async job processing
- **Stateless Design** - Horizontally scalable architecture

## Key Features

### Build-Time Git Clone

The application code is fetched at Docker build time:

**Benefits:**
- Always builds with latest upstream code
- No need to commit application code to this repository
- Clean separation between infrastructure and application
- Easy to update by rebuilding images

### Stateless Design

- **No Local Storage** - All data in external Databases
- **No Session State** - Each request is independent
- **Horizontally Scalable** - Run multiple instances without coordination

## Environment Configuration

The example application requires these environment variables:

```env
# Database
DATABASE_URL=postgresql://user:password@host:port/database

# Redis
REDIS_URL=redis://host:port

# Application
PORT=5000
```

**Environment Configuration:** Scaler reads env vars set on Control Node, transfers them to each Agent Node (into .env file reciding in compose directory) during deployment (via deploy.go), which are then used by docker-compose. Depending on your distributed application, you may need to update environment variables that are passed to scaler and fix the deploy.go file to transfer the correct environment variables to the Agent Node.

## Replacing with Your Own Application

To use your own application instead of example application:

1. **Update Dockerfiles** in `distributed-pluggable-api/docker/`
   - Change git clone URL to your repository
   - Adjust build steps for your stack
   - Ensure health check endpoints exist

2. **Update Docker Compose** in `distributed-pluggable-api/compose/`
   - Modify service definitions
   - Update environment variables
   - Adjust port mappings

3. **Ensure Stateless Design**
   - No local file storage
   - External databases for all state
   - Graceful shutdown handling

Note: If you are using a different language or framework, you may need to adjust the Dockerfiles and Docker Compose files accordingly. Ensure Stateless Design is a must if you want to horizontally scale your application without much complexity and without the need to coordinate between instances. You can go with stateful design if you do understand the design and are willing to change the infrastructure code.

## Next Steps

- [Deployment](./deployment.md) - Learn about how infrastructure deploys pluggable app on agent nodes.
