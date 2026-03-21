package telegram

import (
	"strings"
	"testing"
)

func TestMarkdownToHTML(t *testing.T) {
	tests := []struct {
		name string
		md   string
		want string
	}{
		{"bold", "**bold**", "<b>bold</b>"},
		{"italic", "*italic*", "<i>italic</i>"},
		{"inline code", "`code`", "<code>code</code>"},
		{"code block with lang", "```go\nfunc main() {}\n```", "<pre>func main() {}</pre>"},
		{"code block no lang", "```\nhello\n```", "<pre>hello</pre>"},
		{"link", "[text](https://example.com)", `<a href="https://example.com">text</a>`},
		{"h1 heading", "# Heading", "<b>Heading</b>"},
		{"h2 heading", "## Sub Heading", "<b>Sub Heading</b>"},
		{"list items", "- item1\n- item2", "• item1\n• item2"},
		{"plain text", "just text", "just text"},
		{"mixed bold italic", "**bold** and *italic*", "<b>bold</b> and <i>italic</i>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MarkdownToHTML(tt.md)
			if got != tt.want {
				t.Errorf("MarkdownToHTML(%q):\n  got  %q\n  want %q", tt.md, got, tt.want)
			}
		})
	}
}

func TestChunkMessage(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		maxLen int
		checks func(t *testing.T, chunks []string)
	}{
		{
			name:   "short message",
			text:   "hello",
			maxLen: 100,
			checks: func(t *testing.T, chunks []string) {
				if len(chunks) != 1 {
					t.Errorf("expected 1 chunk, got %d", len(chunks))
				}
				if chunks[0] != "hello" {
					t.Errorf("expected %q, got %q", "hello", chunks[0])
				}
			},
		},
		{
			name:   "exact limit",
			text:   "12345",
			maxLen: 5,
			checks: func(t *testing.T, chunks []string) {
				if len(chunks) != 1 {
					t.Errorf("expected 1 chunk, got %d", len(chunks))
				}
			},
		},
		{
			name:   "over limit splits",
			text:   "123456",
			maxLen: 5,
			checks: func(t *testing.T, chunks []string) {
				if len(chunks) < 2 {
					t.Errorf("expected 2+ chunks, got %d", len(chunks))
				}
			},
		},
		{
			name:   "split at paragraph preferred",
			text:   "aaa\n\nbbb",
			maxLen: 5,
			checks: func(t *testing.T, chunks []string) {
				if len(chunks) != 2 {
					t.Fatalf("expected 2 chunks, got %d: %v", len(chunks), chunks)
				}
				if chunks[0] != "aaa" {
					t.Errorf("chunk[0]: got %q, want %q", chunks[0], "aaa")
				}
				if chunks[1] != "bbb" {
					t.Errorf("chunk[1]: got %q, want %q", chunks[1], "bbb")
				}
			},
		},
		{
			name:   "split at newline if no paragraph",
			text:   "aaa\nbbb",
			maxLen: 5,
			checks: func(t *testing.T, chunks []string) {
				if len(chunks) != 2 {
					t.Fatalf("expected 2 chunks, got %d: %v", len(chunks), chunks)
				}
				if chunks[0] != "aaa" {
					t.Errorf("chunk[0]: got %q, want %q", chunks[0], "aaa")
				}
				if chunks[1] != "bbb" {
					t.Errorf("chunk[1]: got %q, want %q", chunks[1], "bbb")
				}
			},
		},
		{
			name:   "no chunk exceeds maxLen",
			text:   "aaaa\n\nbbbb\n\ncccc\n\ndddd",
			maxLen: 10,
			checks: func(t *testing.T, chunks []string) {
				for i, c := range chunks {
					if len(c) > 10 {
						t.Errorf("chunk[%d] len=%d exceeds maxLen=10: %q", i, len(c), c)
					}
				}
			},
		},
		{
			name:   "all content preserved",
			text:   "hello\n\nworld\n\nfoo",
			maxLen: 8,
			checks: func(t *testing.T, chunks []string) {
				joined := strings.Join(chunks, "\n\n")
				if joined != "hello\n\nworld\n\nfoo" {
					t.Errorf("content not preserved: got %q", joined)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := ChunkMessage(tt.text, tt.maxLen)
			tt.checks(t, chunks)
		})
	}
}

func TestChunkMessage_CodeBlockPreserved(t *testing.T) {
	t.Run("no unbalanced fences", func(t *testing.T) {
		input := "some text\n\n```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```\n\nmore text"
		chunks := ChunkMessage(input, 30)

		for i, c := range chunks {
			fences := strings.Count(c, "```")
			if fences%2 != 0 {
				t.Errorf("chunk[%d] has unbalanced fences (%d): %q", i, fences, c)
			}
		}
	})

	t.Run("code block stays intact", func(t *testing.T) {
		codeBlock := "```\nline1\nline2\nline3\n```"
		input := "before\n\n" + codeBlock + "\n\nafter"
		// maxLen is smaller than the whole input but large enough for the code block.
		chunks := ChunkMessage(input, len(codeBlock)+2)

		found := false
		for _, c := range chunks {
			if strings.Contains(c, "```") {
				// The chunk that has the code block must contain the whole block.
				if !strings.Contains(c, codeBlock) {
					t.Errorf("code block split across chunks; chunk with fence: %q", c)
				}
				found = true
			}
		}
		if !found {
			t.Error("code block not found in any chunk")
		}
	})
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"less than", "<", "&lt;"},
		{"greater than", ">", "&gt;"},
		{"ampersand", "&", "&amp;"},
		{"normal text", "hello world", "hello world"},
		{"mixed", "a < b & c > d", "a &lt; b &amp; c &gt; d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeHTML(tt.in)
			if got != tt.want {
				t.Errorf("EscapeHTML(%q): got %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
