#!/bin/bash
set -euo pipefail

# Check required commands
# Check required commands
for cmd in systemctl tee; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "Error: Required command '$cmd' not found."
        exit 1
    fi
done

# load environment variables
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
else
    echo ".env file not found!"
fi

# update and install nginx
if ! command -v nginx >/dev/null 2>&1; then
    if command -v apt >/dev/null 2>&1; then
        sudo apt update -y
        sudo apt install -y nginx

        # disable default site
        sudo rm -f /etc/nginx/sites-enabled/default
        sudo rm -f /etc/nginx/sites-available/default
        sudo systemctl restart nginx
    else
        echo "Error: Nginx not found and 'apt' package manager is missing."
        echo "Please install Nginx manually."
        exit 1
    fi
fi

if [ ! -d /etc/nginx/conf.d ]; then
    sudo mkdir -p /etc/nginx/conf.d
fi

UPSTREAM_CONF="/etc/nginx/conf.d/upstream.conf"
LB_CONF="/etc/nginx/conf.d/lb.conf"

# Build upstream.conf
# Placeholder configuration to allow Nginx to start.
# The actual configuration will be managed by the scaler.
{
    echo "upstream backend {"
    echo "    server 127.0.0.1:65535; # Placeholder"
    echo "}"
} | sudo tee "$UPSTREAM_CONF" >/dev/null

# Build lb.conf
if [ ! -f "$LB_CONF" ]; then
    sudo tee "$LB_CONF" >/dev/null <<'EOF'
server {
    listen 80;
    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
EOF
fi

# test & reload
sudo nginx -t
sudo systemctl reload nginx
