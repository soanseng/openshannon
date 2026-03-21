# Claude Channels — Design Spec

**Date**: 2026-03-22
**Status**: Reviewed (spec review v1 — 3C/5H/7M/5L issues addressed)
**Author**: scipio + Claude

## Overview

A Go daemon that bridges Telegram to Claude Code, enabling remote control of a Claude Code agent from a mobile device. Replaces OpenClaw gateway with a purpose-built solution that leverages Claude Code's native agent capabilities (tools, MCP servers, plugins, hooks).

## Goals

1. Send prompts to Claude Code from Telegram (text, voice, images, files)
2. Manage multiple isolated sessions mapped to Telegram Forum Topics
3. Stream Claude's responses back in real-time via `editMessageText`
4. Provide safety filtering as a defense layer before Claude Code execution
5. Run as a systemd user service with auto-restart and ntfy notifications
6. Clean enough to open-source

## Non-Goals

- Multi-model support (Claude Code only)
- Multi-user SaaS (single-user with whitelist)
- Web dashboard / GUI
- Replacing Claude Code's built-in permission system (we layer on top)

## Architecture

```
Telegram Bot API (Long Polling)
        │
        ▼
┌─ Go Daemon (claude-channels) ──────────────────────┐
│                                                      │
│  Telegram Adapter → Command Router → Safety Filter   │
│                          │                           │
│                    Session Manager                    │
│                     │          │                      │
│              claude -p      /bin/bash                 │
│              --resume       (for /shell)              │
│                     │                                 │
│              Response Formatter → editMessageText     │
│                                                      │
│              ntfy Integration (notifications)         │
└──────────────────────────────────────────────────────┘
```

### Core Components

| Component | Responsibility |
|---|---|
| **Telegram Adapter** | Receive/send messages, handle message types (text/photo/voice/file/sticker), reactions, streaming via editMessageText |
| **Command Router** | Parse `/` commands vs prompts, dispatch to handlers, enforce user whitelist |
| **Safety Filter** | Blocklist regex matching on prompts and /shell commands, protected path detection |
| **Session Manager** | Session lifecycle (create/clear/kill), topic-session mapping, workdir tracking, persistence |
| **Claude Executor** | Spawn `claude -p --resume` with `--output-format stream-json --verbose`, parse streaming events, filter by type, timeout/cancel |
| **Response Formatter** | Markdown→Telegram HTML, long message chunking, code block preservation |
| **ntfy Integration** | Push notifications for daemon events (start, crash, safety block, long task complete) |

## Session Management

### Topic = Session Mapping

Each Telegram Forum Topic maps to an isolated session:

```
Group (Forum mode)
├── Topic: "infra"           → Session A, workdir: ~/infra
├── Topic: "feedbot"         → Session B, workdir: ~/apps/feedbot
├── Topic: "claude-channels" → Session C, workdir: ~/infra/claude-channels
└── Topic: "General"         → Session D, workdir: ~ (default)
```

### Session Key Strategy

| Context | Key Format | Behavior |
|---|---|---|
| Forum Topic | `topic:<threadID>` | Auto-bind, one topic = one session |
| DM | `dm:<userID>` | Manual management via /new, /resume |
| Plain Group | `group:<chatID>` | Shared session for entire group |

### Session Lifecycle

```
  /new [workdir]          prompt (auto-create)     /kill
      │                       │                      │
      ▼                       ▼                      ▼
  CREATED ──────────────→ ACTIVE ──────────────→ CLOSED
                            │  ▲
                  idle 30m  │  │ /resume
                            ▼  │
                          IDLE
```

### Session Model

```go
type Session struct {
    Key           string    // "topic:12345" / "dm:67890" / "group:11111"
    ClaudeSession string    // claude --resume session ID
    Workdir       string
    State         State     // ACTIVE / IDLE / CLOSED
    Label         string    // topic name from Telegram
    CreatedAt     time.Time
    LastActiveAt  time.Time
}
```

### /clear vs /kill

| Command | Claude context | Workdir | Topic binding |
|---|---|---|---|
| `/clear` | Reset | Preserved | Preserved |
| `/kill` | Reset | Removed | Removed |

### Persistence

Sessions stored in `~/.config/claude-channels/sessions.json`. Loaded at startup, saved on change and graceful shutdown.

### Concurrency

- Each topic has its own queue (topics don't block each other)
- Within a topic: one Claude command at a time, additional messages queued (max 3)
- User notified: `⏳ Queued (1 ahead)`
- Queue overflow (4th+ message): reply `⚠️ Queue full (3 pending). Try again shortly.` — message not enqueued

## Telegram Commands

| Command | Description |
|---|---|
| `/new [workdir]` | Create new session (optional workdir) |
| `/resume [id]` | Resume an idle session |
| `/sessions` | List all sessions |
| `/clear` | Clear Claude context, keep workdir |
| `/kill [id]` | Kill session completely |
| `/cd <path>` | Change current session workdir (protected paths rejected) |
| `/status` | Daemon status, sessions, stats |
| `/cancel` | Cancel running Claude command |
| `/shell <cmd>` | Run shell command directly (bypass Claude) |
| `/long <prompt>` | Run prompt with extended 30m timeout |
| `/help` | List all commands |
| (plain text) | Send as prompt to current Claude session |
| (photo) | Download + send path to Claude |
| (voice) | Groq Whisper STT → prompt |
| (file) | Download to workdir/incoming/ + notify Claude |

## Telegram Adapter

### Inbound Message Handling

| Type | Processing |
|---|---|
| Text | Direct prompt |
| Photo | Download to /tmp/claude-channels/<msgid>.jpg, pass path to Claude. Cleanup: delete after Claude processes, sweep files >1h old on startup. |
| Document | Download to workdir/incoming/<filename>, notify Claude |
| Voice/Audio | Download .ogg → Groq Whisper API → text prompt |
| Sticker | P0: reply "Sticker received: {emoji}" as text prompt. P2: vision recognition |
| Reply | Prepend quoted text as context |

### STT Configuration

```yaml
stt:
  backend: groq           # "groq" | "local"
  groq_key: ${GROQ_API_KEY}
  model: whisper-large-v3-turbo
  language: ""             # auto detect
```

Fallback: if Groq fails, return error with original .ogg attached.

### Streaming Response

Uses `--output-format stream-json --verbose`. The stream outputs NDJSON lines with different `type` fields. We filter for `type: "assistant"` content events and ignore system/hook events.

**Session ID capture**: First invocation runs without `--resume`; parse `session_id` from the JSON result event. Store it in the Session. Subsequent invocations use `--resume <session_id>`. If JSON parsing fails, create a new session on next prompt.

1. Send placeholder message "⏳ Thinking..."
2. Parse stream-json lines, accumulate `assistant` text content
3. Every ~200 chars or ~1 second → `editMessageText` to update
4. On `result` event → final edit + extract `session_id`
5. React ✅ on user's original message

Rate limiting: min 1 second between edits to respect Telegram API limits.
All messages sent with `parse_mode: "HTML"`.

### Long Message Chunking

- Max 4096 chars per Telegram message
- Split priority: paragraph break (`\n\n`) > line break (`\n`) > hard cut
- Code blocks preserved: never split inside ``` fences
- 100ms delay between chunks to prevent ordering issues

### Reaction State Machine

```
Message received  → 👀 (acknowledged)
Safety blocked    → 🚫
Claude running    → ⚡ (processing)
Claude done       → ✅
Claude error      → ❌
```

### Markdown → Telegram HTML

```
**bold**       → <b>bold</b>
*italic*       → <i>italic</i>
`code`         → <code>code</code>
```code```     → <pre>code</pre>
[text](url)    → <a href="url">text</a>
# heading      → <b>heading</b>
- list item    → • list item
| table |      → <pre> formatted </pre>
```

## Safety Filter

### Dual-Layer Protection

**Layer 1 — Go daemon blocklist** (before Claude sees the prompt):
- Regex pattern matching on prompt text
- Separate blocklist for /shell commands
- Protected path detection

**Layer 2 — Claude Code deny list** (existing settings.json):
- `Bash(sudo *)`, `Bash(rm -rf /*)`, etc.

### Blocked Prompt Patterns

```yaml
blocked_prompts:
  - "(?i)rm\\s+-rf\\s+[/~]"
  - "(?i)mkfs"
  - "(?i)dd\\s+if="
  - "(?i)curl.*\\|\\s*sh"
  - "(?i)wget.*\\|\\s*sh"
```

### Blocked Shell Patterns

```yaml
blocked_shell:
  - "(?i)^sudo"
  - "(?i)^su\\s"
  - "(?i)shutdown|reboot"
  - "(?i)git\\s+push\\s+--force"
  - "(?i)git\\s+reset\\s+--hard"
  - "(?i)docker\\s+rm\\s+-f"
```

### Protected Paths

```yaml
protected_paths:
  - "/etc/"
  - "/boot/"
  - "/sys/"
  - "~/.ssh/authorized_keys"
  - "~/.claude/settings.json"
```

### Shell Timeout

`/shell` commands have a 30-second timeout. Exceeded → process killed + `⏰ Command timed out`.

## Configuration

### Config File

`~/.config/claude-channels/config.yaml` — all settings with `${ENV_VAR}` expansion for secrets.

### Environment File

`~/.config/claude-channels/env` — secrets only:
- `TELEGRAM_BOT_TOKEN`
- `GROQ_API_KEY`
- `NTFY_TOPIC`
- `NTFY_TOKEN`

### Claude Code Flags

```yaml
claude:
  binary: claude
  default_workdir: "~"
  flags:
    - "--dangerously-skip-permissions"
    - "--output-format"
    - "stream-json"
    - "--verbose"
  session_idle_timeout: 30m
  default_timeout: 5m
  long_task_timeout: 30m
  max_budget_usd: 10.0         # per-invocation cost cap
```

**Note on `--dangerously-skip-permissions`**: User explicitly chose this (A+B dual layer). Layer 1 (Go daemon blocklist) filters prompts before Claude sees them. Layer 2 (Claude Code `settings.json` deny list) blocks dangerous tool executions at runtime. The regex blocklist cannot inspect what Claude decides to execute — it only filters user input. This is accepted risk for single-user self-hosted use.

## Deployment

### systemd User Service

```ini
[Unit]
Description=Claude Channels Telegram Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/home/scipio/go/bin/claude-channels --config /home/scipio/.config/claude-channels/config.yaml
WorkingDirectory=/home/scipio
Restart=on-failure
RestartSec=10
EnvironmentFile=/home/scipio/.config/claude-channels/env
Environment="HOME=/home/scipio"
Environment=TMPDIR=/tmp
Environment="PATH=/home/scipio/.local/bin:/home/scipio/.cargo/bin:/home/scipio/.local/share/bob/nvim-bin:/home/scipio/.local/share/mise/shims:/home/scipio/go/bin:/home/scipio/.bun/bin:/home/scipio/.local/share/pnpm:/usr/local/bin:/usr/bin:/bin"
Environment="GOPATH=/home/scipio/go"
NoNewPrivileges=true
ReadWritePaths=/home/scipio
StandardOutput=journal
StandardError=journal
SyslogIdentifier=claude-channels

[Install]
WantedBy=default.target
```

**Note**: `ProtectSystem=strict` removed — Claude Code subprocesses need read access to `/usr/share/`, `/proc/`, etc. Security is handled by the daemon's safety filter + Claude Code's deny list instead.

### Dotfiles Integration

```
~/dotfiles/claude-channels.symlink/
├── config.yaml              → ~/.config/claude-channels/config.yaml
├── claude-channels.service  → ~/.config/systemd/user/claude-channels.service
└── env.example              # template only, secrets not in dotfiles
```

### Post-Migration Cleanup

After 24h stable operation:
1. `systemctl --user disable --now openclaw-gateway.service`
2. `systemctl --user disable --now clawdbot-gateway.service`
3. Remove old service files
4. Optional: uninstall openclaw/clawdbot npm packages

## Error Handling

### Error Categories

| Category | Strategy |
|---|---|
| Telegram 429 | Exponential backoff, max 3 retries |
| Telegram 401/403 | Log fatal + ntfy alert, do not retry (token invalid, needs manual fix) |
| Telegram network | Retry, don't crash |
| Claude exit != 0 | Capture stderr, reply to user |
| Claude timeout | Kill process, notify user |
| Claude rate limit | Relay wait time to user |
| STT failure | Reply error + attach original .ogg |
| /shell timeout | Kill after 30s |
| sessions.json corrupt | Backup corrupt file, reset, ntfy alert |
| Panic | Recover, log stack, ntfy, continue |

### ntfy Notification Events

| Event | Priority |
|---|---|
| `daemon_start` | Low |
| `daemon_crash` | High |
| `safety_block` | Medium |
| `long_task_complete` | Low |
| `claude_crash` | High |
| `session_corrupt` | High |
| `panic` | Critical |

### Graceful Shutdown

On SIGINT/SIGTERM:
0. Stop accepting new Telegram updates, drain in-flight updates
1. Save all sessions
2. Cancel running Claude processes
3. Send ntfy notification
4. Exit cleanly

## Monitoring

### Structured Logging

`slog` (Go stdlib) with JSON output to journald.

Fields: session key, workdir, user_id, duration, exit_code.

### /status Command

Displays: uptime, version (embedded via `go build -ldflags "-X main.version=..."`), Claude version, session list, daily stats (prompts, /shell, blocked, errors, cost_usd).

## Testing Strategy

### Unit Tests (main coverage)

- `internal/safety/` — 95%+ (security-critical)
- `internal/session/` — 90%+ (state management)
- `internal/router/` — 85%+
- `internal/telegram/formatter` — 80%+
- `internal/claude/` — 60%+ (relies on integration tests)

All table-driven. `go test -race ./...`.

### Integration Tests

Interface-based mocking: `Executor` interface allows swapping real Claude CLI for mock in tests.

### E2E Smoke Test

Shell script that sends `/status` via Telegram API and verifies response. Run post-deploy.

## Project Structure

```
~/infra/claude-channels/
├── cmd/claude-channels/main.go
├── internal/
│   ├── telegram/
│   │   ├── bot.go
│   │   ├── handler.go
│   │   └── formatter.go
│   ├── router/router.go
│   ├── session/
│   │   ├── manager.go
│   │   └── session.go
│   ├── claude/executor.go
│   ├── safety/filter.go
│   └── config/config.go
├── config.example.yaml
├── claude-channels.service
├── Makefile
├── go.mod
└── go.sum
```

## Implementation Phases

1. **P0 Core** — config, session manager (with basic topic-session mapping), safety filter (including `/cd` path validation), command router, claude executor (stream-json parsing), telegram adapter (text + basic topic support), streaming, systemd service
2. **P1 Media** — photo/file upload, voice STT (Groq), sticker handling, inline keyboards, reply threading, temp file cleanup
3. **P2 Polish** — scheduled tasks, exec approval buttons, /status stats with cost tracking, ntfy integration, dotfiles symlink setup, advanced topic features (workdir inference from topic name)
4. **P3 Cleanup** — OpenClaw/Clawdbot retirement, documentation, open-source prep

**Note on IDLE state**: IDLE means "no messages for `session_idle_timeout`". Implementation: a periodic sweep goroutine (every 5 minutes) checks `LastActiveAt` and transitions ACTIVE → IDLE. IDLE sessions remain resumable via `/resume`. No Claude process to kill — `claude -p` exits after each invocation.
