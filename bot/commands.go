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
