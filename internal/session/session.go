package session

import "time"

// State represents the lifecycle state of a session.
type State string

const (
	StateActive State = "active"
	StateIdle   State = "idle"
	StateClosed State = "closed"
)

// Session holds metadata for a single Claude conversation context
// tied to a Telegram topic, DM, or group.
type Session struct {
	Key           string    `json:"key"`              // "topic:12345" / "dm:67890" / "group:11111"
	ClaudeSession string    `json:"claude_session"`   // claude --resume session ID
	Workdir       string    `json:"workdir"`
	State         State     `json:"state"`
	Label         string    `json:"label"`            // topic name from Telegram
	CreatedAt     time.Time `json:"created_at"`
	LastActiveAt  time.Time `json:"last_active_at"`
}
