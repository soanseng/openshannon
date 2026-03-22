package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	tele "gopkg.in/telebot.v4"

	"github.com/soanseng/openshannon/internal/claude"
	"github.com/soanseng/openshannon/internal/config"
	"github.com/soanseng/openshannon/internal/gemini"
	"github.com/soanseng/openshannon/internal/notify"
	"github.com/soanseng/openshannon/internal/safety"
	"github.com/soanseng/openshannon/internal/session"
)

// Bot wraps a telebot instance with application-level dependencies.
type Bot struct {
	bot       *tele.Bot
	cfg       *config.Config
	sessions  *session.Manager
	executor  claude.Executor
	gemini    *gemini.Executor // optional, for image generation and Gemini models
	filter    *safety.Filter
	notifier  *notify.Notifier
	allowed   map[int64]bool
	startTime time.Time
	stats     *Stats
	ctx       context.Context
	inflight  sync.Map // tracks in-flight prompt executions per key
}

// Stats tracks aggregate bot usage counters.
type Stats struct {
	mu            sync.Mutex
	Prompts       int
	ShellCommands int
	Blocked       int
	Errors        int
}

func (s *Stats) incPrompts() {
	s.mu.Lock()
	s.Prompts++
	s.mu.Unlock()
}

func (s *Stats) incShell() {
	s.mu.Lock()
	s.ShellCommands++
	s.mu.Unlock()
}

func (s *Stats) incBlocked() {
	s.mu.Lock()
	s.Blocked++
	s.mu.Unlock()
}

func (s *Stats) incErrors() {
	s.mu.Lock()
	s.Errors++
	s.mu.Unlock()
}

func (s *Stats) snapshot() (prompts, shell, blocked, errors int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Prompts, s.ShellCommands, s.Blocked, s.Errors
}

// NewBot creates a Bot ready to start polling Telegram for updates.
func NewBot(cfg *config.Config, sessions *session.Manager, executor claude.Executor, filter *safety.Filter, notifier *notify.Notifier) (*Bot, error) {
	pref := tele.Settings{
		Token:  cfg.Telegram.Token,
		Poller: &tele.LongPoller{Timeout: cfg.Telegram.LongPollTimeout},
		OnError: func(err error, c tele.Context) {
			slog.Error("telebot error", "err", err)
		},
	}

	teleBot, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	allowed := make(map[int64]bool, len(cfg.Telegram.AllowedUsers))
	for _, uid := range cfg.Telegram.AllowedUsers {
		allowed[uid] = true
	}

	b := &Bot{
		bot:       teleBot,
		cfg:       cfg,
		sessions:  sessions,
		executor:  executor,
		filter:    filter,
		notifier:  notifier,
		allowed:   allowed,
		startTime: time.Now(),
		stats:     &Stats{},
	}

	// Initialize Gemini executor if API key is configured.
	if cfg.Gemini.APIKey != "" {
		b.gemini = gemini.NewExecutor(cfg.Gemini.APIKey, cfg.Gemini.Model)
		slog.Info("gemini executor initialized", "model", cfg.Gemini.Model)
	}

	teleBot.Handle(tele.OnText, b.handleMessage)

	// Register command menu with Telegram (the "/" autocomplete list).
	if err := teleBot.SetCommands([]tele.Command{
		{Text: "new", Description: "Create new session [workdir]"},
		{Text: "resume", Description: "Resume idle session [id]"},
		{Text: "sessions", Description: "List all sessions"},
		{Text: "clear", Description: "Clear Claude context, keep workdir"},
		{Text: "kill", Description: "Kill session completely [id]"},
		{Text: "cd", Description: "Change working directory"},
		{Text: "status", Description: "Daemon status and stats"},
		{Text: "cancel", Description: "Cancel running command"},
		{Text: "shell", Description: "Run shell command directly"},
		{Text: "long", Description: "Run with extended 30m timeout"},
		{Text: "gog", Description: "Google services (Gmail/Calendar/Drive)"},
		{Text: "imagine", Description: "Generate image with Gemini"},
		{Text: "model", Description: "Switch model (haiku/sonnet/opus/gemini)"},
		{Text: "help", Description: "Show all commands"},
	}); err != nil {
		slog.Warn("failed to register bot commands menu", "err", err)
	}

	return b, nil
}

// Start begins long-polling and blocks until ctx is cancelled.
// On context cancellation it performs a graceful shutdown: stops the
// bot poller and persists sessions.
func (b *Bot) Start(ctx context.Context) error {
	b.ctx = ctx

	slog.Info("bot starting",
		"username", b.bot.Me.Username,
		"allowed_users", len(b.allowed),
	)

	_ = b.notifier.SendCtx(ctx, "daemon_start", fmt.Sprintf("openshannon started as @%s", b.bot.Me.Username))

	// Run the blocking poller in a separate goroutine so we can
	// select on ctx.Done().
	done := make(chan struct{})
	go func() {
		b.bot.Start()
		close(done)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutting down bot")
		b.bot.Stop()
		<-done
	case <-done:
		// Poller stopped on its own (unlikely in normal operation).
	}

	if err := b.sessions.Save(); err != nil {
		slog.Error("failed to save sessions on shutdown", "err", err)
		return fmt.Errorf("failed to save sessions: %w", err)
	}

	slog.Info("bot stopped, sessions saved")
	return nil
}

// sendMessage sends text to a recipient, optionally in a specific forum thread.
func (b *Bot) sendMessage(to tele.Recipient, text string, threadID int, parseMode string) (*tele.Message, error) {
	opts := &tele.SendOptions{
		ThreadID:  threadID,
		ParseMode: parseMode,
	}
	return b.bot.Send(to, text, opts)
}

// editMessage edits an existing message with new text and optional parse mode.
func (b *Bot) editMessage(msg *tele.Message, text string, parseMode string) (*tele.Message, error) {
	if parseMode != "" {
		return b.bot.Edit(msg, text, parseMode)
	}
	return b.bot.Edit(msg, text)
}

// react sets an emoji reaction on a message.
func (b *Bot) react(msg *tele.Message, emoji string) {
	err := b.bot.React(msg.Chat, msg, tele.Reactions{
		Reactions: []tele.Reaction{
			{Type: tele.ReactionTypeEmoji, Emoji: emoji},
		},
	})
	if err != nil {
		slog.Debug("failed to set reaction", "err", err, "emoji", emoji)
	}
}

// sendChunked splits text into Telegram-safe chunks and sends them
// sequentially in the given thread.
func (b *Bot) sendChunked(to tele.Recipient, text string, threadID int) error {
	chunks := ChunkMessage(text, b.cfg.Streaming.MaxMessageLength)
	for _, chunk := range chunks {
		html := MarkdownToHTML(chunk)
		_, err := b.sendMessage(to, html, threadID, tele.ModeHTML)
		if err != nil {
			// Fall back to plain text if HTML parse fails.
			_, err = b.sendMessage(to, chunk, threadID, "")
			if err != nil {
				return fmt.Errorf("failed to send chunk: %w", err)
			}
		}
	}
	return nil
}

// homeDir returns the current user's home directory, falling back to "/tmp".
func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "/tmp"
}
