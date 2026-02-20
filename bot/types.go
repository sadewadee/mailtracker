package bot

import "time"

// Config holds bot configuration
type Config struct {
	// Telegram
	TelegramToken      string
	TelegramAdminIDs   []int64
	TelegramNotifyChat int64

	// Slack
	SlackToken        string
	SlackAdminIDs     []string
	SlackNotifyChannel string

	// Shared
	Enabled bool
}

// State holds shared runtime state
type State struct {
	SuspendedEmails map[string]SuspendedInfo
	Whitelist       map[string]bool
	Config          RuntimeConfig
}

// SuspendedInfo tracks suspension details
type SuspendedInfo struct {
	Email       string
	Domain      string
	SuspendedAt time.Time
	Reason      string
	RatePerMin  int
	RatePerHour int
}

// RuntimeConfig holds adjustable settings
type RuntimeConfig struct {
	MaxPerMin  int16
	MaxPerHour int16
}

// Command represents a parsed bot command
type Command struct {
	Type     CommandType
	Args     []string
	Platform string // "telegram" or "slack"
	UserID   string
	ChatID   string
}

// CommandType enum
type CommandType int

const (
	CmdUnknown CommandType = iota
	CmdStatus
	CmdSuspend
	CmdUnsuspend
	CmdList
	CmdStats
	CmdConfig
	CmdSet
	CmdWhitelistAdd
	CmdWhitelistRemove
	CmdWhitelistList
)

// Bot interface for platform implementations
type Bot interface {
	Start(state *State) error
	Stop()
	IsAdmin(userID string) bool
	SendNotification(message string) error
}

// Notifier interface for sending alerts
type Notifier interface {
	NotifySuspension(info SuspendedInfo) error
	NotifyUnsuspend(email string) error
}
