package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") returned error: %v", err)
	}

	// Claude defaults
	if cfg.Claude.Binary != "claude" {
		t.Errorf("Claude.Binary = %q, want %q", cfg.Claude.Binary, "claude")
	}
	if cfg.Claude.DefaultWorkdir != "~" {
		t.Errorf("Claude.DefaultWorkdir = %q, want %q", cfg.Claude.DefaultWorkdir, "~")
	}
	wantFlags := []string{"--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose"}
	if len(cfg.Claude.Flags) != len(wantFlags) {
		t.Fatalf("Claude.Flags length = %d, want %d", len(cfg.Claude.Flags), len(wantFlags))
	}
	for i, f := range wantFlags {
		if cfg.Claude.Flags[i] != f {
			t.Errorf("Claude.Flags[%d] = %q, want %q", i, cfg.Claude.Flags[i], f)
		}
	}
	if cfg.Claude.SessionIdleTimeout != 30*time.Minute {
		t.Errorf("Claude.SessionIdleTimeout = %v, want %v", cfg.Claude.SessionIdleTimeout, 30*time.Minute)
	}
	if cfg.Claude.DefaultTimeout != 5*time.Minute {
		t.Errorf("Claude.DefaultTimeout = %v, want %v", cfg.Claude.DefaultTimeout, 5*time.Minute)
	}
	if cfg.Claude.LongTaskTimeout != 30*time.Minute {
		t.Errorf("Claude.LongTaskTimeout = %v, want %v", cfg.Claude.LongTaskTimeout, 30*time.Minute)
	}
	if cfg.Claude.MaxBudgetUSD != 10.0 {
		t.Errorf("Claude.MaxBudgetUSD = %f, want %f", cfg.Claude.MaxBudgetUSD, 10.0)
	}

	// STT defaults
	if cfg.STT.Backend != "groq" {
		t.Errorf("STT.Backend = %q, want %q", cfg.STT.Backend, "groq")
	}
	if cfg.STT.Model != "whisper-large-v3-turbo" {
		t.Errorf("STT.Model = %q, want %q", cfg.STT.Model, "whisper-large-v3-turbo")
	}

	// Streaming defaults
	if !cfg.Streaming.Enabled {
		t.Error("Streaming.Enabled = false, want true")
	}
	if cfg.Streaming.MinInterval != 1*time.Second {
		t.Errorf("Streaming.MinInterval = %v, want %v", cfg.Streaming.MinInterval, 1*time.Second)
	}
	if cfg.Streaming.MinChunkSize != 200 {
		t.Errorf("Streaming.MinChunkSize = %d, want %d", cfg.Streaming.MinChunkSize, 200)
	}
	if cfg.Streaming.MaxMessageLength != 4096 {
		t.Errorf("Streaming.MaxMessageLength = %d, want %d", cfg.Streaming.MaxMessageLength, 4096)
	}

	// Safety defaults
	if cfg.Safety.ShellTimeout != 30*time.Second {
		t.Errorf("Safety.ShellTimeout = %v, want %v", cfg.Safety.ShellTimeout, 30*time.Second)
	}

	// Notify defaults
	wantEvents := []string{"daemon_start", "daemon_crash", "safety_block", "long_task_complete"}
	if len(cfg.Notify.Events) != len(wantEvents) {
		t.Fatalf("Notify.Events length = %d, want %d", len(cfg.Notify.Events), len(wantEvents))
	}
	for i, e := range wantEvents {
		if cfg.Notify.Events[i] != e {
			t.Errorf("Notify.Events[%d] = %q, want %q", i, cfg.Notify.Events[i], e)
		}
	}

	// Storage defaults
	if cfg.Storage.Dir != "~/.config/claude-channels" {
		t.Errorf("Storage.Dir = %q, want %q", cfg.Storage.Dir, "~/.config/claude-channels")
	}

	// Telegram defaults
	if cfg.Telegram.LongPollTimeout != 30*time.Second {
		t.Errorf("Telegram.LongPollTimeout = %v, want %v", cfg.Telegram.LongPollTimeout, 30*time.Second)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	yamlContent := `
telegram:
  token: "test-token-123"
  allowed_users: [111, 222]
  long_poll_timeout: 60s
claude:
  binary: "/usr/local/bin/claude"
  default_workdir: "/tmp/work"
  session_idle_timeout: 15m
  max_budget_usd: 25.5
stt:
  backend: "openai"
  model: "whisper-1"
  language: "en"
streaming:
  enabled: false
  min_interval: 2s
  min_chunk_size: 500
  max_message_length: 8192
safety:
  shell_timeout: 10s
  blocked_prompts: ["ignore all", "system prompt"]
  blocked_shell: ["rm -rf /", "curl | bash"]
  protected_paths: ["/etc", "/root"]
notify:
  enabled: true
  ntfy_server: "https://ntfy.example.com"
  ntfy_topic: "claude-bot"
  ntfy_token: "tk_secret"
  events: ["daemon_start"]
storage:
  dir: "/var/lib/claude-channels"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load(%q) returned error: %v", path, err)
	}

	// Telegram overrides
	if cfg.Telegram.Token != "test-token-123" {
		t.Errorf("Telegram.Token = %q, want %q", cfg.Telegram.Token, "test-token-123")
	}
	if len(cfg.Telegram.AllowedUsers) != 2 || cfg.Telegram.AllowedUsers[0] != 111 || cfg.Telegram.AllowedUsers[1] != 222 {
		t.Errorf("Telegram.AllowedUsers = %v, want [111 222]", cfg.Telegram.AllowedUsers)
	}
	if cfg.Telegram.LongPollTimeout != 60*time.Second {
		t.Errorf("Telegram.LongPollTimeout = %v, want %v", cfg.Telegram.LongPollTimeout, 60*time.Second)
	}

	// Claude overrides
	if cfg.Claude.Binary != "/usr/local/bin/claude" {
		t.Errorf("Claude.Binary = %q, want %q", cfg.Claude.Binary, "/usr/local/bin/claude")
	}
	if cfg.Claude.DefaultWorkdir != "/tmp/work" {
		t.Errorf("Claude.DefaultWorkdir = %q, want %q", cfg.Claude.DefaultWorkdir, "/tmp/work")
	}
	if cfg.Claude.SessionIdleTimeout != 15*time.Minute {
		t.Errorf("Claude.SessionIdleTimeout = %v, want %v", cfg.Claude.SessionIdleTimeout, 15*time.Minute)
	}
	if cfg.Claude.MaxBudgetUSD != 25.5 {
		t.Errorf("Claude.MaxBudgetUSD = %f, want %f", cfg.Claude.MaxBudgetUSD, 25.5)
	}

	// Defaults should still apply for fields not in YAML
	if cfg.Claude.DefaultTimeout != 5*time.Minute {
		t.Errorf("Claude.DefaultTimeout = %v, want %v (default)", cfg.Claude.DefaultTimeout, 5*time.Minute)
	}
	if cfg.Claude.LongTaskTimeout != 30*time.Minute {
		t.Errorf("Claude.LongTaskTimeout = %v, want %v (default)", cfg.Claude.LongTaskTimeout, 30*time.Minute)
	}
	wantFlags := []string{"--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose"}
	if len(cfg.Claude.Flags) != len(wantFlags) {
		t.Fatalf("Claude.Flags length = %d, want %d (default)", len(cfg.Claude.Flags), len(wantFlags))
	}
	for i, f := range wantFlags {
		if cfg.Claude.Flags[i] != f {
			t.Errorf("Claude.Flags[%d] = %q, want %q (default)", i, cfg.Claude.Flags[i], f)
		}
	}

	// STT overrides
	if cfg.STT.Backend != "openai" {
		t.Errorf("STT.Backend = %q, want %q", cfg.STT.Backend, "openai")
	}
	if cfg.STT.Model != "whisper-1" {
		t.Errorf("STT.Model = %q, want %q", cfg.STT.Model, "whisper-1")
	}
	if cfg.STT.Language != "en" {
		t.Errorf("STT.Language = %q, want %q", cfg.STT.Language, "en")
	}

	// Streaming overrides
	if cfg.Streaming.Enabled {
		t.Error("Streaming.Enabled = true, want false")
	}
	if cfg.Streaming.MinInterval != 2*time.Second {
		t.Errorf("Streaming.MinInterval = %v, want %v", cfg.Streaming.MinInterval, 2*time.Second)
	}
	if cfg.Streaming.MinChunkSize != 500 {
		t.Errorf("Streaming.MinChunkSize = %d, want %d", cfg.Streaming.MinChunkSize, 500)
	}
	if cfg.Streaming.MaxMessageLength != 8192 {
		t.Errorf("Streaming.MaxMessageLength = %d, want %d", cfg.Streaming.MaxMessageLength, 8192)
	}

	// Safety overrides
	if cfg.Safety.ShellTimeout != 10*time.Second {
		t.Errorf("Safety.ShellTimeout = %v, want %v", cfg.Safety.ShellTimeout, 10*time.Second)
	}
	if len(cfg.Safety.BlockedPrompts) != 2 {
		t.Errorf("Safety.BlockedPrompts length = %d, want 2", len(cfg.Safety.BlockedPrompts))
	}
	if len(cfg.Safety.BlockedShell) != 2 {
		t.Errorf("Safety.BlockedShell length = %d, want 2", len(cfg.Safety.BlockedShell))
	}
	if len(cfg.Safety.ProtectedPaths) != 2 {
		t.Errorf("Safety.ProtectedPaths length = %d, want 2", len(cfg.Safety.ProtectedPaths))
	}

	// Notify overrides
	if !cfg.Notify.Enabled {
		t.Error("Notify.Enabled = false, want true")
	}
	if cfg.Notify.NtfyServer != "https://ntfy.example.com" {
		t.Errorf("Notify.NtfyServer = %q, want %q", cfg.Notify.NtfyServer, "https://ntfy.example.com")
	}
	if cfg.Notify.NtfyTopic != "claude-bot" {
		t.Errorf("Notify.NtfyTopic = %q, want %q", cfg.Notify.NtfyTopic, "claude-bot")
	}
	if cfg.Notify.NtfyToken != "tk_secret" {
		t.Errorf("Notify.NtfyToken = %q, want %q", cfg.Notify.NtfyToken, "tk_secret")
	}
	if len(cfg.Notify.Events) != 1 || cfg.Notify.Events[0] != "daemon_start" {
		t.Errorf("Notify.Events = %v, want [daemon_start]", cfg.Notify.Events)
	}

	// Storage override
	if cfg.Storage.Dir != "/var/lib/claude-channels" {
		t.Errorf("Storage.Dir = %q, want %q", cfg.Storage.Dir, "/var/lib/claude-channels")
	}
}

func TestLoadConfig_EnvExpansion(t *testing.T) {
	const envKey = "TEST_BOT_TOKEN"
	const envVal = "bot123456:ABC-DEF"
	t.Setenv(envKey, envVal)

	yamlContent := `
telegram:
  token: "${TEST_BOT_TOKEN}"
stt:
  groq_key: "${TEST_GROQ_KEY_UNSET}"
notify:
  ntfy_token: "${TEST_BOT_TOKEN}"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load(%q) returned error: %v", path, err)
	}

	if cfg.Telegram.Token != envVal {
		t.Errorf("Telegram.Token = %q, want %q (expanded from $%s)", cfg.Telegram.Token, envVal, envKey)
	}

	// Unset env var should expand to empty string
	if cfg.STT.GroqKey != "" {
		t.Errorf("STT.GroqKey = %q, want \"\" (unset env var)", cfg.STT.GroqKey)
	}

	if cfg.Notify.NtfyToken != envVal {
		t.Errorf("Notify.NtfyToken = %q, want %q (expanded from $%s)", cfg.Notify.NtfyToken, envVal, envKey)
	}
}
