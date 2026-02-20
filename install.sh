#!/bin/bash
#
# Eximmon Installer for cPanel/WHM
# Run as root
#

set -e

INSTALL_DIR="/opt/eximmon"
SERVICE_FILE="/etc/systemd/system/eximmon.service"

echo "=== Eximmon Installer ==="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: Please run as root"
    exit 1
fi

# Create installation directory
echo "Creating directory: $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

# Copy binary
echo "Copying binary..."
cp bin/eximmon "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/eximmon"

# Copy service file
echo "Installing systemd service..."
cp eximmon.service "$SERVICE_FILE"

# Reload systemd
echo "Reloading systemd daemon..."
systemctl daemon-reload

# Enable service
echo "Enabling eximmon service..."
systemctl enable eximmon

echo ""
echo "=== Installation Complete ==="
echo ""
echo "Config file will be created at: $INSTALL_DIR/.eximmon.conf"
echo ""
echo "Setup your config (run once):"
echo "  cd $INSTALL_DIR && API_TOKEN=xxx TELEGRAM_BOT_TOKEN=xxx TELEGRAM_ADMIN_IDS=123 ./eximmon start"
echo "  (Press Ctrl+C after config is saved)"
echo ""
echo "Then start the service:"
echo "  systemctl start eximmon"
echo ""
echo "Useful commands:"
echo "  systemctl status eximmon   - Check status"
echo "  systemctl restart eximmon  - Restart service"
echo "  systemctl stop eximmon     - Stop service"
echo "  journalctl -u eximmon -f   - View logs"
echo "  ./eximmon config           - View config"
echo ""
