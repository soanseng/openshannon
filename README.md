<p align="center">
  <img src="shannon.jpg" alt="OpenShannon" width="200">
</p>

<h1 align="center">OpenShannon</h1>

<p align="center">
  <a href="https://github.com/soanseng/openshannon/actions/workflows/deploy.yml"><img src="https://github.com/soanseng/openshannon/actions/workflows/deploy.yml/badge.svg" alt="Deploy"></a>
  <a href="https://github.com/soanseng/openshannon/releases"><img src="https://img.shields.io/github/v/tag/soanseng/openshannon?label=version" alt="Version"></a>
  <a href="https://goreportcard.com/report/github.com/soanseng/openshannon"><img src="https://goreportcard.com/badge/github.com/soanseng/openshannon" alt="Go Report Card"></a>
  <a href="https://github.com/soanseng/openshannon/blob/main/LICENSE"><img src="https://img.shields.io/github/license/soanseng/openshannon" alt="License"></a>
  <a href="https://openshannon.org"><img src="https://img.shields.io/badge/docs-openshannon.org-blue" alt="Docs"></a>
</p>

<p align="center">
  English | <a href="README.zh-TW.md">繁體中文</a>
</p>

<p align="center">
  A Go daemon that bridges Telegram to <a href="https://docs.anthropic.com/en/docs/claude-code">Claude Code</a> or Codex CLI, enabling remote control of a coding agent from your phone. Think of it as your personal coding assistant that you can message from anywhere.
</p>

Each Telegram Forum Topic maps to an isolated agent session with its own working
directory. Claude Code sessions also preserve conversation context with
`--resume`; Codex runs each prompt in the session workdir and configured
`add_dirs`.

## Features

- **Text, voice, photos, files** — send any message type to the selected agent
- **Agent switching** — use `/agent claude` or `/agent codex` per session
- **Streaming responses** — see agent responses via `editMessageText`
- **Session management** — multiple isolated sessions mapped to Forum Topics
- **Safety filter** — dual-layer protection (Go blocklist + Claude Code deny list)
- **Direct shell** — `/shell` command for quick system commands
- **ntfy notifications** — push alerts for daemon events
- **systemd service** — auto-restart, journald logging
- **Single binary** — no runtime dependencies beyond the selected agent CLI

## Prerequisites

- **Go 1.22+**
- **Claude Code CLI** installed and authenticated (`claude --version`)
- (Optional) **Codex CLI** installed and authenticated (`codex --version`)
- **Telegram Bot** — create one via [@BotFather](https://t.me/BotFather)
- **Your Telegram User ID** — get it from [@userinfobot](https://t.me/userinfobot)
- (Optional) **Groq API key** for voice note transcription

## Quick Start

### 1. Create Telegram Bot

1. Open [@BotFather](https://t.me/BotFather) in Telegram
2. Send `/newbot` and follow the prompts
3. Copy the bot token (looks like `7123456789:AAH...`)
4. Send `/setprivacy` → select your bot → `Disable` (so bot can read group messages)
5. (Optional) Send `/setcommands` and paste:
   ```
   new - Create new session
   resume - Resume idle session
   sessions - List all sessions
   clear - Clear Claude context, keep workdir
   kill - Kill session completely
   cd - Change working directory
   status - Daemon status
   cancel - Cancel running command
   shell - Run shell command directly
   long - Run with extended timeout
   help - Show all commands
   ```

### 2. Set Up Forum Group (Recommended)

1. Create a new Telegram Group
2. Go to Group Settings → Topics → Enable
3. Add your bot to the group
4. Make the bot an admin (needed for topic access)
5. Create topics for your projects: "infra", "feedbot", etc.

Each topic becomes an isolated agent session. Claude Code preserves conversation
context across prompts; Codex runs each prompt in the topic workdir.

### 3. Install

```bash
git clone https://github.com/soanseng/openshannon.git ~/infra/openshannon
cd ~/infra/openshannon

# Interactive setup wizard (recommended)
bash install.sh

# Or non-interactive: make setup
```

The wizard guides you through:

1. **Build** — compiles the Go binary
2. **Telegram** — bot token + user ID setup
3. **Gemini** — (optional) API key for `/imagine` image generation
4. **Google Services** — (optional) gog CLI authentication for Gmail, Calendar, Drive, Tasks, Contacts
5. **Config** — writes config files with correct permissions
6. **Workspace** — creates `~/OpenShannon/` with CLAUDE.md and systemd service

Files created:
- `~/.config/openshannon/config.yaml` — bot config (600)
- `~/.config/openshannon/env` — secrets (600)
- `~/OpenShannon/` — default workspace with git
- `~/OpenShannon/CLAUDE.md` — Claude instructions for Telegram use
- `~/.config/systemd/user/openshannon.service` — systemd service

To add Google services later: `make setup-gog`

### 4. Configure

Edit `~/.config/openshannon/config.yaml`:

```yaml
telegram:
  token: "${TELEGRAM_BOT_TOKEN}"
  allowed_users:
    - YOUR_TELEGRAM_USER_ID    # <-- replace this
```

Edit `~/.config/openshannon/env`:

```bash
TELEGRAM_BOT_TOKEN=7123456789:AAHxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
# Optional:
GEMINI_API_KEY=your_google_ai_api_key
GROQ_API_KEY=gsk_xxxxxxxxxxxxx
GOG_KEYRING_PASSWORD=your_gog_keyring_password
GOG_ACCOUNT=your@gmail.com
NTFY_TOPIC=claude-agent
NTFY_TOKEN=tk_xxxxxxxxxxxxx
```

Set permissions:

```bash
chmod 600 ~/.config/openshannon/env
```

### 5. Test Run

```bash
# Run in foreground first to verify
cd ~/infra/openshannon
make run
```

Open Telegram, send `/status` to your bot. You should see a status message with uptime and version.

### 6. Deploy as Service

```bash
# Enable and start systemd user service
make start

# Verify
make status
make logs
```

Enable lingering so the service runs without a login session:

```bash
loginctl enable-linger $(whoami)
```

## Usage

### Basic Interaction

Just send a text message to the bot — it goes straight to the selected agent as a prompt:

```
You: help me find all TODO comments in the codebase
Bot: ⚡ (processing...)
Bot: I found 12 TODO comments across 5 files...
```

### Commands

| Command | Description | Example |
|---|---|---|
| `/new [workdir]` | Create new session | `/new ~/infra` |
| `/resume [id]` | Resume idle session | `/resume` |
| `/sessions` | List all sessions | `/sessions` |
| `/clear` | Reset agent context, keep workdir | `/clear` |
| `/kill [id]` | Kill session completely | `/kill` |
| `/cd <path>` | Change working directory | `/cd ~/apps/feedbot` |
| `/status` | Daemon status and stats | `/status` |
| `/cancel` | Cancel running command | `/cancel` |
| `/shell <cmd>` | Run shell command directly | `/shell git status` |
| `/long <prompt>` | Extended 30m timeout | `/long refactor the entire module` |
| `/agent [name]` | Switch coding agent | `/agent codex` |
| `/model [name]` | Switch model | `/model haiku` |
| `/imagine <desc>` | Generate image (Gemini) | `/imagine a cat in space` |
| `/gog <cmd>` | Google services | `/gog gmail search newer_than:1d` |
| `/help` | Show all commands | `/help` |

### Forum Topics = Sessions

In a Forum-enabled group, each topic is an isolated session:

```
Topic: "infra"           → workdir: ~/infra
Topic: "feedbot"         → workdir: ~/apps/feedbot
Topic: "openshannon"     → workdir: ~/infra/openshannon
```

First message in a new topic auto-creates a session. Use `/cd` to set the workdir.

### Session Lifecycle

```
/clear  → Reset agent context, keep workdir and topic binding
/kill   → Remove everything, topic returns to unbound state
```

### Direct Shell

`/shell` bypasses the selected agent and runs commands directly:

```
You: /shell docker ps
Bot: CONTAINER ID  IMAGE         STATUS
     a1b2c3d4      nginx:latest  Up 2 hours
```

Shell commands are safety-filtered (no `sudo`, `rm -rf`, `git push --force`, etc.) and have a 30-second timeout.

## Safety

Two layers of protection:

**Layer 1 — Go daemon blocklist** (before the selected agent sees the prompt):
- Blocks dangerous patterns: `rm -rf /`, `mkfs`, `dd if=`, `curl | sh`
- Blocks dangerous shell commands: `sudo`, `shutdown`, `git push --force`
- Protects sensitive paths: `/etc/`, `/boot/`, `~/.ssh/authorized_keys`
- `/cd` to protected paths is rejected

**Layer 2 — Claude Code deny list** (your existing `settings.json`):
- Blocks tool executions: `Bash(sudo *)`, `Bash(rm -rf /*)`, etc.

## Google Services (/gog)

Integrates with [gog CLI](https://github.com/AarynSmith/gog) for Google Workspace access. Requires a Google account authenticated via `gog auth add`.

```bash
# Set up (one time)
GOG_KEYRING_PASSWORD="your_password" gog auth add your@gmail.com
```

Add to your `env` file:
```bash
GOG_KEYRING_PASSWORD=your_password
GOG_ACCOUNT=your@gmail.com
```

Then in Telegram:
```
/gog gmail search newer_than:1d          # Recent emails
/gog gmail send --to x@y.com --subject "Hi" --body "Hello"
/gog calendar events                     # Today's calendar
/gog calendar create primary --title "Meeting" --start "2026-03-23 15:00" --end "2026-03-23 16:00"
/gog drive ls                            # List Drive files
/gog tasks lists list                    # List task lists
/gog contacts search "John"             # Search contacts
```

Type `/gog` without arguments for the full command reference.

## Image Generation (/imagine)

Uses Claude to enhance your prompt, then Gemini 3.1 Flash to generate the image:

```
/imagine a cat wearing a space helmet painting the Mona Lisa
```

Requires `GEMINI_API_KEY` in your env file. Get one from [Google AI Studio](https://aistudio.google.com/apikey).

## Model Switching (/model)

Each topic/session can use a different model:

```
/model haiku       # Claude Haiku 4.5 (fast, cheap)
/model sonnet      # Claude Sonnet 4.6 (balanced)
/model opus        # Claude Opus 4.6 (most capable)
/model gemini      # Gemini 2.5 Flash
/model gemini-pro  # Gemini 2.5 Pro
/model default     # Reset to config default
```

## Agent Switching (/agent)

Each topic/session can use Claude Code or Codex CLI:

```
/agent claude   # Use Claude Code
/agent codex    # Use Codex CLI
/agent default  # Reset to Claude Code
```

Codex runs with `codex exec --cd <workdir>`. Configure `codex.default_workdir`
and `codex.add_dirs` to make the full OpenShannon repository available, for
example `~/infra/openshannon`.

## Configuration Reference

See [`config.example.yaml`](config.example.yaml) for all options with comments.

Key settings:

| Setting | Default | Description |
|---|---|---|
| `claude.default_timeout` | `5m` | Max time per Claude invocation |
| `claude.long_task_timeout` | `30m` | Timeout for `/long` commands |
| `claude.max_budget_usd` | `10.0` | Cost cap per invocation |
| `codex.default_workdir` | `~` | Workdir for new Codex sessions |
| `codex.add_dirs` | `[]` | Extra directories Codex can access |
| `safety.shell_timeout` | `30s` | Max time for `/shell` commands |
| `streaming.min_interval` | `1s` | Min time between message edits |
| `streaming.max_message_length` | `4096` | Telegram message length limit |

## ntfy Notifications

Enable push notifications for daemon events:

```yaml
notify:
  enabled: true
  ntfy_server: "https://ntfy.sh"      # or your self-hosted instance
  ntfy_topic: "claude-agent"
  events:
    - daemon_start
    - daemon_crash
    - safety_block
    - long_task_complete
```

## Monitoring

```bash
# Live logs
make logs

# Service status
make status

# In Telegram
/status
```

## Project Structure

```
internal/
├── config/       Config loading (YAML + env vars)
├── session/      Session lifecycle + persistence
├── safety/       Blocklist filter + path validation
├── claude/       Claude CLI executor + streaming
├── codex/        Codex CLI executor
├── router/       Command parsing + session key mapping
├── telegram/     Bot, handler, formatter
└── notify/       ntfy push notifications
```

## Development

```bash
# Run tests
make test

# Run with race detector + coverage
make cover

# Run in foreground
make run

# Build
make build
```

## Legal Notice

OpenShannon is an independent open-source project. It is **not** affiliated with, endorsed by, or sponsored by Anthropic.

- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) is a product of [Anthropic](https://www.anthropic.com). "Claude" and "Anthropic" are trademarks of Anthropic, PBC.
- OpenShannon uses Claude Code's official CLI (`claude -p`) in print mode — a publicly available, Anthropic-provided interface for programmatic use.
- This project does not modify, reverse-engineer, or redistribute Claude Code itself.
- Users must have their own Anthropic account and comply with Anthropic's [Consumer Terms of Service](https://www.anthropic.com/legal/consumer-terms), [Commercial Terms](https://www.anthropic.com/legal/commercial-terms), and [Usage Policy](https://www.anthropic.com/legal/aup).
- The `--dangerously-skip-permissions` flag is an official Claude Code feature used at your own risk.

**Users are responsible for ensuring their use of this tool complies with all applicable Anthropic terms and policies.**

## License

MIT
