package bot

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"eximmon/whm"
)

// Engine manages all bot instances and shared state
type Engine struct {
	telegram *TelegramBot
	slack    *SlackBot
	state    *State
	started  time.Time
	mu       sync.RWMutex
}

// NewEngine creates a new bot engine with config from environment
func NewEngine() *Engine {
	config := loadConfigFromEnv()

	if !config.Enabled {
		return nil
	}

	state := &State{
		SuspendedEmails: make(map[string]SuspendedInfo),
		Whitelist:       make(map[string]bool),
		Config: RuntimeConfig{
			MaxPerMin:  8,
			MaxPerHour: 100,
		},
	}

	return &Engine{
		state:   state,
		started: time.Now(),
	}
}

func loadConfigFromEnv() Config {
	config := Config{}

	// Telegram
	config.TelegramToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if config.TelegramToken != "" {
		config.Enabled = true
		if ids := os.Getenv("TELEGRAM_ADMIN_IDS"); ids != "" {
			for _, idStr := range strings.Split(ids, ",") {
				if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil {
					config.TelegramAdminIDs = append(config.TelegramAdminIDs, id)
				}
			}
		}
		if id := os.Getenv("TELEGRAM_NOTIFY_CHAT_ID"); id != "" {
			config.TelegramNotifyChat, _ = strconv.ParseInt(id, 10, 64)
		}
	}

	// Slack
	config.SlackToken = os.Getenv("SLACK_BOT_TOKEN")
	if config.SlackToken != "" {
		config.Enabled = true
		if ids := os.Getenv("SLACK_ADMIN_IDS"); ids != "" {
			config.SlackAdminIDs = strings.Split(strings.ReplaceAll(ids, " ", ""), ",")
		}
		config.SlackNotifyChannel = os.Getenv("SLACK_NOTIFY_CHANNEL")
	}

	return config
}

// Start launches all configured bot instances
func (e *Engine) Start() error {
	if e == nil {
		return nil
	}

	config := loadConfigFromEnv()
	handler := e.createHandler()

	// Start Telegram
	if config.TelegramToken != "" {
		tg, err := NewTelegramBot(config, e.state, handler)
		if err != nil {
			return fmt.Errorf("failed to create telegram bot: %w", err)
		}
		if tg != nil {
			e.telegram = tg
			if err := tg.Start(); err != nil {
				return fmt.Errorf("failed to start telegram bot: %w", err)
			}
			Log("Telegram bot started: @%s", tg.GetBotUsername())
		}
	}

	// Start Slack
	if config.SlackToken != "" {
		sl, err := NewSlackBot(config, e.state, handler)
		if err != nil {
			Log("Warning: failed to create slack bot: %v", err)
		} else if sl != nil {
			e.slack = sl
			if err := sl.Start(); err != nil {
				return fmt.Errorf("failed to start slack bot: %w", err)
			}
			Log("Slack bot started")
		}
	}

	return nil
}

// Stop gracefully stops all bots
func (e *Engine) Stop() {
	if e == nil {
		return
	}
	if e.telegram != nil {
		e.telegram.Stop()
	}
	if e.slack != nil {
		e.slack.Stop()
	}
}

// createHandler returns the command handler function
func (e *Engine) createHandler() CommandHandler {
	return func(cmd Command, state *State) string {
		return e.handleCommand(cmd, state)
	}
}

// handleCommand processes a command and returns response
func (e *Engine) handleCommand(cmd Command, state *State) string {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch cmd.Type {
	case CmdStatus:
		uptime := time.Since(e.started).Round(time.Second)
		return FormatStatusMessage(
			uptime.String(),
			len(state.SuspendedEmails),
			len(state.Whitelist),
		)

	case CmdSuspend:
		email := cmd.Args[0]
		if err := whm.SuspendEmail(email); err != nil {
			return fmt.Sprintf("❌ Failed to suspend %s: %v", email, err)
		}
		state.SuspendedEmails[email] = SuspendedInfo{
			Email:       email,
			SuspendedAt: time.Now(),
			Reason:      "Manual suspension via bot",
		}
		return fmt.Sprintf("✅ Suspended: `%s`", email)

	case CmdUnsuspend:
		email := cmd.Args[0]
		if err := whm.UnSuspendEmail(email); err != nil {
			return fmt.Sprintf("❌ Failed to unsuspend %s: %v", email, err)
		}
		delete(state.SuspendedEmails, email)
		return fmt.Sprintf("✅ Unsuspended: `%s`", email)

	case CmdList:
		if len(state.SuspendedEmails) == 0 {
			return "📋 No suspended emails"
		}
		var sb strings.Builder
		sb.WriteString("📋 *Suspended Emails:*\n\n")
		for email, info := range state.SuspendedEmails {
			sb.WriteString(fmt.Sprintf("• `%s` - %s\n", email, info.SuspendedAt.Format("2006-01-02 15:04")))
		}
		return sb.String()

	case CmdStats:
		// TODO: Implement stats lookup from data directory
		return fmt.Sprintf("📊 Stats for `%s` - Feature coming soon", cmd.Args[0])

	case CmdConfig:
		return fmt.Sprintf("⚙️ *Configuration:*\n• Max Per Min: %d\n• Max Per Hour: %d",
			state.Config.MaxPerMin, state.Config.MaxPerHour)

	case CmdSet:
		key := cmd.Args[0]
		value, err := strconv.ParseInt(cmd.Args[1], 10, 16)
		if err != nil {
			return "❌ Invalid value"
		}
		switch key {
		case "max_per_min", "MAX_PER_MIN":
			state.Config.MaxPerMin = int16(value)
			return fmt.Sprintf("✅ Set max_per_min to %d", value)
		case "max_per_hour", "MAX_PER_HOUR":
			state.Config.MaxPerHour = int16(value)
			return fmt.Sprintf("✅ Set max_per_hour to %d", value)
		default:
			return fmt.Sprintf("❌ Unknown config key: %s", key)
		}

	case CmdWhitelistAdd:
		email := cmd.Args[0]
		state.Whitelist[email] = true
		return fmt.Sprintf("✅ Added to whitelist: `%s`", email)

	case CmdWhitelistRemove:
		email := cmd.Args[0]
		delete(state.Whitelist, email)
		return fmt.Sprintf("✅ Removed from whitelist: `%s`", email)

	case CmdWhitelistList:
		if len(state.Whitelist) == 0 {
			return "📋 Whitelist is empty"
		}
		var sb strings.Builder
		sb.WriteString("📋 *Whitelisted Emails:*\n\n")
		for email := range state.Whitelist {
			sb.WriteString(fmt.Sprintf("• `%s`\n", email))
		}
		return sb.String()

	default:
		return "❓ Unknown command. Try /status, /list, /suspend, /unsuspend, /config, /whitelist"
	}
}

// NotifySuspension sends notification to all configured platforms
func (e *Engine) NotifySuspension(info SuspendedInfo) error {
	if e == nil {
		return nil
	}

	e.mu.Lock()
	e.state.SuspendedEmails[info.Email] = info
	e.mu.Unlock()

	var err error
	if e.telegram != nil {
		if e := e.telegram.NotifySuspension(info); e != nil {
			err = e
		}
	}
	if e.slack != nil {
		if e := e.slack.NotifySuspension(info); e != nil {
			err = e
		}
	}
	return err
}

// NotifyUnsuspend sends notification to all configured platforms
func (e *Engine) NotifyUnsuspend(email string) error {
	if e == nil {
		return nil
	}

	e.mu.Lock()
	delete(e.state.SuspendedEmails, email)
	e.mu.Unlock()

	var err error
	if e.telegram != nil {
		if e := e.telegram.NotifyUnsuspend(email); e != nil {
			err = e
		}
	}
	if e.slack != nil {
		if e := e.slack.NotifyUnsuspend(email); e != nil {
			err = e
		}
	}
	return err
}

// IsWhitelisted checks if an email is in the whitelist
func (e *Engine) IsWhitelisted(email string) bool {
	if e == nil {
		return false
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state.Whitelist[email]
}

// GetConfig returns current runtime config
func (e *Engine) GetConfig() RuntimeConfig {
	if e == nil {
		return RuntimeConfig{MaxPerMin: 8, MaxPerHour: 100}
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state.Config
}
