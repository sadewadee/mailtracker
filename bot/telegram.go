package bot

import (
	"fmt"
	"strconv"
	"sync"

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
