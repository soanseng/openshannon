// Package notify sends push notifications via ntfy.
package notify

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/scipio/claude-channels/internal/config"
)

// Priority levels for ntfy.
type Priority int

const (
	PriorityLow     Priority = 2
	PriorityDefault Priority = 3
	PriorityHigh    Priority = 4
	PriorityUrgent  Priority = 5
)

// eventPriorities maps known event names to their default priority.
var eventPriorities = map[string]Priority{
	"daemon_start":       PriorityLow,
	"daemon_stop":        PriorityLow,
	"daemon_crash":       PriorityHigh,
	"safety_block":       PriorityDefault,
	"long_task_complete":  PriorityLow,
	"claude_crash":       PriorityHigh,
	"session_corrupt":    PriorityHigh,
	"panic":              PriorityUrgent,
}

// EventPriority returns the default priority for a known event.
// Unknown events return PriorityDefault.
func EventPriority(event string) Priority {
	if p, ok := eventPriorities[event]; ok {
		return p
	}
	return PriorityDefault
}

// Notifier sends push notifications via ntfy.
type Notifier struct {
	cfg    config.NotifyConfig
	events map[string]bool
	client *http.Client
}

// New creates a Notifier from the given configuration.
func New(cfg config.NotifyConfig) *Notifier {
	events := make(map[string]bool, len(cfg.Events))
	for _, e := range cfg.Events {
		events[e] = true
	}
	return &Notifier{
		cfg:    cfg,
		events: events,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send sends a notification if the event type is enabled.
// Returns nil if notifications are disabled or the event is not in the allowed list.
func (n *Notifier) Send(event, message string) error {
	return n.SendWithPriority(event, message, EventPriority(event))
}

// SendWithPriority sends a notification with an explicit priority.
// Returns nil if notifications are disabled or the event is not in the allowed list.
func (n *Notifier) SendWithPriority(event, message string, priority Priority) error {
	if !n.cfg.Enabled {
		return nil
	}
	if !n.events[event] {
		return nil
	}

	url := strings.TrimRight(n.cfg.NtfyServer, "/") + "/" + n.cfg.NtfyTopic

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(message))
	if err != nil {
		return fmt.Errorf("failed to create ntfy request: %w", err)
	}

	req.Header.Set("Title", event)
	req.Header.Set("Priority", strconv.Itoa(int(priority)))
	req.Header.Set("Tags", "robot")

	if n.cfg.NtfyToken != "" {
		req.Header.Set("Authorization", "Bearer "+n.cfg.NtfyToken)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send ntfy notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
	}

	return nil
}
