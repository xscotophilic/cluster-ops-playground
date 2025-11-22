# Host Infrastructure

This directory contains the tools and documentation for setting up and managing the **Host Machine** - the foundation of the cluster simulation.

## Contents

- **[Host Setup Guide](host-setup.md)**: Detailed instructions for preparing your machine, installing VirtualBox, and configuring the environment.
- **[Server Manager API](./server-manager-api)**: A Go-based service that allows programmatic control (Start/Stop) of the virtual servers.
- **`load-tester.sh`**: A utility script for performing load tests against the deployed services.

## Role of the Host

In this playground, the Host serves two primary functions:
1. **Virtualization**: It runs the VirtualBox VMs that simulate independent server nodes.
2. **Management Layer**: It hosts the Server Manager API, which acts as a "Cloud Provider" endpoint for automated scaling and node management.
