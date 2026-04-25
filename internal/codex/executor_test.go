package codex

import (
	"testing"

	"github.com/soanseng/openshannon/internal/config"
)

func TestBuildArgs_UsesWorkdirAndWritableAddDirs(t *testing.T) {
	cfg := config.CodexConfig{
		Model:          "gpt-5.2",
		Sandbox:        "workspace-write",
		ApprovalPolicy: "never",
		Flags:          []string{"--search"},
		AddDirs:        []string{"/srv/openshannon", "/srv/shared"},
	}
	exec := NewExecutor(cfg)

	args, outputPath := exec.buildArgs("/tmp/codex-last.txt", "/work/project", "inspect the repo")

	want := []string{
		"--ask-for-approval", "never",
		"exec",
		"--color", "never",
		"--output-last-message", "/tmp/codex-last.txt",
		"--cd", "/work/project",
		"--sandbox", "workspace-write",
		"--model", "gpt-5.2",
		"--search",
		"--add-dir", "/srv/openshannon",
		"--add-dir", "/srv/shared",
		"inspect the repo",
	}
	if outputPath != "/tmp/codex-last.txt" {
		t.Fatalf("outputPath = %q, want /tmp/codex-last.txt", outputPath)
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

func TestBinary_DefaultsToCodex(t *testing.T) {
	exec := NewExecutor(config.CodexConfig{})
	if got := exec.binary(); got != "codex" {
		t.Errorf("binary() = %q, want codex", got)
	}
}

func TestCodexEnv_DoesNotInheritDaemonSecrets(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "secret-telegram")
	t.Setenv("GEMINI_API_KEY", "secret-gemini")
	t.Setenv("OPENAI_API_KEY", "openai-secret")

	env := codexEnv()

	for _, item := range env {
		if item == "TELEGRAM_BOT_TOKEN=secret-telegram" || item == "GEMINI_API_KEY=secret-gemini" {
			t.Fatalf("codexEnv leaked daemon secret: %s", item)
		}
	}
	if !containsEnv(env, "OPENAI_API_KEY=openai-secret") {
		t.Fatalf("codexEnv should preserve OPENAI_API_KEY for Codex auth, got %v", env)
	}
}

func containsEnv(env []string, want string) bool {
	for _, item := range env {
		if item == want {
			return true
		}
	}
	return false
}
