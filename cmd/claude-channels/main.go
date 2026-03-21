package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/scipio/claude-channels/internal/claude"
	"github.com/scipio/claude-channels/internal/config"
	"github.com/scipio/claude-channels/internal/notify"
	"github.com/scipio/claude-channels/internal/safety"
	"github.com/scipio/claude-channels/internal/session"
	"github.com/scipio/claude-channels/internal/telegram"
)

var version = "dev"

func main() {
	configPath := flag.String("config", "", "path to config.yaml")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Resolve config path: flag > default location.
	cfgPath := *configPath
	if cfgPath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			candidate := filepath.Join(home, ".config", "claude-channels", "config.yaml")
			if _, statErr := os.Stat(candidate); statErr == nil {
				cfgPath = candidate
			}
		}
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("failed to load config", "path", cfgPath, "err", err)
		os.Exit(1)
	}

	if cfg.Telegram.Token == "" {
		slog.Error("telegram token is required (set telegram.token in config or TELEGRAM_BOT_TOKEN env)")
		os.Exit(1)
	}

	// Set version so /status can display it.
	telegram.Version = version

	// Resolve storage dir.
	storageDir := expandHome(cfg.Storage.Dir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		slog.Error("failed to create storage dir", "dir", storageDir, "err", err)
		os.Exit(1)
	}

	// Init components.
	sessions := session.NewManager(storageDir)
	if err := sessions.Load(); err != nil {
		slog.Warn("failed to load sessions (starting fresh)", "err", err)
	}

	filter := safety.NewFilter(cfg.Safety)
	executor := claude.NewCLIExecutor(cfg.Claude)
	notifier := notify.New(cfg.Notify)

	bot, err := telegram.NewBot(cfg, sessions, executor, filter, notifier)
	if err != nil {
		slog.Error("failed to create bot", "err", err)
		os.Exit(1)
	}

	// Graceful shutdown on SIGINT / SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("starting claude-channels",
		"version", version,
		"config", cfgPath,
		"storage", storageDir,
	)

	if err := bot.Start(ctx); err != nil {
		slog.Error("bot exited with error", "err", err)
		os.Exit(1)
	}

	slog.Info("claude-channels stopped")
}

// expandHome expands a leading ~ to the user's home directory.
func expandHome(path string) string {
	if path == "~" || len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if path == "~" {
			return home
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

