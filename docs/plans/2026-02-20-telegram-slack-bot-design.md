# Telegram & Slack Bot Integration Design

**Date:** 2026-02-20
**Status:** Approved

## Overview

Integrate Telegram and Slack bot into eximmon binary for real-time notifications and remote control of email suspension operations.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      eximmon (satu binary)                  │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Log Scanner │  │  Bot Engine │  │   Notifier          │  │
│  │ (existing)  │  │ (Telegram + │  │ (email + bot)       │  │
│  │             │  │  Slack)     │  │                     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│         │                │                    ▲             │
│         │                │                    │             │
│         ▼                ▼                    │             │
│  ┌─────────────────────────────────────────────┐            │
│  │            Shared State (memory)             │            │
│  │  - suspended emails list                     │            │
│  │  - rate counters                             │            │
│  │  - config (thresholds, whitelist)            │            │
│  └─────────────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────────┘
```

Bot runs in separate goroutines, communicates via shared memory and channels.

## Deployment Method

**Polling** - Bot actively polls Telegram/Slack servers every few seconds for new messages.

Reasons:
- WHM servers typically behind firewall, no public domain
- Simplest setup
- 1-2 second delay acceptable for this use case

## Security Model

**Admin Whitelist + Notification Channel**

- Only whitelisted user IDs can execute control commands (suspend, unsuspend, config)
- Notifications sent to designated channel/group (all members can view)
- Non-whitelisted users receive rejection message

## Environment Variables

### Telegram
```bash
TELEGRAM_BOT_TOKEN=xxx                      # Bot token from @BotFather
TELEGRAM_ADMIN_IDS=123456,789012            # Comma-separated user IDs with control access
TELEGRAM_NOTIFY_CHAT_ID=-123456789          # Channel/Group for notifications (optional)
```

### Slack
```bash
SLACK_BOT_TOKEN=xoxb-xxx                    # Bot User OAuth Token
SLACK_ADMIN_IDS=U12345,U67890               # Comma-separated user IDs with control access
SLACK_NOTIFY_CHANNEL=C12345                 # Channel ID for notifications (optional)
```

## Bot Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/status` | Check eximmon status (running, uptime, stats) | `/status` |
| `/suspend <email>` | Manually suspend email outgoing | `/suspend spam@domain.com` |
| `/unsuspend <email>` | Unsuspend email outgoing | `/unsuspend user@domain.com` |
| `/list` | List all suspended emails | `/list` |
| `/stats <email|domain>` | Show email statistics | `/stats domain.com` |
| `/config` | Show current configuration | `/config` |
| `/set <key> <value>` | Update threshold config | `/set max_per_min 10` |
| `/whitelist add <email>` | Add email to whitelist | `/whitelist add safe@domain.com` |
| `/whitelist remove <email>` | Remove email from whitelist | `/whitelist remove safe@domain.com` |
| `/whitelist list` | Show whitelist | `/whitelist list` |

## Notification Format

When eximmon auto-suspends an account:

```
🚨 SPAM DETECTED

Email: spammer@domain.com
Domain: domain.com
Rate: 15 emails/min (limit: 8)
Total: 234 emails in last hour

Action: SUSPENDED ✅

Reply /unsuspend spammer@domain.com to restore
```

## File Structure

```
mailtracker/
├── bot/
│   ├── bot.go           # Main bot engine, goroutine launcher
│   ├── telegram.go      # Telegram-specific implementation
│   ├── slack.go         # Slack-specific implementation
│   ├── commands.go      # Command handlers (shared)
│   └── types.go         # Types and interfaces
├── main.go              # Updated: init bot if tokens configured
├── config.go            # New: runtime config management
└── whitelist.go         # New: whitelist management
```

## Dependencies

- `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram Bot API
- `github.com/slack-go/slack` - Slack API

## Implementation Priority

1. Core bot engine with command routing
2. Telegram implementation (simpler API)
3. Slack implementation
4. Notification integration
5. Whitelist persistence
6. Runtime config management
