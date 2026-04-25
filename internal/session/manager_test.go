package session

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "session-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestManager_CreateAndGet(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		workdir string
	}{
		{name: "topic session", key: "topic:12345", workdir: "/tmp/work1"},
		{name: "dm session", key: "dm:67890", workdir: "/tmp/work2"},
		{name: "group session", key: "group:11111", workdir: "/tmp/work3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tempDir(t)
			m := NewManager(dir)

			sess, err := m.Create(tt.key, tt.workdir)
			if err != nil {
				t.Fatalf("Create(%q, %q) returned error: %v", tt.key, tt.workdir, err)
			}
			if sess.Key != tt.key {
				t.Errorf("Key = %q, want %q", sess.Key, tt.key)
			}
			if sess.Workdir != tt.workdir {
				t.Errorf("Workdir = %q, want %q", sess.Workdir, tt.workdir)
			}
			if sess.State != StateActive {
				t.Errorf("State = %q, want %q", sess.State, StateActive)
			}
			if sess.CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}
			if sess.LastActiveAt.IsZero() {
				t.Error("LastActiveAt should not be zero")
			}

			got := m.Get(tt.key)
			if got == nil {
				t.Fatalf("Get(%q) returned nil", tt.key)
			}
			if got.Key != tt.key {
				t.Errorf("Get Key = %q, want %q", got.Key, tt.key)
			}
		})
	}
}

func TestManager_CreateDuplicate(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	_, err := m.Create("topic:1", "/tmp/w")
	if err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	_, err = m.Create("topic:1", "/tmp/w2")
	if err == nil {
		t.Fatal("expected error on duplicate Create, got nil")
	}
}

func TestManager_GetMissing(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	got := m.Get("nonexistent")
	if got != nil {
		t.Errorf("Get(nonexistent) = %v, want nil", got)
	}
}

func TestManager_GetOrCreate(t *testing.T) {
	tests := []struct {
		name           string
		preCreate      bool
		key            string
		defaultWorkdir string
		wantWorkdir    string
	}{
		{
			name:           "creates when missing",
			preCreate:      false,
			key:            "topic:100",
			defaultWorkdir: "/tmp/default",
			wantWorkdir:    "/tmp/default",
		},
		{
			name:           "returns existing",
			preCreate:      true,
			key:            "topic:200",
			defaultWorkdir: "/tmp/override",
			wantWorkdir:    "/tmp/original",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tempDir(t)
			m := NewManager(dir)

			if tt.preCreate {
				_, err := m.Create(tt.key, "/tmp/original")
				if err != nil {
					t.Fatalf("pre-create failed: %v", err)
				}
			}

			sess := m.GetOrCreate(tt.key, tt.defaultWorkdir)
			if sess == nil {
				t.Fatal("GetOrCreate returned nil")
			}
			if sess.Key != tt.key {
				t.Errorf("Key = %q, want %q", sess.Key, tt.key)
			}
			if sess.Workdir != tt.wantWorkdir {
				t.Errorf("Workdir = %q, want %q", sess.Workdir, tt.wantWorkdir)
			}
		})
	}
}

func TestManager_Clear(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	sess, err := m.Create("topic:1", "/tmp/work")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := m.SetClaudeSession("topic:1", "claude-abc-123"); err != nil {
		t.Fatalf("SetClaudeSession failed: %v", err)
	}

	if err := m.Clear("topic:1"); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	sess = m.Get("topic:1")
	if sess == nil {
		t.Fatal("session should still exist after Clear")
	}
	if sess.ClaudeSession != "" {
		t.Errorf("ClaudeSession = %q, want empty after Clear", sess.ClaudeSession)
	}
	if sess.Workdir != "/tmp/work" {
		t.Errorf("Workdir = %q, want /tmp/work (preserved after Clear)", sess.Workdir)
	}
	if sess.State != StateActive {
		t.Errorf("State = %q, want %q after Clear", sess.State, StateActive)
	}
}

func TestManager_ClearMissing(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	err := m.Clear("nonexistent")
	if err == nil {
		t.Error("expected error clearing nonexistent session")
	}
}

func TestManager_Kill(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	_, err := m.Create("topic:1", "/tmp/work")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := m.Kill("topic:1"); err != nil {
		t.Fatalf("Kill failed: %v", err)
	}

	got := m.Get("topic:1")
	if got != nil {
		t.Error("session should be nil after Kill")
	}
}

func TestManager_KillMissing(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	err := m.Kill("nonexistent")
	if err == nil {
		t.Error("expected error killing nonexistent session")
	}
}

func TestManager_SetWorkdir(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	_, err := m.Create("topic:1", "/tmp/old")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := m.SetWorkdir("topic:1", "/tmp/new"); err != nil {
		t.Fatalf("SetWorkdir failed: %v", err)
	}

	sess := m.Get("topic:1")
	if sess.Workdir != "/tmp/new" {
		t.Errorf("Workdir = %q, want /tmp/new", sess.Workdir)
	}
}

func TestManager_SetWorkdirMissing(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	err := m.SetWorkdir("nonexistent", "/tmp/x")
	if err == nil {
		t.Error("expected error setting workdir on nonexistent session")
	}
}

func TestManager_SetClaudeSession(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	_, err := m.Create("topic:1", "/tmp/w")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := m.SetClaudeSession("topic:1", "sess-xyz"); err != nil {
		t.Fatalf("SetClaudeSession failed: %v", err)
	}

	sess := m.Get("topic:1")
	if sess.ClaudeSession != "sess-xyz" {
		t.Errorf("ClaudeSession = %q, want sess-xyz", sess.ClaudeSession)
	}
}

func TestManager_SetAgent(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	_, err := m.Create("topic:1", "/tmp/w")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := m.SetAgent("topic:1", "codex"); err != nil {
		t.Fatalf("SetAgent failed: %v", err)
	}

	sess := m.Get("topic:1")
	if sess.Agent != "codex" {
		t.Errorf("Agent = %q, want codex", sess.Agent)
	}
}

func TestManager_List(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	keys := []string{"topic:1", "dm:2", "group:3"}
	for _, k := range keys {
		if _, err := m.Create(k, "/tmp/"+k); err != nil {
			t.Fatalf("Create(%q) failed: %v", k, err)
		}
	}

	list := m.List()
	if len(list) != 3 {
		t.Fatalf("List() returned %d sessions, want 3", len(list))
	}

	found := make(map[string]bool)
	for _, s := range list {
		found[s.Key] = true
	}
	for _, k := range keys {
		if !found[k] {
			t.Errorf("List() missing key %q", k)
		}
	}
}

func TestManager_Persistence(t *testing.T) {
	dir := tempDir(t)

	// Create and save
	m1 := NewManager(dir)
	_, err := m1.Create("topic:1", "/tmp/w1")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := m1.SetClaudeSession("topic:1", "sess-1"); err != nil {
		t.Fatalf("SetClaudeSession failed: %v", err)
	}
	_, err = m1.Create("dm:2", "/tmp/w2")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := m1.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	jsonPath := filepath.Join(dir, "sessions.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Fatal("sessions.json not created")
	}

	// Load into fresh manager
	m2 := NewManager(dir)
	if err := m2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	list := m2.List()
	if len(list) != 2 {
		t.Fatalf("loaded %d sessions, want 2", len(list))
	}

	sess := m2.Get("topic:1")
	if sess == nil {
		t.Fatal("topic:1 not found after Load")
	}
	if sess.ClaudeSession != "sess-1" {
		t.Errorf("ClaudeSession = %q, want sess-1", sess.ClaudeSession)
	}
	if sess.Workdir != "/tmp/w1" {
		t.Errorf("Workdir = %q, want /tmp/w1", sess.Workdir)
	}
	if sess.State != StateActive {
		t.Errorf("State = %q, want %q", sess.State, StateActive)
	}
}

func TestManager_LoadMissing(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	// Loading when no file exists should not error (fresh start)
	if err := m.Load(); err != nil {
		t.Fatalf("Load on missing file should not error, got: %v", err)
	}
	if len(m.List()) != 0 {
		t.Error("expected empty session list after loading missing file")
	}
}

func TestManager_ActiveSession(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	keys := []string{"topic:1", "topic:2", "dm:3"}
	for _, k := range keys {
		if _, err := m.Create(k, "/tmp/"+k); err != nil {
			t.Fatalf("Create(%q) failed: %v", k, err)
		}
	}

	// All sessions should be active simultaneously
	for _, k := range keys {
		sess := m.Get(k)
		if sess == nil {
			t.Fatalf("Get(%q) returned nil", k)
		}
		if sess.State != StateActive {
			t.Errorf("session %q State = %q, want %q", k, sess.State, StateActive)
		}
	}
}

func TestManager_Touch(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	_, err := m.Create("topic:1", "/tmp/w")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	sess := m.Get("topic:1")
	before := sess.LastActiveAt

	// Small sleep to ensure time difference
	time.Sleep(10 * time.Millisecond)

	m.Touch("topic:1")

	sess = m.Get("topic:1")
	if !sess.LastActiveAt.After(before) {
		t.Errorf("LastActiveAt not updated: before=%v, after=%v", before, sess.LastActiveAt)
	}
}

func TestManager_TouchMissing(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	// Touch on nonexistent key should be a no-op (no panic)
	m.Touch("nonexistent")
}

func TestManager_ConcurrentAccess(t *testing.T) {
	dir := tempDir(t)
	m := NewManager(dir)

	var wg sync.WaitGroup
	const n = 50

	// Concurrent creates with different keys
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "topic:" + string(rune('A'+i))
			_, _ = m.Create(key, "/tmp/"+key)
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = m.List()
		}(i)
	}
	wg.Wait()

	// Concurrent touches
	list := m.List()
	for _, s := range list {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			m.Touch(key)
		}(s.Key)
	}
	wg.Wait()
}
