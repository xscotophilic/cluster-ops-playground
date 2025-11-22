# Host Machine Setup

Your "host" is just your physical computer - the one running VirtualBox or whatever virtualization software you're using.

### Why do we need a host?

Simple: we need multiple "servers" to create a realistic cluster. Instead of buying actual hardware, we spin up VMs that act like separate machines. Your host manages all of these VMs.

## Prepare Your Machine

Make sure your machine is ready to handle the load. You'll be running at least 3 Virtual Machines at once (1 Control Node + 2 Agent Nodes), plus the Host software.

## Install VirtualBox

We use **VirtualBox** to create our virtual servers. It's free, open-source, and works everywhere.

1.  Go to the [VirtualBox Downloads page](https://www.virtualbox.org/wiki/Downloads).
2.  Download and install the version for your OS (Windows, macOS, or Linux).
3.  **Verification**: Open your terminal and type:
    ```bash
    VBoxManage --version
    ```
    If you see a version number (like `x.x.x`), you're golden!

## Network Magic

> **Note:** You will configure these rules individually when in future you create & setup each VM.

## Set Up the Server Manager API

We have a custom API that lets us control these VMs using code.

1. **Go to the directory**:
    ```bash
    cd infrastructure/host/server-manager-api
    ```
2. **Create your config**:
    Copy the example file to `.env`:
    ```bash
    cp .env.example .env
    ```
3. **Edit the config**:
    Open `.env` and list the names of the VMs you created in VirtualBox and set the port on which the API will listen.
    ```env
    SERVERS=control-node,agent-node-one,agent-node-two
    PORT=3000
    ```
4. **Run it**:
    ```bash
    go run main.go
    ```

Now, your Host is listening! You can send HTTP requests to `x.x.x.x:3000` to turn your virtual servers on and off.

You can find your local IP address by running:

```bash
# Linux
hostname -I

# macOS
ipconfig getifaddr en0 || ipconfig getifaddr en1

# Windows
ipconfig
```

Use the IP address that looks like `192.168.x.x` or `10.x.x.x` (your local network IP).
