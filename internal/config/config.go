package config

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level application configuration.
type Config struct {
	Telegram  TelegramConfig `yaml:"telegram"`
	Claude    ClaudeConfig   `yaml:"claude"`
	Gemini    GeminiConfig   `yaml:"gemini"`
	STT       STTConfig      `yaml:"stt"`
	Streaming StreamConfig   `yaml:"streaming"`
	Safety    SafetyConfig   `yaml:"safety"`
	Notify    NotifyConfig   `yaml:"notify"`
	Storage   StorageConfig  `yaml:"storage"`
}

// GeminiConfig holds Google Gemini API settings.
type GeminiConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// TelegramConfig holds Telegram bot settings.
type TelegramConfig struct {
	Token           string        `yaml:"token"`
	AllowedUsers    []int64       `yaml:"allowed_users"`
	LongPollTimeout time.Duration `yaml:"long_poll_timeout"`
}

// ClaudeConfig holds Claude CLI invocation settings.
type ClaudeConfig struct {
	Binary             string        `yaml:"binary"`
	DefaultWorkdir     string        `yaml:"default_workdir"`
	Flags              []string      `yaml:"flags"`
	SessionIdleTimeout time.Duration `yaml:"session_idle_timeout"`
	DefaultTimeout     time.Duration `yaml:"default_timeout"`
	LongTaskTimeout    time.Duration `yaml:"long_task_timeout"`
	MaxBudgetUSD       float64       `yaml:"max_budget_usd"`
}

// STTConfig holds speech-to-text settings.
type STTConfig struct {
	Backend  string `yaml:"backend"`
	GroqKey  string `yaml:"groq_key"`
	Model    string `yaml:"model"`
	Language string `yaml:"language"`
}

// StreamConfig holds streaming output settings.
type StreamConfig struct {
	Enabled          bool          `yaml:"enabled"`
	MinInterval      time.Duration `yaml:"min_interval"`
	MinChunkSize     int           `yaml:"min_chunk_size"`
	MaxMessageLength int           `yaml:"max_message_length"`
}

// SafetyConfig holds safety guard settings.
type SafetyConfig struct {
	ShellTimeout   time.Duration `yaml:"shell_timeout"`
	BlockedPrompts []string      `yaml:"blocked_prompts"`
	BlockedShell   []string      `yaml:"blocked_shell"`
	ProtectedPaths []string      `yaml:"protected_paths"`
}

// NotifyConfig holds push notification settings.
type NotifyConfig struct {
	Enabled    bool     `yaml:"enabled"`
	NtfyServer string   `yaml:"ntfy_server"`
	NtfyTopic  string   `yaml:"ntfy_topic"`
	NtfyToken  string   `yaml:"ntfy_token"`
	Events     []string `yaml:"events"`
}

// StorageConfig holds local storage settings.
type StorageConfig struct {
	Dir string `yaml:"dir"`
}

// envVarRe matches ${VAR_NAME} patterns for environment variable expansion.
var envVarRe = regexp.MustCompile(`\$\{([^}]+)\}`)

// defaults returns a Config populated with all default values.
func defaults() *Config {
	return &Config{
		Telegram: TelegramConfig{
			LongPollTimeout: 30 * time.Second,
		},
		Claude: ClaudeConfig{
			Binary:             "claude",
			DefaultWorkdir:     "~",
			Flags:              []string{"--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose"},
			SessionIdleTimeout: 30 * time.Minute,
			DefaultTimeout:     5 * time.Minute,
			LongTaskTimeout:    30 * time.Minute,
			MaxBudgetUSD:       10.0,
		},
		STT: STTConfig{
			Backend: "groq",
			Model:   "whisper-large-v3-turbo",
		},
		Streaming: StreamConfig{
			Enabled:          true,
			MinInterval:      1 * time.Second,
			MinChunkSize:     200,
			MaxMessageLength: 4096,
		},
		Safety: SafetyConfig{
			ShellTimeout: 30 * time.Second,
		},
		Notify: NotifyConfig{
			Events: []string{"daemon_start", "daemon_crash", "safety_block", "long_task_complete"},
		},
		Storage: StorageConfig{
			Dir: "~/.config/claude-channels",
		},
	}
}

// expandEnv replaces all ${VAR} references in data with their environment
// variable values. Undefined variables expand to the empty string.
func expandEnv(data []byte) []byte {
	return envVarRe.ReplaceAllFunc(data, func(match []byte) []byte {
		varName := envVarRe.FindSubmatch(match)[1]
		return []byte(os.Getenv(string(varName)))
	})
}

// Load reads configuration from the given YAML file path.
// If path is empty, returns a Config with all default values.
// Environment variables in the form ${VAR} are expanded before parsing.
func Load(path string) (*Config, error) {
	cfg := defaults()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	data = expandEnv(data)

	// First pass: unmarshal into a raw map to detect which top-level
	// sections and fields are explicitly present in the YAML file.
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Second pass: unmarshal into the typed struct for proper parsing.
	var overlay Config
	if err := yaml.Unmarshal(data, &overlay); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	mergeConfig(cfg, &overlay, raw)
	return cfg, nil
}

// mergeConfig applies non-zero values from overlay onto base.
// The raw map is used to detect which fields were explicitly set in YAML,
// which is needed for bool fields where the zero value (false) is meaningful.
func mergeConfig(base, overlay *Config, raw map[string]interface{}) {
	// Telegram
	if overlay.Telegram.Token != "" {
		base.Telegram.Token = overlay.Telegram.Token
	}
	if len(overlay.Telegram.AllowedUsers) > 0 {
		base.Telegram.AllowedUsers = overlay.Telegram.AllowedUsers
	}
	if overlay.Telegram.LongPollTimeout != 0 {
		base.Telegram.LongPollTimeout = overlay.Telegram.LongPollTimeout
	}

	// Claude
	if overlay.Claude.Binary != "" {
		base.Claude.Binary = overlay.Claude.Binary
	}
	if overlay.Claude.DefaultWorkdir != "" {
		base.Claude.DefaultWorkdir = overlay.Claude.DefaultWorkdir
	}
	if len(overlay.Claude.Flags) > 0 {
		base.Claude.Flags = overlay.Claude.Flags
	}
	if overlay.Claude.SessionIdleTimeout != 0 {
		base.Claude.SessionIdleTimeout = overlay.Claude.SessionIdleTimeout
	}
	if overlay.Claude.DefaultTimeout != 0 {
		base.Claude.DefaultTimeout = overlay.Claude.DefaultTimeout
	}
	if overlay.Claude.LongTaskTimeout != 0 {
		base.Claude.LongTaskTimeout = overlay.Claude.LongTaskTimeout
	}
	if overlay.Claude.MaxBudgetUSD != 0 {
		base.Claude.MaxBudgetUSD = overlay.Claude.MaxBudgetUSD
	}

	// STT
	if overlay.STT.Backend != "" {
		base.STT.Backend = overlay.STT.Backend
	}
	if overlay.STT.GroqKey != "" {
		base.STT.GroqKey = overlay.STT.GroqKey
	}
	if overlay.STT.Model != "" {
		base.STT.Model = overlay.STT.Model
	}
	if overlay.STT.Language != "" {
		base.STT.Language = overlay.STT.Language
	}

	// Streaming -- uses raw map to correctly handle bool zero-value
	if streamingRaw, ok := rawSection(raw, "streaming"); ok {
		if _, ok := streamingRaw["enabled"]; ok {
			base.Streaming.Enabled = overlay.Streaming.Enabled
		}
	}
	if overlay.Streaming.MinInterval != 0 {
		base.Streaming.MinInterval = overlay.Streaming.MinInterval
	}
	if overlay.Streaming.MinChunkSize != 0 {
		base.Streaming.MinChunkSize = overlay.Streaming.MinChunkSize
	}
	if overlay.Streaming.MaxMessageLength != 0 {
		base.Streaming.MaxMessageLength = overlay.Streaming.MaxMessageLength
	}

	// Safety
	if overlay.Safety.ShellTimeout != 0 {
		base.Safety.ShellTimeout = overlay.Safety.ShellTimeout
	}
	if len(overlay.Safety.BlockedPrompts) > 0 {
		base.Safety.BlockedPrompts = overlay.Safety.BlockedPrompts
	}
	if len(overlay.Safety.BlockedShell) > 0 {
		base.Safety.BlockedShell = overlay.Safety.BlockedShell
	}
	if len(overlay.Safety.ProtectedPaths) > 0 {
		base.Safety.ProtectedPaths = overlay.Safety.ProtectedPaths
	}

	// Notify -- uses raw map to correctly handle bool zero-value
	if notifyRaw, ok := rawSection(raw, "notify"); ok {
		if _, ok := notifyRaw["enabled"]; ok {
			base.Notify.Enabled = overlay.Notify.Enabled
		}
	}
	if overlay.Notify.NtfyServer != "" {
		base.Notify.NtfyServer = overlay.Notify.NtfyServer
	}
	if overlay.Notify.NtfyTopic != "" {
		base.Notify.NtfyTopic = overlay.Notify.NtfyTopic
	}
	if overlay.Notify.NtfyToken != "" {
		base.Notify.NtfyToken = overlay.Notify.NtfyToken
	}
	if len(overlay.Notify.Events) > 0 {
		base.Notify.Events = overlay.Notify.Events
	}

	// Storage
	if overlay.Storage.Dir != "" {
		base.Storage.Dir = overlay.Storage.Dir
	}
}

// rawSection extracts a named section from the raw YAML map as a
// map[string]interface{}. Returns the section and true if present.
func rawSection(raw map[string]interface{}, key string) (map[string]interface{}, bool) {
	v, ok := raw[key]
	if !ok {
		return nil, false
	}
	m, ok := v.(map[string]interface{})
	return m, ok
}
