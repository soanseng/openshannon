package telegram

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	tele "gopkg.in/telebot.v4"

	"github.com/scipio/claude-channels/internal/claude"
	"github.com/scipio/claude-channels/internal/router"
)

// sessionIDRe validates Claude session IDs (UUID-like or hex strings).
var sessionIDRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,128}$`)

// handleMessage is the main entry point registered with telebot for all text messages.
func (b *Bot) handleMessage(c tele.Context) error {
	msg := c.Message()
	if msg == nil || msg.Sender == nil {
		return nil
	}

	if !router.IsAllowed(b.allowed, msg.Sender.ID) {
		return nil // silent ignore for unauthorised users
	}

	key := router.SessionKey(
		string(msg.Chat.Type),
		msg.Chat.ID,
		msg.ThreadID,
		msg.TopicMessage,
		msg.Sender.ID,
	)

	cmd, args := router.ParseCommand(msg.Text)

	switch cmd {
	case "new":
		return b.handleNew(c, key, args)
	case "resume":
		return b.handleResume(c, key, args)
	case "sessions":
		return b.handleSessions(c)
	case "clear":
		return b.handleClear(c, key)
	case "kill":
		return b.handleKill(c, key, args)
	case "cd":
		return b.handleCd(c, key, args)
	case "status":
		return b.handleStatus(c)
	case "cancel":
		return b.handleCancel(c, key)
	case "shell":
		return b.handleShell(c, key, args)
	case "long":
		return b.handleLong(c, key, args)
	case "model":
		return b.handleModel(c, key, args)
	case "imagine":
		return b.handleImagine(c, key, args)
	case "gog":
		return b.handleGog(c, key, args)
	case "help":
		return b.handleHelp(c)
	default:
		if cmd != "" {
			return c.Reply(fmt.Sprintf("Unknown command: /%s\nUse /help for available commands.", cmd))
		}
		return b.handlePrompt(c, key, msg.Text)
	}
}

// handleNew starts a fresh session (no Claude --resume) in the given workdir.
func (b *Bot) handleNew(c tele.Context, key, args string) error {
	workdir := strings.TrimSpace(args)
	if workdir == "" {
		workdir = b.cfg.Claude.DefaultWorkdir
	}
	workdir = filepath.Clean(router.ExpandHome(workdir, homeDir()))
	// Try to resolve symlinks for extra safety.
	if resolved, err := filepath.EvalSymlinks(workdir); err == nil {
		workdir = resolved
	}

	// Safety check: ensure path is not protected.
	if result := b.filter.CheckPath(workdir); !result.Allowed {
		b.stats.incBlocked()
		slog.Warn("new workdir blocked by safety filter",
			"dir", workdir,
			"reason", result.Reason,
			"rule", result.Rule,
		)
		_ = b.notifier.SendCtx(b.ctx, "safety_block", fmt.Sprintf("new workdir blocked: %s -> %s", workdir, result.Rule))
		return c.Reply(fmt.Sprintf("Blocked: %s", result.Reason))
	}

	// Kill any existing session for this key so Create succeeds.
	_ = b.sessions.Kill(key)

	sess, err := b.sessions.Create(key, workdir)
	if err != nil {
		slog.Error("failed to create session", "key", key, "err", err)
		return c.Reply("Failed to create session.")
	}

	_ = b.sessions.Save()
	return c.Reply(fmt.Sprintf("New session started.\nWorkdir: <code>%s</code>\nKey: <code>%s</code>", EscapeHTML(sess.Workdir), EscapeHTML(sess.Key)), tele.ModeHTML)
}

// handleResume resumes a specific Claude session by ID.
func (b *Bot) handleResume(c tele.Context, key, args string) error {
	claudeSessionID := strings.TrimSpace(args)
	if claudeSessionID == "" {
		return c.Reply("Usage: /resume <session-id>")
	}

	// Validate session ID format to prevent injection.
	if !sessionIDRe.MatchString(claudeSessionID) {
		return c.Reply("Invalid session ID format.")
	}

	_ = b.sessions.GetOrCreate(key, router.ExpandHome(b.cfg.Claude.DefaultWorkdir, homeDir()))
	if err := b.sessions.SetClaudeSession(key, claudeSessionID); err != nil {
		slog.Error("failed to resume session", "key", key, "err", err)
		return c.Reply("Failed to resume session.")
	}
	b.sessions.Touch(key)
	_ = b.sessions.Save()

	return c.Reply(fmt.Sprintf("Resumed Claude session: <code>%s</code>", EscapeHTML(claudeSessionID)), tele.ModeHTML)
}

// handleSessions lists all tracked sessions.
func (b *Bot) handleSessions(c tele.Context) error {
	sessions := b.sessions.List()
	if len(sessions) == 0 {
		return c.Reply("No active sessions.")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Sessions</b> (%d)\n\n", len(sessions)))
	for _, sess := range sessions {
		sb.WriteString(fmt.Sprintf(
			"<b>%s</b>\n  State: %s\n  Workdir: <code>%s</code>\n  Claude: <code>%s</code>\n  Last active: %s\n\n",
			EscapeHTML(sess.Key),
			EscapeHTML(string(sess.State)),
			EscapeHTML(sess.Workdir),
			EscapeHTML(sess.ClaudeSession),
			sess.LastActiveAt.Format(time.DateTime),
		))
	}

	return c.Reply(sb.String(), tele.ModeHTML)
}

// handleClear resets the Claude session ID (fresh conversation) but keeps the workdir.
func (b *Bot) handleClear(c tele.Context, key string) error {
	if err := b.sessions.Clear(key); err != nil {
		slog.Error("failed to clear session", "key", key, "err", err)
		return c.Reply("No session to clear.")
	}
	_ = b.sessions.Save()
	return c.Reply("Session cleared. Next prompt starts a fresh conversation.")
}

// handleKill removes a session entirely.
func (b *Bot) handleKill(c tele.Context, key, args string) error {
	target := strings.TrimSpace(args)
	if target == "" {
		target = key
	}
	if err := b.sessions.Kill(target); err != nil {
		slog.Error("failed to kill session", "key", target, "err", err)
		return c.Reply("Failed to kill session.")
	}
	_ = b.sessions.Save()
	return c.Reply(fmt.Sprintf("Session <code>%s</code> killed.", EscapeHTML(target)), tele.ModeHTML)
}

// handleCd changes the working directory for the current session.
func (b *Bot) handleCd(c tele.Context, key, args string) error {
	dir := strings.TrimSpace(args)
	if dir == "" {
		return c.Reply("Usage: /cd <path>")
	}
	dir = filepath.Clean(router.ExpandHome(dir, homeDir()))
	// Try to resolve symlinks for extra safety.
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}

	// Validate that the directory exists.
	info, err := os.Stat(dir)
	if err != nil {
		return c.Reply(fmt.Sprintf("Directory not found: %s", EscapeHTML(dir)), tele.ModeHTML)
	}
	if !info.IsDir() {
		return c.Reply(fmt.Sprintf("Not a directory: %s", EscapeHTML(dir)), tele.ModeHTML)
	}

	// Safety check: ensure path is not protected.
	if result := b.filter.CheckPath(dir); !result.Allowed {
		b.stats.incBlocked()
		slog.Warn("cd blocked by safety filter",
			"dir", dir,
			"reason", result.Reason,
			"rule", result.Rule,
		)
		_ = b.notifier.SendCtx(b.ctx, "safety_block", fmt.Sprintf("cd blocked: %s -> %s", dir, result.Rule))
		return c.Reply(fmt.Sprintf("Blocked: %s", result.Reason))
	}

	_ = b.sessions.GetOrCreate(key, dir)
	if err := b.sessions.SetWorkdir(key, dir); err != nil {
		slog.Error("failed to set workdir", "key", key, "dir", dir, "err", err)
		return c.Reply("Failed to change directory.")
	}

	_ = b.sessions.Save()
	return c.Reply(fmt.Sprintf("Workdir changed to <code>%s</code>", EscapeHTML(dir)), tele.ModeHTML)
}

// handleStatus shows bot uptime, version, sessions, and stats.
func (b *Bot) handleStatus(c tele.Context) error {
	uptime := time.Since(b.startTime).Round(time.Second)
	sessions := b.sessions.List()
	prompts, shell, blocked, errors := b.stats.snapshot()

	var sb strings.Builder
	sb.WriteString("<b>claude-channels status</b>\n\n")
	sb.WriteString(fmt.Sprintf("Version: <code>%s</code>\n", EscapeHTML(Version)))
	sb.WriteString(fmt.Sprintf("Uptime: %s\n", uptime))
	sb.WriteString(fmt.Sprintf("Bot: @%s\n\n", EscapeHTML(b.bot.Me.Username)))

	sb.WriteString(fmt.Sprintf("<b>Stats</b>\n"))
	sb.WriteString(fmt.Sprintf("  Prompts: %d\n", prompts))
	sb.WriteString(fmt.Sprintf("  Shell commands: %d\n", shell))
	sb.WriteString(fmt.Sprintf("  Blocked: %d\n", blocked))
	sb.WriteString(fmt.Sprintf("  Errors: %d\n\n", errors))

	sb.WriteString(fmt.Sprintf("<b>Sessions</b> (%d)\n", len(sessions)))
	for _, sess := range sessions {
		sb.WriteString(fmt.Sprintf(
			"  <code>%s</code> [%s] %s (last: %s)\n",
			EscapeHTML(sess.Key),
			EscapeHTML(string(sess.State)),
			EscapeHTML(sess.Workdir),
			sess.LastActiveAt.Format("15:04:05"),
		))
	}

	return c.Reply(sb.String(), tele.ModeHTML)
}

// handleCancel terminates the running Claude process for this session, if any.
func (b *Bot) handleCancel(c tele.Context, key string) error {
	if err := b.executor.Cancel(key); err != nil {
		slog.Error("cancel failed", "key", key, "err", err)
		return c.Reply("Cancel failed.")
	}
	return c.Reply("Cancelled running Claude process.")
}

// handleShell executes a shell command in the session's workdir.
func (b *Bot) handleShell(c tele.Context, key, args string) error {
	cmdStr := strings.TrimSpace(args)
	if cmdStr == "" {
		return c.Reply("Usage: /shell <command>")
	}

	// Safety check.
	if result := b.filter.CheckShell(cmdStr); !result.Allowed {
		b.stats.incBlocked()
		slog.Warn("shell command blocked",
			"cmd", cmdStr,
			"reason", result.Reason,
			"rule", result.Rule,
		)
		_ = b.notifier.SendCtx(b.ctx, "safety_block", fmt.Sprintf("shell blocked: %s -> %s", cmdStr, result.Rule))
		return c.Reply(fmt.Sprintf("Blocked: %s", result.Reason))
	}

	sess := b.sessions.GetOrCreate(key, router.ExpandHome(b.cfg.Claude.DefaultWorkdir, homeDir()))
	workdir := router.ExpandHome(sess.Workdir, homeDir())

	ctx, cancel := context.WithTimeout(b.ctx, b.cfg.Safety.ShellTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	cmd.Dir = workdir
	// CRIT-1: Minimal explicit env — do NOT inherit the full process environment.
	cmd.Env = []string{
		"HOME=" + homeDir(),
		"TMPDIR=/tmp",
		"PATH=" + os.Getenv("PATH"),
		"SHELL=/bin/sh",
		"LANG=en_US.UTF-8",
	}
	// CRIT-2: Kill entire process group on timeout to prevent orphaned children.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	b.react(c.Message(), "⚡")
	b.stats.incShell()

	err := cmd.Run()
	b.sessions.Touch(key)

	var sb strings.Builder
	if stdout.Len() > 0 {
		sb.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(stderr.String())
	}

	output := sb.String()
	if err != nil {
		output += fmt.Sprintf("\n\nexit: %s", err)
	}

	if output == "" {
		output = "(no output)"
	}

	// Wrap in <pre> for readability.
	reply := fmt.Sprintf("<pre>%s</pre>", EscapeHTML(strings.TrimSpace(output)))
	if len(reply) > b.cfg.Streaming.MaxMessageLength {
		return b.sendChunked(c.Message().Chat, output, c.Message().ThreadID)
	}
	return c.Reply(reply, tele.ModeHTML)
}

// handleLong sends a prompt to Claude with the long task timeout.
func (b *Bot) handleLong(c tele.Context, key, args string) error {
	prompt := strings.TrimSpace(args)
	if prompt == "" {
		return c.Reply("Usage: /long <prompt>")
	}
	return b.runPrompt(c, key, prompt, b.cfg.Claude.LongTaskTimeout)
}

// handleHelp sends a list of available commands.
func (b *Bot) handleHelp(c tele.Context) error {
	help := `<b>claude-channels commands</b>

<b>Session management</b>
/new [workdir] — Start fresh session
/resume &lt;session-id&gt; — Resume Claude session
/sessions — List all sessions
/clear — Clear Claude session (keep workdir)
/kill [key] — Remove session entirely
/cd &lt;path&gt; — Change working directory
/cancel — Cancel running Claude process

<b>Interaction</b>
/shell &lt;command&gt; — Run shell command
/long &lt;prompt&gt; — Prompt with long timeout (30m)
/imagine &lt;description&gt; — Generate image (Gemini)
/gog &lt;service&gt; &lt;cmd&gt; — Google services (Gmail/Cal/Drive)

<b>Info</b>
/status — Show bot status and stats
/model [haiku|sonnet|opus|gemini] — Switch model
/help — This message

<b>Prompting</b>
Send any text without a / prefix to prompt Claude.`

	return c.Reply(help, tele.ModeHTML)
}

// validModels maps short names to Claude model IDs.
// validModels maps short names to model identifiers.
// Prefix "gemini:" means use Gemini executor; otherwise use Claude CLI.
var validModels = map[string]string{
	"haiku":      "claude-haiku-4-5-20251001",
	"sonnet":     "claude-sonnet-4-6",
	"opus":       "claude-opus-4-6",
	"gemini":     "gemini:gemini-2.5-flash",
	"gemini-pro": "gemini:gemini-2.5-pro",
}

// IsGeminiModel returns true if the model string indicates a Gemini model.
func IsGeminiModel(model string) bool {
	return strings.HasPrefix(model, "gemini:")
}

// GeminiModelID extracts the Gemini model ID from the prefixed string.
func GeminiModelID(model string) string {
	return strings.TrimPrefix(model, "gemini:")
}

// handleModel switches the model for the current session.
func (b *Bot) handleModel(c tele.Context, key, args string) error {
	model := strings.TrimSpace(strings.ToLower(args))

	// No arg — show current model
	if model == "" {
		sess := b.sessions.GetOrCreate(key, router.ExpandHome(b.cfg.Claude.DefaultWorkdir, homeDir()))
		current := sess.Model
		if current == "" {
			current = "default (Claude from config)"
		}
		lines := fmt.Sprintf(`Current model: <b>%s</b>

<b>Claude</b>
/model haiku — Fast, cheap
/model sonnet — Balanced
/model opus — Most capable

<b>Gemini</b>
/model gemini — Gemini 2.5 Flash
/model gemini-pro — Gemini 2.5 Pro

/model default — Reset to config default`, current)
		return c.Reply(lines, tele.ModeHTML)
	}

	// Reset to default
	if model == "default" || model == "reset" {
		sess := b.sessions.GetOrCreate(key, router.ExpandHome(b.cfg.Claude.DefaultWorkdir, homeDir()))
		_ = b.sessions.SetModel(sess.Key, "")
		_ = b.sessions.Save()
		return c.Reply("Model reset to config default (Claude).")
	}

	// Set model
	modelID, ok := validModels[model]
	if !ok {
		return c.Reply(fmt.Sprintf("Unknown model: %s\nUse: haiku, sonnet, opus, gemini, gemini-pro, or default", model))
	}

	// Check Gemini API key is configured
	if IsGeminiModel(modelID) && b.cfg.Gemini.APIKey == "" {
		return c.Reply("Gemini API key not configured. Set gemini.api_key in config.yaml.")
	}

	sess := b.sessions.GetOrCreate(key, router.ExpandHome(b.cfg.Claude.DefaultWorkdir, homeDir()))
	if err := b.sessions.SetModel(sess.Key, modelID); err != nil {
		slog.Error("failed to set model", "key", key, "err", err)
		return c.Reply("Failed to set model.")
	}
	_ = b.sessions.Save()
	return c.Reply(fmt.Sprintf("Model switched to <b>%s</b> (%s)", model, modelID), tele.ModeHTML)
}

// handleGog runs gog CLI commands for Google services (Gmail, Calendar, Drive, etc).
func (b *Bot) handleGog(c tele.Context, key, args string) error {
	subcmd := strings.TrimSpace(args)

	// No args — show usage
	if subcmd == "" {
		help := `<b>/gog — Google Services</b>

<b>Gmail</b>
/gog gmail search newer_than:1d
/gog gmail send --to user@gmail.com --subject "Hi" --body "Hello"
/gog gmail get &lt;messageId&gt;

<b>Calendar</b>
/gog calendar events
/gog calendar create primary --title "Meeting" --start "2026-03-23 15:00" --end "2026-03-23 16:00"
/gog calendar search "meeting"

<b>Drive</b>
/gog drive ls
/gog drive search "report"
/gog drive upload &lt;file&gt;
/gog drive download &lt;fileId&gt;

<b>Tasks</b>
/gog tasks lists list
/gog tasks list &lt;listId&gt;
/gog tasks add &lt;listId&gt; --title "Todo"

<b>Contacts</b>
/gog contacts search "name"
/gog contacts list

<b>Any gog command works:</b>
/gog &lt;service&gt; &lt;command&gt; [flags]`
		return c.Reply(help, tele.ModeHTML)
	}

	msg := c.Message()
	b.react(msg, "📧")

	// Build gog command with account flag
	account := os.Getenv("GOG_ACCOUNT")
	if account == "" {
		account = "your@gmail.com"
	}

	// Inject --account if not already present
	gogArgs := subcmd
	if !strings.Contains(gogArgs, "--account") {
		gogArgs = gogArgs + " --account " + account
	}
	gogArgs += " --plain --no-input"

	// Safety check
	if result := b.filter.CheckShell("gog " + gogArgs); !result.Allowed {
		b.react(msg, "🚫")
		return c.Reply(fmt.Sprintf("Blocked: %s", result.Reason))
	}

	sess := b.sessions.GetOrCreate(key, router.ExpandHome(b.cfg.Claude.DefaultWorkdir, homeDir()))
	workdir := router.ExpandHome(sess.Workdir, homeDir())

	ctx, cancel := context.WithTimeout(b.ctx, b.cfg.Safety.ShellTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", "gog "+gogArgs)
	cmd.Dir = workdir
	cmd.Env = []string{
		"HOME=" + homeDir(),
		"TMPDIR=/tmp",
		"PATH=" + os.Getenv("PATH"),
		"GOG_KEYRING_PASSWORD=" + os.Getenv("GOG_KEYRING_PASSWORD"),
		"LANG=en_US.UTF-8",
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	output, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(output))

	if err != nil && result == "" {
		b.react(msg, "❌")
		slog.Error("gog command failed", "cmd", gogArgs, "err", err)
		return c.Reply("gog command failed. Check server logs.")
	}

	if result == "" {
		result = "(no output)"
	}

	b.react(msg, "✅")

	// Send as pre-formatted for tables
	if len(result) <= b.cfg.Streaming.MaxMessageLength-20 {
		sendErr := c.Reply(fmt.Sprintf("<pre>%s</pre>", EscapeHTML(result)), tele.ModeHTML)
		if sendErr != nil {
			return c.Reply(result)
		}
		return nil
	}
	return b.sendChunked(msg.Chat, result, msg.ThreadID)
}

// handleImagine generates an image using Claude (prompt enhancement) + Gemini (image gen).
func (b *Bot) handleImagine(c tele.Context, key, args string) error {
	prompt := strings.TrimSpace(args)
	if prompt == "" {
		return c.Reply("Usage: /imagine <description>\nExample: /imagine a cat wearing a space helmet")
	}

	if b.gemini == nil {
		return c.Reply("Gemini not configured. Set gemini.api_key in config.yaml.")
	}

	msg := c.Message()
	b.react(msg, "🎨")

	placeholder, err := b.sendMessage(msg.Chat, "Enhancing prompt with Claude...", msg.ThreadID, "")
	if err != nil {
		return fmt.Errorf("failed to send placeholder: %w", err)
	}

	ctx, cancel := context.WithTimeout(b.ctx, 3*time.Minute)
	defer cancel()

	// Step 1: Use Claude to enhance the image prompt
	enhancePrompt := fmt.Sprintf(
		`You are an expert image prompt engineer. Enhance this image description into a detailed, vivid prompt for an AI image generator.
Keep it under 200 words. Output ONLY the enhanced prompt, nothing else.

User request: %s`, prompt)

	enhancedPrompt := prompt // fallback to original if Claude fails
	result, err := b.executor.Run(ctx, key+":imagine", "", "/tmp", enhancePrompt, claude.RunOpts{Model: "claude-haiku-4-5-20251001"})
	if err == nil && result.Text != "" {
		enhancedPrompt = result.Text
		slog.Info("prompt enhanced", "original", prompt, "enhanced", enhancedPrompt)
	} else {
		slog.Warn("prompt enhancement failed, using original", "err", err)
	}

	// Step 2: Generate image with Gemini
	_, _ = b.editMessage(placeholder, fmt.Sprintf("Generating image...\n\n<i>%s</i>", EscapeHTML(enhancedPrompt)), tele.ModeHTML)

	text, imagePath, genErr := b.gemini.GenerateImage(ctx, enhancedPrompt, "")
	if genErr != nil {
		b.react(msg, "❌")
		slog.Error("gemini image generation failed", "err", genErr)
		_, _ = b.editMessage(placeholder, "Image generation failed. Check server logs.", "")
		return nil
	}

	// Delete placeholder
	_ = b.bot.Delete(placeholder)

	// Send image if we got one
	if imagePath != "" {
		photo := &tele.Photo{File: tele.FromDisk(imagePath)}
		caption := prompt
		if text != "" {
			caption = fmt.Sprintf("%s\n\n%s", prompt, text)
		}
		if len(caption) > 1024 {
			caption = caption[:1024]
		}
		photo.Caption = caption
		_, sendErr := b.bot.Send(msg.Chat, photo, &tele.SendOptions{ThreadID: msg.ThreadID})
		if sendErr != nil {
			slog.Error("failed to send generated image", "err", sendErr)
			if text != "" {
				_, _ = b.sendMessage(msg.Chat, text, msg.ThreadID, "")
			}
		}
		_ = os.Remove(imagePath)
		b.react(msg, "✅")
		return nil
	}

	// Text only (no image)
	if text != "" {
		_, _ = b.sendMessage(msg.Chat, text, msg.ThreadID, "")
		b.react(msg, "✅")
		return nil
	}

	_, _ = b.sendMessage(msg.Chat, "(Gemini returned empty response)", msg.ThreadID, "")
	return nil
}

// handlePrompt sends the user's text to Claude and streams the response.
func (b *Bot) handlePrompt(c tele.Context, key, prompt string) error {
	return b.runPrompt(c, key, prompt, b.cfg.Claude.DefaultTimeout)
}

// runPrompt is the shared implementation for handlePrompt and handleLong.
func (b *Bot) runPrompt(c tele.Context, key, prompt string, timeout time.Duration) error {
	// HIGH-3: Prevent concurrent execution for the same key.
	if _, loaded := b.inflight.LoadOrStore(key, struct{}{}); loaded {
		return c.Reply("A command is already running. Use /cancel first.")
	}
	defer b.inflight.Delete(key)

	msg := c.Message()

	// 1. React with eyes to acknowledge receipt.
	b.react(msg, "👀")

	// 2. Get or create session.
	sess := b.sessions.GetOrCreate(key, router.ExpandHome(b.cfg.Claude.DefaultWorkdir, homeDir()))
	workdir := router.ExpandHome(sess.Workdir, homeDir())

	// 3. Safety filter.
	if result := b.filter.CheckPrompt(prompt); !result.Allowed {
		b.stats.incBlocked()
		b.react(msg, "🚫")
		slog.Warn("prompt blocked by safety filter",
			"reason", result.Reason,
			"rule", result.Rule,
		)
		_ = b.notifier.SendCtx(b.ctx, "safety_block", fmt.Sprintf("prompt blocked: %s", result.Rule))
		return c.Reply(fmt.Sprintf("Blocked: %s", result.Reason))
	}

	// 4. React with lightning to signal processing started.
	b.react(msg, "⚡")
	b.stats.incPrompts()

	// 5. Send placeholder.
	placeholder, err := b.sendMessage(msg.Chat, "Thinking...", msg.ThreadID, "")
	if err != nil {
		slog.Error("failed to send placeholder", "err", err)
		b.stats.incErrors()
		return fmt.Errorf("failed to send placeholder: %w", err)
	}

	// 6. Run Claude with streaming.
	ctx, cancel := context.WithTimeout(b.ctx, timeout)
	defer cancel()

	var (
		accumulated strings.Builder
		lastEdit    time.Time
		editMu      sync.Mutex
	)

	streamCb := func(text string) {
		editMu.Lock()
		defer editMu.Unlock()

		accumulated.WriteString(text)

		// Throttle edits: at least MinInterval apart and MinChunkSize accumulated.
		now := time.Now()
		sinceLastEdit := now.Sub(lastEdit)
		accLen := accumulated.Len()

		if sinceLastEdit < b.cfg.Streaming.MinInterval && accLen < b.cfg.Streaming.MinChunkSize {
			return
		}
		if accLen == 0 {
			return
		}

		// Truncate if the accumulated text exceeds max message length.
		preview := accumulated.String()
		if len(preview) > b.cfg.Streaming.MaxMessageLength-20 {
			preview = preview[:b.cfg.Streaming.MaxMessageLength-30] + "\n\n... (streaming)"
		}

		_, editErr := b.editMessage(placeholder, preview, "")
		if editErr != nil {
			slog.Debug("stream edit failed", "err", editErr)
			return
		}
		lastEdit = now
	}

	// Route to Gemini or Claude depending on session model.
	var executor claude.Executor = b.executor
	if IsGeminiModel(sess.Model) && b.gemini != nil {
		executor = b.gemini
	}

	opts := claude.RunOpts{Model: sess.Model}
	result, err := executor.RunWithStream(ctx, key, sess.ClaudeSession, workdir, prompt, opts, streamCb)

	b.sessions.Touch(key)

	// 7. Handle errors.
	if err != nil {
		b.stats.incErrors()
		b.react(msg, "❌")
		slog.Error("claude execution failed", "key", key, "err", err)
		_, _ = b.editMessage(placeholder, "Claude encountered an error. Check server logs for details.", "")
		return nil // don't propagate; user already informed
	}

	// 8. Save session ID from result.
	if result.SessionID != "" {
		_ = b.sessions.SetClaudeSession(key, result.SessionID)
	}

	_ = b.sessions.Save()

	// 9. Final response.
	finalText := result.Text
	if finalText == "" {
		finalText = accumulated.String()
	}

	if finalText == "" {
		finalText = "(empty response)"
	}

	// Try to format as HTML and edit the placeholder.
	html := MarkdownToHTML(finalText)

	if len(html) <= b.cfg.Streaming.MaxMessageLength {
		_, editErr := b.editMessage(placeholder, html, tele.ModeHTML)
		if editErr != nil {
			// Fall back to plain text.
			_, _ = b.editMessage(placeholder, finalText, "")
		}
	} else {
		// Delete placeholder and send as chunks.
		_ = b.bot.Delete(placeholder)
		if chunkErr := b.sendChunked(msg.Chat, finalText, msg.ThreadID); chunkErr != nil {
			b.stats.incErrors()
			slog.Error("failed to send chunked response", "err", chunkErr)
			return c.Reply("Failed to send response (too long).")
		}
	}

	// 10. Success reaction.
	b.react(msg, "✅")

	if result.CostUSD > 0 {
		slog.Info("prompt completed",
			"key", key,
			"cost_usd", result.CostUSD,
			"session_id", result.SessionID,
		)
	}

	// Notify on long tasks.
	if timeout >= b.cfg.Claude.LongTaskTimeout {
		_ = b.notifier.SendCtx(b.ctx, "long_task_complete", fmt.Sprintf("Long task done (cost: $%.4f)", result.CostUSD))
	}

	return nil
}

// Version is set at build time via -ldflags.
var Version = "dev"
