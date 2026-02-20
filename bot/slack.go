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
