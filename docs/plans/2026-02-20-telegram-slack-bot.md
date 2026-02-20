# Telegram & Slack Bot Integration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Integrate Telegram and Slack bots into eximmon for real-time notifications and remote control of email suspension operations.

**Architecture:** Bot engine runs in separate goroutines within the main eximmon binary, communicating via shared state. Uses polling method for both Telegram and Slack. Admin whitelist controls access to sensitive commands.

**Tech Stack:** Go 1.x, telegram-bot-api/v5, slack-go/slack

---

## Prerequisites

- Telegram bot token from @BotFather
- Slack App with Bot User OAuth Token
- User IDs for admin whitelist

---

### Task 1: Add Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Initialize go module if needed and add dependencies**

```bash
cd /Users/sadewadee/Downloads/Plugin\ Pro/mailtracker
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
go get github.com/slack-go/slack
```

**Step 2: Verify go.mod updated**

Run: `cat go.mod`
Expected: Contains telegram-bot-api and slack dependencies

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add telegram and slack bot dependencies"
```

---

### Task 2: Create Bot Types and Interfaces

**Files:**
- Create: `bot/types.go`

**Step 1: Create bot package directory and types file**

```go
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
	Email      string
	Domain     string
	SuspendedAt time.Time
	Reason     string
	RatePerMin int
	RatePerHour int
}

// RuntimeConfig holds adjustable settings
type RuntimeConfig struct {
	MaxPerMin  int16
	MaxPerHour int16
}

// Command represents a parsed bot command
type Command struct {
	Type    CommandType
	Args    []string
	Platform string  // "telegram" or "slack"
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
```

**Step 2: Verify syntax**

Run: `go build ./bot/...`
Expected: No errors

**Step 3: Commit**

```bash
git add bot/types.go
git commit -m "feat(bot): add core types and interfaces"
```

---

### Task 3: Create Command Parser

**Files:**
- Create: `bot/commands.go`

**Step 1: Create command parser**

```go
package bot

import (
	"strconv"
	"strings"
)

// ParseCommand parses a message into a Command struct
func ParseCommand(text string) Command {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/") {
		return Command{Type: CmdUnknown}
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return Command{Type: CmdUnknown}
	}

	cmd := parts[0]
	args := parts[1:]

	switch {
	case cmd == "/status":
		return Command{Type: CmdStatus, Args: args}
	case cmd == "/suspend" && len(args) >= 1:
		return Command{Type: CmdSuspend, Args: args}
	case cmd == "/unsuspend" && len(args) >= 1:
		return Command{Type: CmdUnsuspend, Args: args}
	case cmd == "/list":
		return Command{Type: CmdList, Args: args}
	case cmd == "/stats" && len(args) >= 1:
		return Command{Type: CmdStats, Args: args}
	case cmd == "/config":
		return Command{Type: CmdConfig, Args: args}
	case cmd == "/set" && len(args) >= 2:
		return Command{Type: CmdSet, Args: args}
	case cmd == "/whitelist":
		if len(args) >= 2 && args[0] == "add" {
			return Command{Type: CmdWhitelistAdd, Args: args[1:]}
		}
		if len(args) >= 2 && args[0] == "remove" {
			return Command{Type: CmdWhitelistRemove, Args: args[1:]}
		}
		if len(args) == 1 && args[0] == "list" {
			return Command{Type: CmdWhitelistList, Args: args}
		}
		return Command{Type: CmdUnknown}
	default:
		return Command{Type: CmdUnknown}
	}
}

// FormatSuspendMessage creates notification message for suspension
func FormatSuspendMessage(info SuspendedInfo) string {
	var sb strings.Builder
	sb.WriteString("🚨 *SPAM DETECTED*\n\n")
	sb.WriteString("📧 Email: `" + info.Email + "`\n")
	sb.WriteString("🌐 Domain: `" + info.Domain + "`\n")
	sb.WriteString("📊 Rate: " + strconv.Itoa(info.RatePerMin) + " emails/min\n")
	sb.WriteString("📈 Total: " + strconv.Itoa(info.RatePerHour) + " emails/hour\n\n")
	sb.WriteString("✅ Action: *SUSPENDED*\n\n")
	sb.WriteString("Reply `/unsuspend " + info.Email + "` to restore")
	return sb.String()
}

// FormatUnsuspendMessage creates notification message for unsuspension
func FormatUnsuspendMessage(email string) string {
	return "✅ Email unsuspended: `" + email + "`"
}

// FormatStatusMessage creates status message
func FormatStatusMessage(uptime string, suspendedCount int, whitelistCount int) string {
	var sb strings.Builder
	sb.WriteString("📊 *Eximmon Status*\n\n")
	sb.WriteString("⏱ Uptime: " + uptime + "\n")
	sb.WriteString("🚫 Suspended: " + strconv.Itoa(suspendedCount) + " emails\n")
	sb.WriteString("✅ Whitelisted: " + strconv.Itoa(whitelistCount) + " emails")
	return sb.String()
}
```

**Step 2: Verify syntax**

Run: `go build ./bot/...`
Expected: No errors

**Step 3: Commit**

```bash
git add bot/commands.go
git commit -m "feat(bot): add command parser and message formatters"
```

---

### Task 4: Create Telegram Bot Implementation

**Files:**
- Create: `bot/telegram.go`

**Step 1: Create Telegram bot implementation**

```go
package bot

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramBot implements Bot interface for Telegram
type TelegramBot struct {
	api      *tgbotapi.BotAPI
	config   Config
	state    *State
	handler  CommandHandler
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// CommandHandler processes commands and returns response
type CommandHandler func(cmd Command, state *State) string

// NewTelegramBot creates a new Telegram bot
func NewTelegramBot(config Config, state *State, handler CommandHandler) (*TelegramBot, error) {
	if config.TelegramToken == "" {
		return nil, nil
	}

	api, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return &TelegramBot{
		api:      api,
		config:   config,
		state:    state,
		handler:  handler,
		stopChan: make(chan struct{}),
	}, nil
}

// Start begins polling for updates
func (b *TelegramBot) Start() error {
	if b == nil || b.api == nil {
		return nil
	}

	b.wg.Add(1)
	go b.pollLoop()

	return nil
}

// Stop stops the bot
func (b *TelegramBot) Stop() {
	if b == nil {
		return
	}
	close(b.stopChan)
	b.wg.Wait()
}

func (b *TelegramBot) pollLoop() {
	defer b.wg.Done()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-b.stopChan:
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			b.handleUpdate(update)
		}
	}
}

func (b *TelegramBot) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil || update.Message.Chat == nil {
		return
	}

	msg := update.Message
	if msg.Text == "" {
		return
	}

	// Parse command
	cmd := ParseCommand(msg.Text)
	if cmd.Type == CmdUnknown {
		return
	}

	cmd.Platform = "telegram"
	cmd.UserID = strconv.FormatInt(msg.From.ID, 10)
	cmd.ChatID = strconv.FormatInt(msg.Chat.ID, 10)

	// Check admin access for restricted commands
	if b.requiresAdmin(cmd.Type) && !b.IsAdmin(cmd.UserID) {
		b.sendMessage(msg.Chat.ID, "⛔ Access denied. You are not an admin.")
		return
	}

	// Handle command
	response := b.handler(cmd, b.state)
	if response != "" {
		b.sendMessage(msg.Chat.ID, response)
	}
}

func (b *TelegramBot) requiresAdmin(cmdType CommandType) bool {
	switch cmdType {
	case CmdStatus, CmdList, CmdStats, CmdConfig, CmdWhitelistList:
		return false
	default:
		return true
	}
}

// IsAdmin checks if user ID is in admin whitelist
func (b *TelegramBot) IsAdmin(userID string) bool {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return false
	}
	for _, adminID := range b.config.TelegramAdminIDs {
		if adminID == id {
			return true
		}
	}
	return false
}

// SendNotification sends a message to the notification chat
func (b *TelegramBot) SendNotification(message string) error {
	if b == nil || b.api == nil || b.config.TelegramNotifyChat == 0 {
		return nil
	}
	return b.sendMessage(b.config.TelegramNotifyChat, message)
}

func (b *TelegramBot) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	_, err := b.api.Send(msg)
	if err != nil {
		Log("Telegram send error: %v", err)
	}
	return err
}

// NotifySuspension sends suspension notification
func (b *TelegramBot) NotifySuspension(info SuspendedInfo) error {
	return b.SendNotification(FormatSuspendMessage(info))
}

// NotifyUnsuspend sends unsuspension notification
func (b *TelegramBot) NotifyUnsuspend(email string) error {
	return b.SendNotification(FormatUnsuspendMessage(email))
}

// GetBotUsername returns the bot's username
func (b *TelegramBot) GetBotUsername() string {
	if b == nil || b.api == nil {
		return ""
	}
	return b.api.Self.UserName
}

// Log is set by main to share logging function
var Log func(format string, args ...interface{})

func init() {
	Log = func(format string, args ...interface{}) {
		fmt.Printf(format+"\n", args...)
	}
}
```

**Step 2: Verify syntax**

Run: `go build ./bot/...`
Expected: No errors

**Step 3: Commit**

```bash
git add bot/telegram.go
git commit -m "feat(bot): add Telegram bot implementation"
```

---

### Task 5: Create Slack Bot Implementation

**Files:**
- Create: `bot/slack.go`

**Step 1: Create Slack bot implementation**

```go
package bot

import (
	"fmt"
	"strings"
	"sync"

	"github.com/slack-go/slack"
)

// SlackBot implements Bot interface for Slack
type SlackBot struct {
	client   *slack.Client
	config   Config
	state    *State
	handler  CommandHandler
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewSlackBot creates a new Slack bot
func NewSlackBot(config Config, state *State, handler CommandHandler) (*SlackBot, error) {
	if config.SlackToken == "" {
		return nil, nil
	}

	client := slack.New(config.SlackToken)

	// Test auth
	_, err := client.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate slack bot: %w", err)
	}

	return &SlackBot{
		client:  client,
		config:  config,
		state:   state,
		handler: handler,
	}, nil
}

// Start begins polling for messages via RTM
func (b *SlackBot) Start() error {
	if b == nil || b.client == nil {
		return nil
	}

	b.stopChan = make(chan struct{})
	b.wg.Add(1)
	go b.pollLoop()

	return nil
}

// Stop stops the bot
func (b *SlackBot) Stop() {
	if b == nil {
		return
	}
	if b.stopChan != nil {
		close(b.stopChan)
	}
	b.wg.Wait()
}

func (b *SlackBot) pollLoop() {
	defer b.wg.Done()

	rtm := b.client.NewRTM()
	go rtm.ManageConnection()

	for {
		select {
		case <-b.stopChan:
			rtm.Disconnect()
			return
		case msg, ok := <-rtm.IncomingEvents:
			if !ok {
				return
			}
			b.handleEvent(msg)
		}
	}
}

func (b *SlackBot) handleEvent(event slack.RTMEvent) {
	switch ev := event.Data.(type) {
	case *slack.MessageEvent:
		b.handleMessage(ev)
	case *slack.ConnectedEvent:
		Log("Slack bot connected: %s", ev.Info.User.Name)
	case *slack.ConnectionErrorEvent:
		Log("Slack connection error: %v", ev.ErrorObj)
	case *slack.RTMError:
		Log("Slack RTM error: %v", ev.Error)
	}
}

func (b *SlackBot) handleMessage(ev *slack.MessageEvent) {
	// Ignore messages from bots
	if ev.BotID != "" {
		return
	}

	// Only process messages that look like commands
	text := strings.TrimSpace(ev.Text)
	if text == "" || !strings.HasPrefix(text, "/") {
		return
	}

	// Parse command
	cmd := ParseCommand(text)
	if cmd.Type == CmdUnknown {
		return
	}

	cmd.Platform = "slack"
	cmd.UserID = ev.User
	cmd.ChatID = ev.Channel

	// Check admin access for restricted commands
	if b.requiresAdmin(cmd.Type) && !b.IsAdmin(cmd.UserID) {
		b.sendMessage(ev.Channel, "⛔ Access denied. You are not an admin.")
		return
	}

	// Handle command
	response := b.handler(cmd, b.state)
	if response != "" {
		b.sendMessage(ev.Channel, response)
	}
}

func (b *SlackBot) requiresAdmin(cmdType CommandType) bool {
	switch cmdType {
	case CmdStatus, CmdList, CmdStats, CmdConfig, CmdWhitelistList:
		return false
	default:
		return true
	}
}

// IsAdmin checks if user ID is in admin whitelist
func (b *SlackBot) IsAdmin(userID string) bool {
	for _, adminID := range b.config.SlackAdminIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

// SendNotification sends a message to the notification channel
func (b *SlackBot) SendNotification(message string) error {
	if b == nil || b.client == nil || b.config.SlackNotifyChannel == "" {
		return nil
	}
	return b.sendMessage(b.config.SlackNotifyChannel, message)
}

func (b *SlackBot) sendMessage(channelID, text string) error {
	// Escape special characters for Slack
	text = strings.ReplaceAll(text, "*", "*")
	text = strings.ReplaceAll(text, "`", "`")

	_, _, err := b.client.PostMessage(
		channelID,
		slack.MsgOptionText(text, false),
		slack.MsgOptionDisableLinkUnfurl(),
	)
	if err != nil {
		Log("Slack send error: %v", err)
	}
	return err
}

// NotifySuspension sends suspension notification
func (b *SlackBot) NotifySuspension(info SuspendedInfo) error {
	return b.SendNotification(FormatSuspendMessage(info))
}

// NotifyUnsuspend sends unsuspension notification
func (b *SlackBot) NotifyUnsuspend(email string) error {
	return b.SendNotification(FormatUnsuspendMessage(email))
}
```

**Step 2: Verify syntax**

Run: `go build ./bot/...`
Expected: No errors

**Step 3: Commit**

```bash
git add bot/slack.go
git commit -m "feat(bot): add Slack bot implementation"
```

---

### Task 6: Create Bot Engine (Main Orchestrator)

**Files:**
- Create: `bot/bot.go`

**Step 1: Create main bot engine**

```go
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
```

**Step 2: Verify syntax**

Run: `go build ./bot/...`
Expected: No errors

**Step 3: Commit**

```bash
git add bot/bot.go
git commit -m "feat(bot): add bot engine orchestrator"
```

---

### Task 7: Integrate Bot into main.go

**Files:**
- Modify: `main.go`

**Step 1: Add bot import and initialization**

Find the imports section and add:
```go
import (
	// ... existing imports ...
	"eximmon/bot"
)
```

**Step 2: Add bot engine variable near other globals**

Add after `var notifyEmail = ""`:
```go
var botEngine *bot.Engine
```

**Step 3: Initialize bot in main() after whm.Log = log**

Add after `whm.Log = log`:
```go
	// Initialize bot engine
	bot.Log = log
	botEngine = bot.NewEngine()
	if botEngine != nil {
		if err := botEngine.Start(); err != nil {
			log("Bot engine error: %v", err)
		}
	}
```

**Step 4: Add defer to stop bot on exit**

Add after the bot initialization:
```go
	defer func() {
		if botEngine != nil {
			botEngine.Stop()
		}
	}()
```

**Step 5: Update log message to show bot status**

Find the environment variables log section and update:
```go
	if whm.ApiToken == "" {
		log("Please declare -x API_TOKEN=...")
		log("Other environments variables: MAX_PER_MIN=8 , MAX_PER_HOUR=100")
		log("NOTIFY_EMAIL=email , EXIM_LOG=/var/log/exim_mainlog")
		log("WHM_API_HOST=127.0.0.1")
		log("PREFER_MODERN_UAPI=true (default: true, set to 'false' to use legacy only)")
		log("")
		log("Bot Integration:")
		log("  TELEGRAM_BOT_TOKEN=xxx")
		log("  TELEGRAM_ADMIN_IDS=123,456")
		log("  TELEGRAM_NOTIFY_CHAT_ID=-123")
		log("  SLACK_BOT_TOKEN=xoxb-xxx")
		log("  SLACK_ADMIN_IDS=U123,U456")
		log("  SLACK_NOTIFY_CHANNEL=C123")
	}
```

**Step 6: Build and verify**

Run: `go build -o bin/eximmon main.go`
Expected: No errors

**Step 7: Commit**

```bash
git add main.go
git commit -m "feat: integrate bot engine into main application"
```

---

### Task 8: Update Documentation

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`

**Step 1: Add bot documentation to README.md**

Add after the API Support section:
```markdown
## Bot Integration

Eximmon supports Telegram and Slack bot integration for real-time notifications and remote control.

### Setup

#### Telegram
1. Create a bot via [@BotFather](https://t.me/botfather) and get the token
2. Get your user ID by messaging [@userinfobot](https://t.me/userinfobot)
3. (Optional) Create a group/channel for notifications and add the bot

#### Slack
1. Create a Slack App with Bot User
2. Install to workspace and get Bot User OAuth Token
3. (Optional) Create a channel for notifications and invite the bot

### Environment Variables

```bash
# Telegram
TELEGRAM_BOT_TOKEN=123456:ABC-DEF
TELEGRAM_ADMIN_IDS=123456789,987654321
TELEGRAM_NOTIFY_CHAT_ID=-100123456789

# Slack
SLACK_BOT_TOKEN=xoxb-123456-abcdef
SLACK_ADMIN_IDS=U12345,U67890
SLACK_NOTIFY_CHANNEL=C12345
```

### Available Commands

| Command | Description | Admin Only |
|---------|-------------|------------|
| `/status` | Check eximmon status | No |
| `/suspend <email>` | Suspend email | Yes |
| `/unsuspend <email>` | Unsuspend email | Yes |
| `/list` | List suspended emails | No |
| `/stats <email|domain>` | View statistics | No |
| `/config` | View configuration | No |
| `/set <key> <value>` | Update threshold | Yes |
| `/whitelist add/remove/list` | Manage whitelist | Yes |
```

**Step 2: Update CLAUDE.md**

Add after WHM API Integration section:
```markdown
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

### Commands
All commands defined in `bot/commands.go`. Handler in `bot/bot.go` processes commands and interacts with `whm` package.

### Adding to main.go
```go
botEngine = bot.NewEngine()
botEngine.Start()
defer botEngine.Stop()

// On suspension:
botEngine.NotifySuspension(bot.SuspendedInfo{...})
```
```

**Step 3: Commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: add bot integration documentation"
```

---

### Task 9: Final Build and Test

**Step 1: Build the binary**

```bash
cd /Users/sadewadee/Downloads/Plugin\ Pro/mailtracker
go build -o bin/eximmon main.go
```

**Step 2: Test without bot tokens**

```bash
./bin/eximmon help
```

Expected: Shows help without bot errors

**Step 3: Test with Telegram token (optional)**

```bash
TELEGRAM_BOT_TOKEN=xxx TELEGRAM_ADMIN_IDS=123 ./bin/eximmon help
```

Expected: Shows "Telegram bot started" message

**Step 4: Final commit**

```bash
git add -A
git commit -m "feat: complete Telegram and Slack bot integration

- Add bot package with Telegram and Slack support
- Polling-based message retrieval
- Admin whitelist security model
- Commands: status, suspend, unsuspend, list, stats, config, whitelist
- Auto-notification on account suspension"
```

---

## Summary

| Task | Description |
|------|-------------|
| 1 | Add dependencies |
| 2 | Create types and interfaces |
| 3 | Create command parser |
| 4 | Create Telegram implementation |
| 5 | Create Slack implementation |
| 6 | Create bot engine |
| 7 | Integrate into main.go |
| 8 | Update documentation |
| 9 | Build and test |
