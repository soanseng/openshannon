package codex

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/soanseng/openshannon/internal/claude"
	"github.com/soanseng/openshannon/internal/config"
)

// Executor spawns codex exec processes.
type Executor struct {
	cfg config.CodexConfig

	mu   sync.Mutex
	cmds map[string]*exec.Cmd
}

// NewExecutor creates a Codex CLI executor.
func NewExecutor(cfg config.CodexConfig) *Executor {
	return &Executor{
		cfg:  cfg,
		cmds: make(map[string]*exec.Cmd),
	}
}

func (e *Executor) binary() string {
	if e.cfg.Binary != "" {
		return e.cfg.Binary
	}
	return "codex"
}

func (e *Executor) buildArgs(outputPath, workdir, prompt string) ([]string, string) {
	args := make([]string, 0, 16+len(e.cfg.Flags)+len(e.cfg.AddDirs)*2)
	if e.cfg.ApprovalPolicy != "" {
		args = append(args, "--ask-for-approval", e.cfg.ApprovalPolicy)
	}
	args = append(args,
		"exec",
		"--color", "never",
		"--output-last-message", outputPath,
		"--cd", workdir,
	)
	if e.cfg.Sandbox != "" {
		args = append(args, "--sandbox", e.cfg.Sandbox)
	}
	if e.cfg.Model != "" {
		args = append(args, "--model", e.cfg.Model)
	}
	args = append(args, e.cfg.Flags...)
	for _, dir := range e.cfg.AddDirs {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		args = append(args, "--add-dir", expandHome(dir))
	}
	args = append(args, prompt)
	return args, outputPath
}

// Run invokes Codex and returns the final assistant message.
func (e *Executor) Run(ctx context.Context, key, _, workdir, prompt string, _ claude.RunOpts) (*claude.Result, error) {
	tmp, err := os.CreateTemp("", "openshannon-codex-*.txt")
	if err != nil {
		return nil, fmt.Errorf("create codex output file: %w", err)
	}
	outputPath := tmp.Name()
	_ = tmp.Close()
	defer os.Remove(outputPath)

	workdir = expandHome(workdir)
	args, _ := e.buildArgs(outputPath, workdir, prompt)
	cmd := exec.CommandContext(ctx, e.binary(), args...)
	cmd.Dir = workdir
	cmd.Env = codexEnv()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	e.mu.Lock()
	e.cmds[key] = cmd
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.cmds, key)
		e.mu.Unlock()
	}()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("codex CLI failed: %w (stderr: %s)", err, stderr.String())
	}

	textBytes, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("read codex output file: %w", err)
	}
	text := strings.TrimSpace(string(textBytes))
	if text == "" {
		text = strings.TrimSpace(stdout.String())
	}
	if text == "" {
		slog.Warn("codex returned empty output", "stderr", stderr.String())
	}

	return &claude.Result{
		Text:     text,
		ExitCode: cmd.ProcessState.ExitCode(),
	}, nil
}

// RunWithStream returns the final Codex response through the stream callback.
func (e *Executor) RunWithStream(ctx context.Context, key, sessionID, workdir, prompt string, opts claude.RunOpts, cb claude.StreamCallback) (*claude.Result, error) {
	result, err := e.Run(ctx, key, sessionID, workdir, prompt, opts)
	if err != nil {
		return nil, err
	}
	if cb != nil && result.Text != "" {
		cb(result.Text)
	}
	return result, nil
}

// Cancel terminates the running Codex CLI process for the given session key.
func (e *Executor) Cancel(key string) error {
	e.mu.Lock()
	cmd := e.cmds[key]
	e.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

func codexEnv() []string {
	keys := []string{
		"HOME",
		"TMPDIR",
		"PATH",
		"SHELL",
		"LANG",
		"CODEX_HOME",
		"OPENAI_API_KEY",
		"OPENAI_BASE_URL",
		"HTTPS_PROXY",
		"HTTP_PROXY",
		"NO_PROXY",
	}
	env := make([]string, 0, len(keys))
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+value)
		}
	}
	if !hasEnv(env, "HOME") {
		env = append(env, "HOME="+expandHome("~"))
	}
	if !hasEnv(env, "TMPDIR") {
		env = append(env, "TMPDIR=/tmp")
	}
	if !hasEnv(env, "SHELL") {
		env = append(env, "SHELL=/bin/sh")
	}
	if !hasEnv(env, "LANG") {
		env = append(env, "LANG=en_US.UTF-8")
	}
	return env
}

func hasEnv(env []string, key string) bool {
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return true
		}
	}
	return false
}

func expandHome(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return path
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}

var _ claude.Executor = (*Executor)(nil)
