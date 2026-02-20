# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**eximmon** is a Go application that monitors cPanel/WHM Exim mail logs to detect and automatically suspend compromised email accounts sending spam. It scans `/var/log/exim_mainlog` in real-time, tracks per-minute and per-hour email rates per authenticated user, and suspends accounts exceeding configured thresholds via the WHM API.

## Build Commands

```bash
# Build locally
make build
# or
go build -o bin/eximmon main.go

# Build for Linux production
GOOS=linux go build -o bin/eximmon main.go

# Run during development
go run main.go <command>
```

## Commands

- `start` - Continue from last position or start from yesterday, loop continuously
- `run` - Single run from last position
- `rerun <date>` - Rerun from specific date (format: "2006-01-02" or "2006-01-02 15:04:05")
- `skip` - Skip existing data, monitor only new logs
- `reset` - Clear all stored data and config
- `suspend <email>` - Manually suspend outgoing email
- `unsuspend <email>` - Manually unsuspend outgoing email
- `info <domain>` - Get WHM account info for domain
- `test-notify` - Test notification email
- `help` - Show help

## Environment Variables

Required for production:
- `API_TOKEN` - WHM API token (required)
- `NOTIFY_EMAIL` - Email address for suspension notifications
- `EXIM_LOG` - Path to exim_mainlog (default: "exim_mainlog")
- `WHM_API_HOST` - WHM server hostname (default: "127.0.0.1")

Rate limiting:
- `MAX_PER_MIN` - Max emails per minute per account (default: 8)
- `MAX_PER_HOUR` - Max emails per hour per account (default: 100)

API method:
- `PREFER_MODERN_UAPI` - Use modern UAPI first (default: "true", set to "false" for legacy only)

## Architecture

```
main.go           - Entry point, log scanner loop, rate limiting logic
exim/exim.go      - Exim log date parsing utilities
whm/              - WHM/cPanel API integration
  ├── dialer.go       - TLS connection dialers for WHM (port 2087) and cPanel (port 2083)
  ├── email.go        - Suspend/unsuspend email via cPanel API
  ├── account.go      - Account summary queries
  ├── domain.go       - Domain info queries
  ├── user_data.go    - Domain user data lookup
  └── response.go     - API response types and shared variables
tools/file.go     - File system utilities
```

### Data Flow

1. `eximLogScanner()` reads the Exim log file line by line
2. Regex matches authenticated dovecot login entries (`A=dovecot_*`)
3. For each outgoing email (`<=`), extracts sender email and recipients
4. Skips internal emails (same domain sender/recipient)
5. Counts external recipients per minute/hour using file-based storage
6. When thresholds exceeded, calls `whm.SuspendEmail()` via cPanel API
7. Optionally sends notification email via system `mail` command

### Data Storage

- `.config` - Stores last scanned line position and timestamp
- `data/<email>/<date>/<hour>` - Hourly count per email
- `data/<email>/<date>/<minute>` - Per-minute count per email

Email paths are sanitized (`@` and `-` replaced with `_`).

## WHM API Integration

The `whm` package supports both modern UAPI and legacy WHM proxy methods with automatic fallback:

### Modern UAPI (Preferred)
- Port: 2083 (cPanel)
- Endpoint: `/execute/Email/suspend_outgoing?email=xxx`
- Auth: `Authorization: cpanel <username>:<token>`
- Used by default for cPanel v122.x.x and newer

### Legacy WHM Proxy
- Port: 2087 (WHM)
- Endpoint: `/json-api/cpanel?cpanel_jsonapi_user=<user>&cpanel_jsonapi_apiversion=3&cpanel_jsonapi_module=Email&cpanel_jsonapi_func=suspend_outgoing`
- Auth: `Authorization: whm root:<token>`
- Used as fallback for older cPanel versions

### Flow
1. If `PREFER_MODERN_UAPI=true` (default), try modern UAPI first
2. If modern UAPI fails, automatically fall back to legacy WHM proxy
3. If `PREFER_MODERN_UAPI=false`, use legacy WHM proxy only

### Key endpoints used:
- `/json-api/domainuserdata` (WHM API 1) - Get user info for a domain
- `/execute/Email/suspend_outgoing` (UAPI) - Suspend email via modern API
- `/execute/Email/unsuspend_outgoing` (UAPI) - Unsuspend email via modern API
- `/json-api/cpanel` (cPanel API v3 via WHM proxy) - Legacy suspend/unsuspend

## Bot Integration

The `bot` package provides Telegram and Slack integration for notifications and remote control.

### Architecture
```
bot/
├── bot.go        # Engine orchestrator, starts/stops bots
├── telegram.go   # Telegram implementation (polling)
├── slack.go      # Slack implementation (RTM)
├── commands.go   # Command parser and formatters
└── types.go      # Types and interfaces
```

### Environment Variables
```bash
# Telegram
TELEGRAM_BOT_TOKEN=xxx
TELEGRAM_ADMIN_IDS=123,456
TELEGRAM_NOTIFY_CHAT_ID=-123

# Slack
SLACK_BOT_TOKEN=xoxb-xxx
SLACK_ADMIN_IDS=U123,U456
SLACK_NOTIFY_CHANNEL=C123
```

### Commands
All commands defined in `bot/commands.go`. Handler in `bot/bot.go` processes commands and interacts with `whm` package.

### Integration in main.go
```go
botEngine = bot.NewEngine()
botEngine.Start()
defer botEngine.Stop()

// On suspension:
botEngine.NotifySuspension(bot.SuspendedInfo{...})
```
