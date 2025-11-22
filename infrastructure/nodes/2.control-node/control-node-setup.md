# Control Node (Ubuntu Server) Setup

Once the [base setup](../1.base-setup/base-setup.md) is complete, we can proceed with setting up the **Control Node** - your main Ubuntu server that will manage the other servers.

To enable automated server management, the Control Node requires **SSH Access to Agent Nodes**. This allows the Control Node to connect to other servers securely.

## Configure SSH Access from Control Node

The Control Node must be able to connect to other servers without manually entering passwords. This is essential for automating operations such as scaling clusters up or down.

To enable this, copy the private SSH key you generated earlier to the control node.

Replace:
* `<path/to/private_ssh_key>` -> with the path to your private SSH key (e.g., `$(pwd)/.ssh/id_ed25519`)
* `<port_number>` -> with the SSH port of the target server (e.g., `2223`)
* `<user>` -> with the username on the target server (e.g., `ubuntu`)
* `<server_ip>` -> with the IP address of the target server (e.g., `192.168.1.101`)

Run the following command from your **local machine**:

```bash
scp -i <path/to/private_ssh_key> -P <port_number> <path/to/private_ssh_key> <user>@<server_ip>:/home/<user>/.ssh/id_ed25519
```

Then, on the **control node**, set correct permissions for the private key:

```bash
chmod 600 /home/$USER/.ssh/id_ed25519
```

### Connect to the Other Servers via Control Node

Once the key is copied and permissions are set, control node can connect to any agent node using:

```bash
ssh -o StrictHostKeyChecking=accept-new -p <port_number> <user>@<server_ip>
```

## Spin up Load Balancer

The Load Balancer runs on the Control Node and distributes incoming traffic across the Agent Nodes for high availability and scalability. We'll deploy and configure it using the provided [load balancer setup script](./setup-load-balancer.sh).

### Copy the Setup Script

create a directory `app` on the control node:

```bash
ssh -i <path/to/private_ssh_key> -p <port_number> <user>@<server_ip> "mkdir -p /home/<user>/app"
```

From your local machine, copy the setup script to the Control Node:

```bash
scp -i <path/to/private_ssh_key> -P <port_number> \
    $(pwd)/infrastructure/nodes/2.control-node/setup-load-balancer.sh \
    <user>@<server_ip>:/home/<user>/app/
```

### Grant Execute Permission

Once copied, SSH into the Control Node and make the script executable:

```bash
chmod +x /home/<user>/app/setup-load-balancer.sh
```

### Run the Setup Script

```bash
/home/<user>/app/setup-load-balancer.sh
```

## Spin up Scaler

The Scaler monitors agent nodes and updates the load balancer configuration.

### Build the Scaler Binary

From your **local machine**, build the binary.

Since the Control Node is Linux (and running on ARM64/aarch64), you need to cross-compile for that architecture:

```bash
# Build the binary for Linux ARM64
GOOS=linux GOARCH=arm64 make build

# OR Build the binary for Linux x86_64
GOOS=linux GOARCH=amd64 make build
```

### Copy Scaler Binary & Setup Script

Create a directory `app` on the control node:

```bash
ssh -i <path/to/private_ssh_key> -p <port_number> <user>@<server_ip> "mkdir -p /home/<user>/app"
```

From your **local machine**, copy the binary, the setup script, and the environment example to the Control Node:

```bash
scp -i <path/to/private_ssh_key> -P <port_number> \
    infrastructure/nodes/2.control-node/scaler/scaler \
    infrastructure/nodes/2.control-node/setup-scaler.sh \
    infrastructure/nodes/2.control-node/scaler/.env.example \
    <user>@<server_ip>:/home/<user>/app/
```

### Configure Scaler

SSH into the Control Node and navigate to the app directory:

```bash
cd /home/<user>/app
```

Create a `.env` file based on the example:

```bash
cp .env.example .env
```

Edit the `.env` file to configure the `SERVER_MANAGER_API` and `AGENTS`.

```bash
nano .env
```

Example `.env` content:

```env
SERVER_MANAGER_API=http://<server_manager_ip>:3000
AGENTS='[
    {
        "server_name": "agent-1",
        "upstream_url": "http://<agent_1_ip>:5001",
        "telemetry_url": "http://<agent_1_ip>:5101/metrics",
        "ssh": {
            "port": "222x",
            "ip": "<agent_1_ip>"
        }
    }
]'

# For Pluggable API
CORS_ORIGINS=http://192.168.1.8:5173
POSTGRES_URL=postgres://username:password@192.168.1.8/learninglabdb
REDIS_URL=redis://192.168.1.8:6379
```

**Environment Configuration:** Scaler reads env vars set on Control Node, transfers them to each Agent Node (into .env file reciding in compose directory) during deployment (via deploy.go), which are then used by docker-compose. Depending on your distributed application, you may need to update environment variables that are passed to scaler and fix the deploy.go file to transfer the correct environment variables to the Agent Node.

### Setup Scaler

Run the setup script to install and start the Scaler as a systemd service:

```bash
chmod +x setup-scaler.sh
./setup-scaler.sh
```

Check the status to ensure it's running:

```bash
sudo systemctl status scaler
```

### Adding Agents Later On

To add new agents to the cluster:

1.  **Provision the new Agent Node** following the [Agent Node Setup](../3.agent-nodes/agent-node-setup.md).
2.  **Update the `.env` file** in `/home/<user>/app/.env` to include the new agent in the `AGENTS` JSON array.
3.  **Restart the Scaler**:
    ```bash
    sudo systemctl restart scaler
    ```

## Firewall Configuration

Since the Trafic will flow from control node (load balancer) to agent nodes, we need to allow the load balancer port from our local machine to access it. The firewall on the Control Node might be blocking the connections to api.

### Load Balancer

You need to allow traffic on port `80`.

```bash
sudo ufw allow 80/tcp
sudo ufw reload
```

## Expose the Port / Forward the Port

Since the Trafic will flow from control node (load balancer) to agent nodes, we need to forward a port from our local machine to access it.

### Load Balancer

Since the Load Balancer will run on port `80` inside the VM, we need to forward a port from our local machine to access it.

1.  Go to **Settings** -> **Network** -> **Adapter 1** -> **Advanced** -> **Port Forwarding**.
2.  Add a new rule:
    *   **Name**: `load-balancer`
    *   **Protocol**: `TCP`
    *   **Host Port**: `8080`
    *   **Guest Port**: `80`
    *   **Guest IP**: *(leave blank)*

Now you can test the Load Balancer from your local machine:

```bash
curl http://192.168.1.8:8080/health
curl http://192.168.1.8:8080/ready
```
