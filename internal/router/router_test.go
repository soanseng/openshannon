package router

import "testing"

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantCmd string
		wantArg string
	}{
		{"new with path", "/new ~/infra", "new", "~/infra"},
		{"shell with args", "/shell git status", "shell", "git status"},
		{"cd with path", "/cd ~/apps", "cd", "~/apps"},
		{"status no args", "/status", "status", ""},
		{"clear no args", "/clear", "clear", ""},
		{"kill with id", "/kill abc", "kill", "abc"},
		{"resume with id", "/resume abc", "resume", "abc"},
		{"sessions no args", "/sessions", "sessions", ""},
		{"cancel no args", "/cancel", "cancel", ""},
		{"long with args", "/long do something big", "long", "do something big"},
		{"help no args", "/help", "help", ""},
		{"plain text", "just a normal prompt", "", ""},
		{"empty string", "", "", ""},
		{"unknown command", "/unknown foo", "unknown", "foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := ParseCommand(tt.text)
			if cmd != tt.wantCmd {
				t.Errorf("ParseCommand(%q) cmd = %q, want %q", tt.text, cmd, tt.wantCmd)
			}
			if args != tt.wantArg {
				t.Errorf("ParseCommand(%q) args = %q, want %q", tt.text, args, tt.wantArg)
			}
		})
	}
}

func TestSessionKey(t *testing.T) {
	tests := []struct {
		name     string
		chatType string
		chatID   int64
		threadID int
		isTopic  bool
		userID   int64
		want     string
	}{
		{
			name:     "forum topic",
			chatType: "supergroup",
			chatID:   111,
			threadID: 42,
			isTopic:  true,
			userID:   888,
			want:     "topic:42",
		},
		{
			name:     "dm",
			chatType: "private",
			chatID:   999,
			threadID: 0,
			isTopic:  false,
			userID:   999,
			want:     "dm:999",
		},
		{
			name:     "plain group",
			chatType: "group",
			chatID:   222,
			threadID: 0,
			isTopic:  false,
			userID:   777,
			want:     "group:222",
		},
		{
			name:     "supergroup no topic",
			chatType: "supergroup",
			chatID:   333,
			threadID: 0,
			isTopic:  false,
			userID:   666,
			want:     "group:333",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SessionKey(tt.chatType, tt.chatID, tt.threadID, tt.isTopic, tt.userID)
			if got != tt.want {
				t.Errorf("SessionKey(%q, %d, %d, %v, %d) = %q, want %q",
					tt.chatType, tt.chatID, tt.threadID, tt.isTopic, tt.userID, got, tt.want)
			}
		})
	}
}

func TestIsAllowed(t *testing.T) {
	tests := []struct {
		name    string
		allowed map[int64]bool
		userID  int64
		want    bool
	}{
		{"user in map", map[int64]bool{100: true, 200: true}, 100, true},
		{"user not in map", map[int64]bool{100: true, 200: true}, 300, false},
		{"empty map", map[int64]bool{}, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAllowed(tt.allowed, tt.userID)
			if got != tt.want {
				t.Errorf("IsAllowed(%v, %d) = %v, want %v", tt.allowed, tt.userID, got, tt.want)
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	tests := []struct {
		name string
		path string
		home string
		want string
	}{
		{"tilde with subpath", "~/infra", "/home/scipio", "/home/scipio/infra"},
		{"absolute path", "/abs/path", "/home/scipio", "/abs/path"},
		{"tilde only", "~", "/home/scipio", "/home/scipio"},
		{"relative path", "relative", "/home/scipio", "relative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandHome(tt.path, tt.home)
			if got != tt.want {
				t.Errorf("ExpandHome(%q, %q) = %q, want %q", tt.path, tt.home, got, tt.want)
			}
		})
	}
}
