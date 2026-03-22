package claude

import (
	"context"
	"testing"

	"github.com/soanseng/openshannon/internal/config"
)

func TestMockExecutor(t *testing.T) {
	mock := &MockExecutor{
		Response:  "Hello from Claude",
		SessionID: "sess-123",
		CostUSD:   0.05,
	}

	result, err := mock.Run(context.Background(), "test:1", "", "/tmp", "say hello", RunOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "Hello from Claude" {
		t.Errorf("got Text=%q, want %q", result.Text, "Hello from Claude")
	}
	if result.SessionID != "sess-123" {
		t.Errorf("got SessionID=%q, want %q", result.SessionID, "sess-123")
	}
	if result.CostUSD != 0.05 {
		t.Errorf("got CostUSD=%f, want %f", result.CostUSD, 0.05)
	}
	if result.ExitCode != 0 {
		t.Errorf("got ExitCode=%d, want 0", result.ExitCode)
	}
}

func TestMockExecutor_Error(t *testing.T) {
	mock := &MockExecutor{
		Err: context.DeadlineExceeded,
	}

	result, err := mock.Run(context.Background(), "test:1", "", "/tmp", "say hello", RunOpts{})
	if err != context.DeadlineExceeded {
		t.Fatalf("got err=%v, want %v", err, context.DeadlineExceeded)
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

func TestMockExecutor_RunWithStream(t *testing.T) {
	mock := &MockExecutor{
		Response:     "final answer",
		SessionID:    "sess-456",
		CostUSD:      0.10,
		StreamChunks: []string{"chunk1", "chunk2", "chunk3"},
	}

	var collected []string
	cb := func(text string) {
		collected = append(collected, text)
	}

	result, err := mock.RunWithStream(context.Background(), "test:1", "", "/tmp", "test prompt", RunOpts{}, cb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "final answer" {
		t.Errorf("got Text=%q, want %q", result.Text, "final answer")
	}
	if len(collected) != 3 {
		t.Fatalf("got %d chunks, want 3", len(collected))
	}
	for i, want := range []string{"chunk1", "chunk2", "chunk3"} {
		if collected[i] != want {
			t.Errorf("chunk[%d]=%q, want %q", i, collected[i], want)
		}
	}
}

func TestBuildArgs_NewSession(t *testing.T) {
	cfg := config.ClaudeConfig{
		Flags:        []string{"--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose"},
		MaxBudgetUSD: 10.0,
	}
	exec := NewCLIExecutor(cfg)
	args := exec.buildArgs("", "explain this code", RunOpts{})

	// workdir is set via cmd.Dir, not via -w flag (which means --worktree in Claude CLI)
	want := []string{
		"-p",
		"--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose",
		"--max-turns", "50",
		"explain this code",
	}

	if len(args) != len(want) {
		t.Fatalf("got %d args, want %d: %v", len(args), len(want), args)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Errorf("args[%d]=%q, want %q", i, args[i], want[i])
		}
	}
}

func TestBuildArgs_ResumeSession(t *testing.T) {
	cfg := config.ClaudeConfig{
		Flags:        []string{"--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose"},
		MaxBudgetUSD: 10.0,
	}
	exec := NewCLIExecutor(cfg)
	args := exec.buildArgs("sess-abc", "continue", RunOpts{})

	want := []string{
		"-p",
		"--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose",
		"--resume", "sess-abc",
		"--max-turns", "50",
		"continue",
	}

	if len(args) != len(want) {
		t.Fatalf("got %d args, want %d: %v", len(args), len(want), args)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Errorf("args[%d]=%q, want %q", i, args[i], want[i])
		}
	}
}

func TestBuildArgs_NoFlags(t *testing.T) {
	cfg := config.ClaudeConfig{}
	exec := NewCLIExecutor(cfg)
	args := exec.buildArgs("", "hello", RunOpts{})

	want := []string{"-p", "hello"}
	if len(args) != len(want) {
		t.Fatalf("got %d args, want %d: %v", len(args), len(want), args)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Errorf("args[%d]=%q, want %q", i, args[i], want[i])
		}
	}
}

func TestParseResult(t *testing.T) {
	input := `{"result":"Hello world","session_id":"sess-xyz","total_cost_usd":0.03}`
	result, err := parseResult([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "Hello world" {
		t.Errorf("got Text=%q, want %q", result.Text, "Hello world")
	}
	if result.SessionID != "sess-xyz" {
		t.Errorf("got SessionID=%q, want %q", result.SessionID, "sess-xyz")
	}
	if result.CostUSD != 0.03 {
		t.Errorf("got CostUSD=%f, want %f", result.CostUSD, 0.03)
	}
}

func TestParseResult_InvalidJSON(t *testing.T) {
	_, err := parseResult([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseResult_NDJSONLastLine(t *testing.T) {
	// Simulates stream-json output where multiple lines precede the result
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"hi"}]}}
{"type":"result","result":"final answer","session_id":"sess-ndjson","total_cost_usd":0.07}`

	result, err := parseResult([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "final answer" {
		t.Errorf("got Text=%q, want %q", result.Text, "final answer")
	}
	if result.SessionID != "sess-ndjson" {
		t.Errorf("got SessionID=%q, want %q", result.SessionID, "sess-ndjson")
	}
	if result.CostUSD != 0.07 {
		t.Errorf("got CostUSD=%f, want %f", result.CostUSD, 0.07)
	}
}

func TestParseStreamLine_TextContent(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello "},{"type":"text","text":"world"}]}}`
	text, isResult, result, err := parseStreamLine([]byte(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isResult {
		t.Error("expected isResult=false")
	}
	if result != nil {
		t.Error("expected nil result")
	}
	if text != "Hello world" {
		t.Errorf("got text=%q, want %q", text, "Hello world")
	}
}

func TestParseStreamLine_ContentBlockDelta(t *testing.T) {
	line := `{"type":"content_block_delta","delta":{"type":"text_delta","text":"streaming chunk"}}`
	text, isResult, _, err := parseStreamLine([]byte(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isResult {
		t.Error("expected isResult=false")
	}
	if text != "streaming chunk" {
		t.Errorf("got text=%q, want %q", text, "streaming chunk")
	}
}

func TestParseStreamLine_ResultEvent(t *testing.T) {
	line := `{"type":"result","result":"done","session_id":"sess-final","total_cost_usd":0.12}`
	text, isResult, result, err := parseStreamLine([]byte(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isResult {
		t.Error("expected isResult=true")
	}
	if text != "" {
		t.Errorf("got text=%q, want empty", text)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Text != "done" {
		t.Errorf("got Text=%q, want %q", result.Text, "done")
	}
	if result.SessionID != "sess-final" {
		t.Errorf("got SessionID=%q, want %q", result.SessionID, "sess-final")
	}
	if result.CostUSD != 0.12 {
		t.Errorf("got CostUSD=%f, want %f", result.CostUSD, 0.12)
	}
}

func TestParseStreamLine_UnknownType(t *testing.T) {
	line := `{"type":"system","message":"initializing"}`
	text, isResult, result, err := parseStreamLine([]byte(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isResult {
		t.Error("expected isResult=false")
	}
	if result != nil {
		t.Error("expected nil result")
	}
	if text != "" {
		t.Errorf("got text=%q, want empty", text)
	}
}

func TestParseStreamLine_InvalidJSON(t *testing.T) {
	_, _, _, err := parseStreamLine([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseStreamLine_EmptyLine(t *testing.T) {
	text, isResult, result, err := parseStreamLine([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isResult || result != nil || text != "" {
		t.Error("expected all zero values for empty line")
	}
}

// TestMockExecutor_ImplementsInterface verifies MockExecutor satisfies Executor.
func TestMockExecutor_ImplementsInterface(t *testing.T) {
	var _ Executor = (*MockExecutor)(nil)
}

// TestCLIExecutor_ImplementsInterface verifies CLIExecutor satisfies Executor.
func TestCLIExecutor_ImplementsInterface(t *testing.T) {
	var _ Executor = (*CLIExecutor)(nil)
}
