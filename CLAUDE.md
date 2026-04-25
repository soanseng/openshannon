# openshannon

Go daemon bridging Telegram to Claude Code or Codex CLI.

## Build

```bash
go build ./... && go vet ./... && go test -race ./...
make build      # build binary to ~/go/bin/openshannon
make install    # install config + systemd service + workspace CLAUDE.md
make start      # enable + start systemd service
make restart    # restart after code changes
make logs       # tail journald logs
```

## Architecture

```
Telegram → Go daemon → selected agent CLI → response → Telegram
```

Packages:
- `internal/config` — YAML config with ${ENV} expansion
- `internal/session` — Session lifecycle, topic mapping, persistence
- `internal/safety` — Dual-layer safety filter (blocklist + protected paths)
- `internal/claude` — CLI executor, stream-json parsing
- `internal/codex` — Codex CLI executor
- `internal/gemini` — Gemini API for image generation
- `internal/router` — Command parsing, session key mapping
- `internal/telegram` — Bot, handler, formatter
- `internal/notify` — ntfy push notifications

## Key Design Decisions

- `-w` flag is `--worktree` in Claude CLI, NOT workdir. We use `cmd.Dir` instead.
- `--output-format stream-json --verbose` for real-time streaming.
- Session model per session key (topic/dm/group) stored in `sessions.json`.
- `/agent claude|codex` switches the selected coding agent per session.
- Codex uses `codex exec --cd <workdir>` and `--add-dir` entries from config.
- Codex prompts currently run as fresh `codex exec` invocations; Codex resume IDs
  are not persisted yet.
- Safety: Go blocklist (Layer 1) + Claude Code settings.json deny list (Layer 2).
- `/shell` uses minimal env (no secrets inheritance), process group kill on timeout.
- Gemini image gen: Claude Haiku enhances prompt → Gemini 3.1 Flash generates.

## Testing

All table-driven. `go test -race ./...` must pass before commit.
Security-critical packages (safety, session) require 90%+ coverage.
