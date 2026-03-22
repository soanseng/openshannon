package notify

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/scipio/openshannon/internal/config"
)

// captured holds request data recorded by the test HTTP server.
type captured struct {
	mu      sync.Mutex
	method  string
	path    string
	body    string
	headers http.Header
}

func (c *captured) set(r *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.method = r.Method
	c.path = r.URL.Path
	body, _ := io.ReadAll(r.Body)
	c.body = string(body)
	c.headers = r.Header.Clone()
}

func (c *captured) get() (string, string, string, http.Header) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.method, c.path, c.body, c.headers
}

func TestNotifier_Send(t *testing.T) {
	cap := &captured{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.set(r)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.NotifyConfig{
		Enabled:    true,
		NtfyServer: srv.URL,
		NtfyTopic:  "test-topic",
		NtfyToken:  "tk_secret123",
		Events:     []string{"daemon_start", "daemon_crash"},
	}

	n := New(cfg)
	err := n.Send("daemon_start", "Daemon started successfully")
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	method, path, body, headers := cap.get()

	if method != http.MethodPost {
		t.Errorf("expected POST, got %s", method)
	}
	if path != "/test-topic" {
		t.Errorf("expected path /test-topic, got %s", path)
	}
	if body != "Daemon started successfully" {
		t.Errorf("expected body 'Daemon started successfully', got %q", body)
	}
	if got := headers.Get("Title"); got != "daemon_start" {
		t.Errorf("expected Title header 'daemon_start', got %q", got)
	}
	if got := headers.Get("Priority"); got != "2" {
		t.Errorf("expected Priority header '2' (low), got %q", got)
	}
	if got := headers.Get("Tags"); got != "robot" {
		t.Errorf("expected Tags header 'robot', got %q", got)
	}
	if got := headers.Get("Authorization"); got != "Bearer tk_secret123" {
		t.Errorf("expected Authorization header 'Bearer tk_secret123', got %q", got)
	}
}

func TestNotifier_DisabledSkips(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.NotifyConfig{
		Enabled:    false,
		NtfyServer: srv.URL,
		NtfyTopic:  "test-topic",
		NtfyToken:  "tk_secret123",
		Events:     []string{"daemon_start"},
	}

	n := New(cfg)
	err := n.Send("daemon_start", "Daemon started")
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if called {
		t.Error("expected no HTTP request when notifications disabled")
	}
}

func TestNotifier_EventFiltering(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.NotifyConfig{
		Enabled:    true,
		NtfyServer: srv.URL,
		NtfyTopic:  "test-topic",
		NtfyToken:  "tk_secret123",
		Events:     []string{"daemon_start"},
	}

	n := New(cfg)

	// Allowed event should be sent.
	err := n.Send("daemon_start", "Started")
	if err != nil {
		t.Fatalf("Send returned error for allowed event: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 HTTP call for allowed event, got %d", callCount)
	}

	// Unknown event should be skipped (returns nil, no HTTP call).
	err = n.Send("unknown_event", "Should not send")
	if err != nil {
		t.Fatalf("Send returned error for filtered event: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected still 1 HTTP call after filtered event, got %d", callCount)
	}
}

func TestNotifier_NoToken(t *testing.T) {
	cap := &captured{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.set(r)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.NotifyConfig{
		Enabled:    true,
		NtfyServer: srv.URL,
		NtfyTopic:  "test-topic",
		NtfyToken:  "", // no token
		Events:     []string{"daemon_start"},
	}

	n := New(cfg)
	err := n.Send("daemon_start", "Started")
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	_, _, _, headers := cap.get()
	if got := headers.Get("Authorization"); got != "" {
		t.Errorf("expected no Authorization header, got %q", got)
	}
}

func TestEventPriority(t *testing.T) {
	tests := []struct {
		event    string
		expected Priority
	}{
		{"daemon_start", PriorityLow},
		{"daemon_crash", PriorityHigh},
		{"safety_block", PriorityDefault},
		{"long_task_complete", PriorityLow},
		{"claude_crash", PriorityHigh},
		{"session_corrupt", PriorityHigh},
		{"panic", PriorityUrgent},
		{"daemon_stop", PriorityLow},
		{"unknown", PriorityDefault},
	}

	for _, tt := range tests {
		t.Run(tt.event, func(t *testing.T) {
			got := EventPriority(tt.event)
			if got != tt.expected {
				t.Errorf("EventPriority(%q) = %d, want %d", tt.event, got, tt.expected)
			}
		})
	}
}

func TestNotifier_SendWithPriority(t *testing.T) {
	cap := &captured{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.set(r)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.NotifyConfig{
		Enabled:    true,
		NtfyServer: srv.URL,
		NtfyTopic:  "test-topic",
		NtfyToken:  "tk_abc",
		Events:     []string{"daemon_crash"},
	}

	n := New(cfg)
	err := n.SendWithPriority(context.Background(), "daemon_crash", "Process crashed!", PriorityUrgent)
	if err != nil {
		t.Fatalf("SendWithPriority returned error: %v", err)
	}

	_, _, body, headers := cap.get()
	if body != "Process crashed!" {
		t.Errorf("expected body 'Process crashed!', got %q", body)
	}
	// Priority should be the explicitly passed value (5), not the default for daemon_crash (4).
	if got := headers.Get("Priority"); got != "5" {
		t.Errorf("expected Priority header '5' (urgent), got %q", got)
	}
}
