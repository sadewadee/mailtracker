#!/bin/bash
#
# Eximmon Installer for cPanel/WHM
# Run as root
#

set -e

INSTALL_DIR="/opt/eximmon"
SERVICE_FILE="/etc/systemd/system/eximmon.service"
CONFIG_FILE="$INSTALL_DIR/.eximmon.conf"
BACKUP_DIR="/opt/eximmon/backups"

echo ""
echo "╔══════════════════════════════════════════════╗"
echo "║       Eximmon Installer v1.3.4               ║"
echo "║   cPanel/WHM Exim Mail Monitor               ║"
echo "╚══════════════════════════════════════════════╝"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: Please run as root (sudo ./install.sh)"
    exit 1
fi

# Check if service is running and stop it
SERVICE_WAS_RUNNING=false
if systemctl is-active --quiet eximmon 2>/dev/null; then
    SERVICE_WAS_RUNNING=true
    echo "[1/5] Stopping eximmon service..."
    systemctl stop eximmon
    echo "Service stopped."
else
    echo "[1/5] Service not running, skipping stop."
fi

# Backup existing binary if exists
if [ -f "$INSTALL_DIR/eximmon" ]; then
    echo "[2/5] Backing up existing binary..."
    mkdir -p "$BACKUP_DIR"
    BACKUP_FILE="$BACKUP_DIR/eximmon.$(date +%Y%m%d_%H%M%S)"
    cp "$INSTALL_DIR/eximmon" "$BACKUP_FILE"
    echo "Backup saved: $BACKUP_FILE"

    # Keep only last 5 backups
    ls -t "$BACKUP_DIR"/eximmon.* 2>/dev/null | tail -n +6 | xargs -r rm -f
else
    echo "[2/5] No existing binary to backup."
fi

# Create installation directory
echo "[3/5] Installing binary..."
mkdir -p "$INSTALL_DIR"

# Copy binary
if [ -f "./eximmon" ]; then
    cp ./eximmon "$INSTALL_DIR/"
elif [ -f "./bin/eximmon" ]; then
    cp ./bin/eximmon "$INSTALL_DIR/"
else
    echo "Error: eximmon binary not found"
    echo "Please run this script from the extracted folder (eximmon/)"
    exit 1
fi
chmod +x "$INSTALL_DIR/eximmon"

# Copy service file
if [ -f "./eximmon.service" ]; then
    cp ./eximmon.service "$SERVICE_FILE"
fi

# Setup config if not exists
echo "[4/5] Checking configuration..."
if [ -f "$CONFIG_FILE" ]; then
    echo "Config exists: $CONFIG_FILE"
    echo "Keeping existing configuration."
else
    echo "No config found. Setting up..."
    echo ""

    # API Token (required)
    while [ -z "$API_TOKEN" ]; do
        read -p "WHM API Token (required): " API_TOKEN
        if [ -z "$API_TOKEN" ]; then
            echo "Error: API Token is required!"
        fi
    done

    # Optional configs
    read -p "Notification Email (optional): " NOTIFY_EMAIL
    read -p "WHM API Host [127.0.0.1]: " WHM_API_HOST
    WHM_API_HOST=${WHM_API_HOST:-127.0.0.1}

    read -p "Max emails per minute [8]: " MAX_PER_MIN
    MAX_PER_MIN=${MAX_PER_MIN:-8}

    read -p "Max emails per hour [100]: " MAX_PER_HOUR
    MAX_PER_HOUR=${MAX_PER_HOUR:-100}

    # Telegram
    echo ""
    echo "Telegram Bot (optional, press Enter to skip):"
    read -p "Bot Token: " TELEGRAM_BOT_TOKEN
    if [ -n "$TELEGRAM_BOT_TOKEN" ]; then
        read -p "Admin User IDs (comma separated): " TELEGRAM_ADMIN_IDS
        read -p "Notification Chat ID: " TELEGRAM_NOTIFY_CHAT_ID
    fi

    # Slack
    echo ""
    echo "Slack Bot (optional, press Enter to skip):"
    read -p "Bot Token: " SLACK_BOT_TOKEN
    if [ -n "$SLACK_BOT_TOKEN" ]; then
        read -p "Admin User IDs (comma separated): " SLACK_ADMIN_IDS
        read -p "Notification Channel ID: " SLACK_NOTIFY_CHANNEL
    fi

    # Create config file
    echo ""
    cat > "$CONFIG_FILE" << EOF
{
  "api_token": "$API_TOKEN",
  "notify_email": "$NOTIFY_EMAIL",
  "exim_log": "/var/log/exim_mainlog",
  "whm_api_host": "$WHM_API_HOST",
  "prefer_modern_uapi": "true",
  "max_per_min": $MAX_PER_MIN,
  "max_per_hour": $MAX_PER_HOUR,
  "telegram_bot_token": "$TELEGRAM_BOT_TOKEN",
  "telegram_admin_ids": "$TELEGRAM_ADMIN_IDS",
  "telegram_notify_chat_id": "$TELEGRAM_NOTIFY_CHAT_ID",
  "slack_bot_token": "$SLACK_BOT_TOKEN",
  "slack_admin_ids": "$SLACK_ADMIN_IDS",
  "slack_notify_channel": "$SLACK_NOTIFY_CHANNEL"
}
EOF
    chmod 600 "$CONFIG_FILE"
    echo "Config saved: $CONFIG_FILE"
fi

# Reload systemd and enable
echo "[5/5] Installing systemd service..."
systemctl daemon-reload
systemctl enable eximmon

# Restart service if it was running before
if [ "$SERVICE_WAS_RUNNING" = true ]; then
    echo ""
    echo "Restarting service..."
    systemctl start eximmon
    sleep 1
    if systemctl is-active --quiet eximmon; then
        echo "Service restarted successfully."
    else
        echo "Warning: Service failed to start. Check logs:"
        echo "  journalctl -u eximmon -n 20"
    fi
fi

echo ""
echo "╔══════════════════════════════════════════════╗"
echo "║          Installation Complete!              ║"
echo "╚══════════════════════════════════════════════╝"
echo ""
echo "Install directory: $INSTALL_DIR"
echo "Config file: $CONFIG_FILE"
[ -d "$BACKUP_DIR" ] && echo "Backups: $BACKUP_DIR"
echo ""
echo "Commands:"
echo "  systemctl status eximmon   - Check status"
echo "  systemctl start eximmon    - Start monitoring"
echo "  systemctl stop eximmon     - Stop monitoring"
echo "  journalctl -u eximmon -f   - View logs"
echo ""
