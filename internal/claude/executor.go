package claude

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"

	"github.com/scipio/claude-channels/internal/config"
)

// Result represents output from a Claude CLI invocation.
type Result struct {
	Text      string  `json:"result"`
	SessionID string  `json:"session_id"`
	ExitCode  int     `json:"-"`
	CostUSD   float64 `json:"total_cost_usd"`
}

// StreamCallback is called with incremental text as Claude streams output.
type StreamCallback func(text string)

// Executor interface for running Claude CLI commands.
// The key parameter identifies the session (e.g. "topic:123") so that
// concurrent topics each get their own process.
// RunOpts holds per-invocation options beyond the base config.
type RunOpts struct {
	Model string // optional model override (e.g. "haiku", "sonnet", "opus")
}

type Executor interface {
	Run(ctx context.Context, key, sessionID, workdir, prompt string, opts RunOpts) (*Result, error)
	RunWithStream(ctx context.Context, key, sessionID, workdir, prompt string, opts RunOpts, cb StreamCallback) (*Result, error)
	Cancel(key string) error
}

// CLIExecutor spawns claude -p processes.
type CLIExecutor struct {
	cfg config.ClaudeConfig

	mu   sync.Mutex
	cmds map[string]*exec.Cmd
}

// NewCLIExecutor creates a new CLIExecutor with the given configuration.
func NewCLIExecutor(cfg config.ClaudeConfig) *CLIExecutor {
	return &CLIExecutor{
		cfg:  cfg,
		cmds: make(map[string]*exec.Cmd),
	}
}

// buildArgs constructs the CLI argument list for a Claude invocation.
// Note: workdir is set via cmd.Dir, not via -w flag.
// Claude CLI's -w flag means --worktree (git worktree), not working directory.
func (e *CLIExecutor) buildArgs(sessionID, prompt string, opts RunOpts) []string {
	args := []string{"-p"}
	args = append(args, e.cfg.Flags...)
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	if sessionID != "" {
		args = append(args, "--resume", sessionID)
	}
	// Safety net: limit turns when a budget is configured.
	if e.cfg.MaxBudgetUSD > 0 {
		args = append(args, "--max-turns", "50")
	}
	args = append(args, prompt)
	return args
}

// binary returns the configured binary name, defaulting to "claude".
func (e *CLIExecutor) binary() string {
	if e.cfg.Binary != "" {
		return e.cfg.Binary
	}
	return "claude"
}

// Run invokes the Claude CLI and returns the complete result.
// All stdout is collected, and the last line is parsed as the result JSON.
func (e *CLIExecutor) Run(ctx context.Context, key, sessionID, workdir, prompt string, opts RunOpts) (*Result, error) {
	args := e.buildArgs(sessionID, prompt, opts)
	cmd := exec.CommandContext(ctx, e.binary(), args...)
	cmd.Dir = workdir

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
		return nil, fmt.Errorf("claude CLI failed: %w (stderr: %s)", err, stderr.String())
	}

	result, err := parseResult(stdout.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to parse claude output: %w", err)
	}
	result.ExitCode = cmd.ProcessState.ExitCode()

	if result != nil && e.cfg.MaxBudgetUSD > 0 && result.CostUSD > e.cfg.MaxBudgetUSD {
		slog.Warn("cost exceeded budget", "cost", result.CostUSD, "budget", e.cfg.MaxBudgetUSD)
	}

	return result, nil
}

// RunWithStream invokes the Claude CLI and streams output line by line.
// The StreamCallback is called for each text chunk received. The final
// Result is returned when the process completes.
func (e *CLIExecutor) RunWithStream(ctx context.Context, key, sessionID, workdir, prompt string, opts RunOpts, cb StreamCallback) (*Result, error) {
	args := e.buildArgs(sessionID, prompt, opts)
	cmd := exec.CommandContext(ctx, e.binary(), args...)
	cmd.Dir = workdir

	e.mu.Lock()
	e.cmds[key] = cmd
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.cmds, key)
		e.mu.Unlock()
	}()

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start claude CLI: %w", err)
	}

	var finalResult *Result
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		text, isResult, result, parseErr := parseStreamLine(line)
		if parseErr != nil {
			continue // skip malformed lines
		}
		if text != "" && cb != nil {
			cb(text)
		}
		if isResult && result != nil {
			finalResult = result
		}
	}
	if err := scanner.Err(); err != nil {
		slog.Warn("scanner error reading claude output", "err", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("claude CLI failed: %w (stderr: %s)", err, stderr.String())
	}

	if finalResult == nil {
		return nil, fmt.Errorf("no result event received from claude CLI")
	}
	finalResult.ExitCode = cmd.ProcessState.ExitCode()

	if finalResult != nil && e.cfg.MaxBudgetUSD > 0 && finalResult.CostUSD > e.cfg.MaxBudgetUSD {
		slog.Warn("cost exceeded budget", "cost", finalResult.CostUSD, "budget", e.cfg.MaxBudgetUSD)
	}

	return finalResult, nil
}

// Cancel terminates the running Claude CLI process for the given session key.
func (e *CLIExecutor) Cancel(key string) error {
	e.mu.Lock()
	cmd := e.cmds[key]
	e.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}

// parseResult parses the output from Claude CLI. It handles both single-line
// JSON and NDJSON (stream-json) output, using the last non-empty line.
func parseResult(data []byte) (*Result, error) {
	// Find the last non-empty line.
	lines := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty output")
	}

	lastLine := lines[len(lines)-1]
	var result Result
	if err := json.Unmarshal(lastLine, &result); err != nil {
		return nil, fmt.Errorf("failed to parse result JSON: %w", err)
	}
	return &result, nil
}

// streamLine is the intermediate JSON structure for NDJSON stream parsing.
// The Message field is json.RawMessage because different event types use
// different shapes for this field (object vs string).
type streamLine struct {
	Type    string          `json:"type"`
	Result  string          `json:"result"`
	Session string          `json:"session_id"`
	Cost    float64         `json:"total_cost_usd"`
	Message json.RawMessage `json:"message"`
	Delta   *streamDelta    `json:"delta"`
}

type streamMsg struct {
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type streamDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// parseStreamLine parses a single NDJSON line from Claude's stream-json output.
// It returns extracted text content, whether this is a result event, the parsed
// result (if applicable), and any parse error.
func parseStreamLine(line []byte) (text string, isResult bool, result *Result, err error) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return "", false, nil, nil
	}

	var sl streamLine
	if err := json.Unmarshal(line, &sl); err != nil {
		return "", false, nil, fmt.Errorf("failed to parse stream line: %w", err)
	}

	switch sl.Type {
	case "result":
		r := &Result{
			Text:      sl.Result,
			SessionID: sl.Session,
			CostUSD:   sl.Cost,
		}
		return "", true, r, nil

	case "assistant":
		if len(sl.Message) > 0 {
			var msg streamMsg
			if err := json.Unmarshal(sl.Message, &msg); err == nil {
				var sb strings.Builder
				for _, block := range msg.Content {
					if block.Type == "text" {
						sb.WriteString(block.Text)
					}
				}
				return sb.String(), false, nil, nil
			}
		}

	case "content_block_delta":
		if sl.Delta != nil && sl.Delta.Type == "text_delta" {
			return sl.Delta.Text, false, nil, nil
		}
	}

	return "", false, nil, nil
}

// MockExecutor is a test double for the Executor interface.
type MockExecutor struct {
	Response     string
	SessionID    string
	CostUSD      float64
	Err          error
	StreamChunks []string
}

// Run returns the pre-configured response or error.
func (m *MockExecutor) Run(_ context.Context, _, _, _, _ string, _ RunOpts) (*Result, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return &Result{
		Text:      m.Response,
		SessionID: m.SessionID,
		CostUSD:   m.CostUSD,
	}, nil
}

// RunWithStream calls the callback for each StreamChunk, then returns the result.
func (m *MockExecutor) RunWithStream(_ context.Context, _, _, _, _ string, _ RunOpts, cb StreamCallback) (*Result, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	for _, chunk := range m.StreamChunks {
		if cb != nil {
			cb(chunk)
		}
	}
	return &Result{
		Text:      m.Response,
		SessionID: m.SessionID,
		CostUSD:   m.CostUSD,
	}, nil
}

// Cancel is a no-op for the mock executor.
func (m *MockExecutor) Cancel(_ string) error {
	return nil
}
