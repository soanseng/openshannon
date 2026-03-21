package router

import (
	"fmt"
	"strings"
)

// ParseCommand extracts command name and args from a Telegram message.
// "/new ~/infra" -> ("new", "~/infra")
// "/status"      -> ("status", "")
// "just a prompt" -> ("", "")
func ParseCommand(text string) (cmd, args string) {
	if !strings.HasPrefix(text, "/") {
		return "", ""
	}

	// Strip the leading slash.
	text = text[1:]

	// Split into command and the rest.
	if idx := strings.IndexByte(text, ' '); idx >= 0 {
		return text[:idx], text[idx+1:]
	}
	return text, ""
}

// SessionKey generates a session key based on Telegram chat context.
// Forum topic: "topic:<threadID>"
// DM: "dm:<userID>"
// Plain group/supergroup without topic: "group:<chatID>"
func SessionKey(chatType string, chatID int64, threadID int, isTopic bool, userID int64) string {
	if isTopic && threadID != 0 {
		return fmt.Sprintf("topic:%d", threadID)
	}
	if chatType == "private" {
		return fmt.Sprintf("dm:%d", userID)
	}
	return fmt.Sprintf("group:%d", chatID)
}

// IsAllowed checks if a user ID is in the allowed map.
func IsAllowed(allowed map[int64]bool, userID int64) bool {
	return allowed[userID]
}

// ExpandHome expands ~ to the given home directory.
// "~/infra" with home="/home/scipio" -> "/home/scipio/infra"
// "/abs/path" -> "/abs/path" (unchanged)
func ExpandHome(path, home string) string {
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return home + path[1:]
	}
	return path
}
