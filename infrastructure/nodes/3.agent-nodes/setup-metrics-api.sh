#!/bin/bash
set -euo pipefail

# Variables
USER_NAME=$(whoami)
APP_DIR="${APP_DIR:-$HOME/app}"
BINARY_NAME="metrics-api"
SERVICE_NAME="metrics-api"

# Check required commands
if ! command -v systemctl >/dev/null 2>&1; then
    echo "Error: systemctl is required but not installed."
    exit 1
fi

echo "Setting up $SERVICE_NAME..."

if [ ! -f "$APP_DIR/$BINARY_NAME" ]; then
    echo "Error: Binary $BINARY_NAME not found in $APP_DIR"
    echo "Please ensure you have copied the binary to $APP_DIR before running this script."
    exit 1
fi

echo "Making binary executable..."
chmod +x "$APP_DIR/$BINARY_NAME"

echo "Creating systemd service..."
sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null <<EOF
[Unit]
Description=Metrics API Service
After=network.target

[Service]
Type=simple
User=$USER_NAME
WorkingDirectory=$APP_DIR
EnvironmentFile=$APP_DIR/.env
ExecStart=$APP_DIR/$BINARY_NAME
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF


echo "Starting service..."
sudo systemctl daemon-reload
sudo systemctl enable $SERVICE_NAME
sudo systemctl restart $SERVICE_NAME

echo "Checking service status..."
sudo systemctl status $SERVICE_NAME --no-pager

echo "$SERVICE_NAME setup complete!"
