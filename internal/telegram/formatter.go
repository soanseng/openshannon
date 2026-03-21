package telegram

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Code block: ```lang\n...\n``` (multiline, non-greedy).
	reCodeBlock = regexp.MustCompile("(?s)```(?:\\w*)\n(.*?)\n```")

	// Inline code: `...`
	reInlineCode = regexp.MustCompile("`([^`]+)`")

	// Link: [text](url)
	reLink = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

	// Bold: **text**
	reBold = regexp.MustCompile(`\*\*(.+?)\*\*`)

	// Italic: *text* (but not inside **)
	reItalic = regexp.MustCompile(`(?:^|[^*])\*([^*]+)\*(?:[^*]|$)`)

	// reItalicSimple matches single-star delimited text for italic conversion.
	reItalicSimple = regexp.MustCompile(`\*([^*]+)\*`)

	// Heading: # ... or ## ... etc. (at start of line)
	reHeading = regexp.MustCompile(`(?m)^#{1,6}\s+(.+)$`)

	// List item: - item (at start of line)
	reListItem = regexp.MustCompile(`(?m)^- (.+)$`)
)

// EscapeHTML escapes special HTML characters for Telegram.
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// MarkdownToHTML converts GitHub-flavored Markdown to Telegram HTML subset.
//
// Supported conversions:
//
//	**bold**       -> <b>bold</b>
//	*italic*       -> <i>italic</i>
//	`code`         -> <code>code</code>
//	```code```     -> <pre>code</pre>  (with optional language hint stripped)
//	[text](url)    -> <a href="url">text</a>
//	# heading      -> <b>heading</b>
//	- list item    -> bullet list item
//	| table |      -> <pre> formatted </pre>
func MarkdownToHTML(md string) string {
	// 1. Extract code blocks first to protect them from other transformations.
	var codeBlocks []string
	result := reCodeBlock.ReplaceAllStringFunc(md, func(match string) string {
		sub := reCodeBlock.FindStringSubmatch(match)
		placeholder := fmt.Sprintf("\x00CODEBLOCK%d\x00", len(codeBlocks))
		codeBlocks = append(codeBlocks, "<pre>"+EscapeHTML(sub[1])+"</pre>")
		return placeholder
	})

	// 2. Extract inline code to protect from further transformations.
	var inlineCodes []string
	result = reInlineCode.ReplaceAllStringFunc(result, func(match string) string {
		sub := reInlineCode.FindStringSubmatch(match)
		placeholder := fmt.Sprintf("\x00INLINECODE%d\x00", len(inlineCodes))
		inlineCodes = append(inlineCodes, "<code>"+EscapeHTML(sub[1])+"</code>")
		return placeholder
	})

	// 3. Escape HTML entities in non-code text before applying transformations.
	result = EscapeHTML(result)

	// 4. Links.
	result = reLink.ReplaceAllString(result, `<a href="$2">$1</a>`)

	// 5. Headings (before bold, since headings use # not **).
	result = reHeading.ReplaceAllString(result, "<b>$1</b>")

	// 6. Bold (**text**) — must come before italic.
	result = reBold.ReplaceAllString(result, "<b>$1</b>")

	// 7. Italic (*text*) — single stars only.
	result = convertItalic(result)

	// 8. List items.
	result = reListItem.ReplaceAllString(result, "• $1")

	// 9. Restore inline code placeholders.
	for i, code := range inlineCodes {
		placeholder := fmt.Sprintf("\x00INLINECODE%d\x00", i)
		result = strings.ReplaceAll(result, placeholder, code)
	}

	// 10. Restore code block placeholders.
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("\x00CODEBLOCK%d\x00", i)
		result = strings.ReplaceAll(result, placeholder, block)
	}

	return result
}

// convertItalic handles *italic* conversion while avoiding **bold** stars.
// Uses the package-level reItalicSimple regex.
func convertItalic(s string) string {
	return reItalicSimple.ReplaceAllString(s, "<i>$1</i>")
}

// ChunkMessage splits text into chunks that fit within maxLen.
// Split priority: paragraph break (\n\n) > line break (\n) > hard cut.
// Code blocks (``` fences) are never split — if a split point falls inside
// a code block, the chunk boundary moves to before the block starts.
// If a code block is larger than maxLen, it is emitted as a single oversized chunk.
func ChunkMessage(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	remaining := text

	for len(remaining) > 0 {
		if len(remaining) <= maxLen {
			chunks = append(chunks, remaining)
			break
		}

		// If remaining starts with (or immediately leads into) a code block
		// that exceeds maxLen, emit the code block as one chunk.
		if cbEnd := codeBlockEnd(remaining); cbEnd > maxLen {
			chunks = append(chunks, remaining[:cbEnd])
			remaining = advancePastSeparator(remaining[cbEnd:])
			continue
		}

		cutPoint := findSplitPoint(remaining, maxLen)
		chunk := remaining[:cutPoint]
		chunks = append(chunks, chunk)

		remaining = advancePastSeparator(remaining[cutPoint:])
	}

	return chunks
}

// advancePastSeparator skips a leading paragraph break or line break.
func advancePastSeparator(s string) string {
	if strings.HasPrefix(s, "\n\n") {
		return s[2:]
	}
	if strings.HasPrefix(s, "\n") {
		return s[1:]
	}
	return s
}

// codeBlockEnd returns the end position of a code block that starts at the
// beginning of text. If text doesn't start with ```, returns 0.
func codeBlockEnd(text string) int {
	if !strings.HasPrefix(text, "```") {
		return 0
	}
	// Find the closing fence after the opening one.
	closeIdx := strings.Index(text[3:], "```")
	if closeIdx < 0 {
		// Unclosed code block — treat the whole remaining text as the block.
		return len(text)
	}
	end := 3 + closeIdx + 3 // past the closing ```
	return end
}

// findSplitPoint returns the best position to split text at, no greater than maxLen.
// It prefers paragraph breaks, then line breaks, then does a hard cut.
// It avoids splitting inside code blocks.
func findSplitPoint(text string, maxLen int) int {
	window := text[:maxLen]

	// Try paragraph break (\n\n) — find the last one within the window.
	if idx := strings.LastIndex(window, "\n\n"); idx > 0 {
		if !isInsideCodeBlock(text, idx) {
			return idx
		}
		// If inside a code block, move to before the code block.
		blockStart := findCodeBlockStart(text, idx)
		if blockStart > 0 {
			return blockStart
		}
	}

	// Try line break (\n).
	if idx := strings.LastIndex(window, "\n"); idx > 0 {
		if !isInsideCodeBlock(text, idx) {
			return idx
		}
		blockStart := findCodeBlockStart(text, idx)
		if blockStart > 0 {
			return blockStart
		}
	}

	// Hard cut at maxLen.
	return maxLen
}

// isInsideCodeBlock returns true if position pos in text is between an opening
// ``` fence and its closing ``` fence.
func isInsideCodeBlock(text string, pos int) bool {
	// Count ``` fences before pos. If odd, we're inside a code block.
	before := text[:pos]
	fences := countFences(before)
	return fences%2 != 0
}

// findCodeBlockStart finds the start of the code block that contains position pos.
// It searches backward from pos for the opening ``` fence.
// Returns the position just before the opening fence line, or 0 if not found.
func findCodeBlockStart(text string, pos int) int {
	before := text[:pos]
	// Find the last ``` that opens a block (odd fence count).
	idx := strings.LastIndex(before, "```")
	if idx < 0 {
		return 0
	}

	// Move to the start of the line containing the fence.
	lineStart := strings.LastIndex(before[:idx], "\n")
	if lineStart < 0 {
		return 0
	}

	// If there's a paragraph break before the code block, split there.
	parBreak := strings.LastIndex(before[:lineStart+1], "\n\n")
	if parBreak > 0 {
		return parBreak
	}

	// Otherwise split at the line break before the code block.
	if lineStart > 0 {
		return lineStart
	}

	return 0
}

// countFences counts the number of ``` occurrences in s.
func countFences(s string) int {
	count := 0
	for {
		idx := strings.Index(s, "```")
		if idx < 0 {
			break
		}
		count++
		s = s[idx+3:]
	}
	return count
}
