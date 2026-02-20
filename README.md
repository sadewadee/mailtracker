# eximmon - cPanel/WHM Exim Mail Monitor

[![Release](https://github.com/sadewadee/mailtracker/actions/workflows/release.yml/badge.svg)](https://github.com/sadewadee/mailtracker/releases)

Real-time spam detection for cPanel/WHM servers. Automatically monitors Exim mail logs and suspends compromised email accounts sending spam.

## Features

- 🚨 **Auto-detect spam** - Suspends accounts exceeding email rate limits
- 🤖 **Telegram & Slack bots** - Real-time notifications and remote control
- 🔄 **UAPI fallback** - Compatible with cPanel v122.x.x and older versions
- 🔐 **Secure config** - Auto-saved with permission 600
- ⚡ **Systemd service** - Auto-start on boot, auto-restart on crash

## Quick Install

```bash
# Download latest release
curl -sL https://github.com/sadewadee/mailtracker/releases/latest/download/eximmon-linux-amd64.tar.gz | tar xz

# Install
chmod +x install.sh && sudo ./install.sh

# Setup config (first time only)
cd /opt/eximmon
API_TOKEN=xxxxx TELEGRAM_BOT_TOKEN=xxx TELEGRAM_ADMIN_IDS=123456 ./eximmon start
# Press Ctrl+C after config saved

# Start service
systemctl start eximmon
systemctl enable eximmon
```

## Requirements

Before starting, configure email rate limits in WHM:
[How to Prevent Spam with Mail Limiting Features](https://documentation.cpanel.net/display/CKB/How+to+Prevent+Spam+with+Mail+Limiting+Features)

1. Login to WHM
2. Go to Development > Manage API Tokens
3. Create token with "Everything" permission

## Configuration

Config saved to `.eximmon.conf` (permission: 600). Set once via environment variables:

```bash
API_TOKEN=xxxxx                      # WHM API token (required)
NOTIFY_EMAIL=admin@example.com       # Email for notifications
EXIM_LOG=/var/log/exim_mainlog       # Exim log path
WHM_API_HOST=127.0.0.1               # WHM hostname
MAX_PER_MIN=8                        # Max emails per minute
MAX_PER_HOUR=100                     # Max emails per hour
PREFER_MODERN_UAPI=true              # Use modern UAPI first

# Telegram Bot
TELEGRAM_BOT_TOKEN=123456:ABC-DEF
TELEGRAM_ADMIN_IDS=123456789
TELEGRAM_NOTIFY_CHAT_ID=-100123456789

# Slack Bot
SLACK_BOT_TOKEN=xoxb-123456-abcdef
SLACK_ADMIN_IDS=U12345,U67890
SLACK_NOTIFY_CHANNEL=C12345
```

## CLI Commands

```bash
./eximmon start       # Start monitoring (continuous)
./eximmon run         # Single run
./eximmon rerun DATE  # Rerun from specific date
./eximmon skip        # Skip existing, monitor new only
./eximmon suspend EMAIL   # Manual suspend
./eximmon unsuspend EMAIL # Manual unsuspend
./eximmon info DOMAIN     # Get domain info
./eximmon config      # Show current config
./eximmon reset       # Reset all data
./eximmon help        # Show help
```

## Bot Commands

| Command | Description | Admin Only |
|---------|-------------|------------|
| `/status` | Check eximmon status | No |
| `/suspend <email>` | Suspend email | Yes |
| `/unsuspend <email>` | Unsuspend email | Yes |
| `/list` | List suspended emails | No |
| `/config` | View configuration | No |
| `/set <key> <value>` | Update threshold | Yes |
| `/whitelist add/remove/list` | Manage whitelist | Yes |

## Bot Setup

### Telegram
1. Create bot via [@BotFather](https://t.me/botfather)
2. Get your user ID from [@userinfobot](https://t.me/userinfobot)
3. (Optional) Add bot to group/channel for notifications

### Slack
1. Create Slack App with Bot User
2. Install to workspace and get Bot User OAuth Token
3. (Optional) Create notification channel and invite bot

## Systemd Commands

```bash
systemctl start eximmon     # Start service
systemctl stop eximmon      # Stop service
systemctl restart eximmon   # Restart service
systemctl status eximmon    # Check status
journalctl -u eximmon -f    # View live logs
```

## API Compatibility

| Method | Port | cPanel Version |
|--------|------|----------------|
| Modern UAPI | 2083 | v122.x.x+ |
| Legacy WHM Proxy | 2087 | All versions |

Auto-fallback from UAPI to legacy if needed.

## How It Works

1. Scans `/var/log/exim_mainlog` continuously (every 15 seconds)
2. Matches authenticated dovecot login entries
3. Counts external recipients per email per minute/hour
4. Suspends accounts exceeding thresholds via WHM API
5. Sends notification to configured channels

## Data Storage

- `.config` - Last scanned position
- `data/<email>/<date>/<hour>` - Hourly counts
- `data/<email>/<date>/<minute>` - Per-minute counts

## Cleanup Old Data

Add to crontab:
```bash
0 1 * * * find /opt/eximmon/data/*/ -type f -mtime +30 -exec rm -Rf {} \;
```

## Development

```bash
# Build
make build

# Build for Linux
make build-linux

# Install (requires root)
sudo make install
```

## License

MIT License - see [LICENSE](LICENSE)

## Credits

Original project by [u007/eximmon](https://github.com/u007/eximmon)
