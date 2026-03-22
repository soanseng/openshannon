package safety

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/soanseng/openshannon/internal/config"
)

// FilterResult describes the outcome of a safety check.
type FilterResult struct {
	Allowed bool
	Reason  string // user-facing message
	Rule    string // internal rule that matched
}

// Filter validates prompts, shell commands, and paths against safety rules.
type Filter struct {
	blockedPrompts []*regexp.Regexp
	blockedShell   []*regexp.Regexp
	protectedPaths []*regexp.Regexp
	// rawPaths stores the original pattern strings for protected paths,
	// used by CheckPath for prefix-based directory matching.
	rawPaths []string
}

// NewFilter compiles the patterns from cfg and returns a ready Filter.
func NewFilter(cfg config.SafetyConfig) *Filter {
	f := &Filter{
		blockedPrompts: compilePatterns(cfg.BlockedPrompts),
		blockedShell:   compilePatterns(cfg.BlockedShell),
		protectedPaths: compilePatterns(cfg.ProtectedPaths),
		rawPaths:       cfg.ProtectedPaths,
	}
	return f
}

// compilePatterns compiles a slice of regex strings into regexp objects.
// Invalid patterns are skipped with a warning log.
func compilePatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			slog.Warn("invalid safety regex pattern, skipping",
				"pattern", p,
				"err", err,
			)
			continue
		}
		compiled = append(compiled, re)
	}
	return compiled
}

// CheckPrompt checks user prompt text against blocked prompt patterns
// and protected path patterns.
func (f *Filter) CheckPrompt(text string) FilterResult {
	// Check blocked prompt patterns.
	for _, re := range f.blockedPrompts {
		if re.MatchString(text) {
			return FilterResult{
				Allowed: false,
				Reason:  "prompt contains a blocked pattern",
				Rule:    fmt.Sprintf("blocked_prompt:%s", re.String()),
			}
		}
	}

	// Check protected path patterns against the prompt text.
	for _, re := range f.protectedPaths {
		if re.MatchString(text) {
			return FilterResult{
				Allowed: false,
				Reason:  "prompt references a protected path",
				Rule:    fmt.Sprintf("protected_path:%s", re.String()),
			}
		}
	}

	return FilterResult{Allowed: true}
}

// CheckShell checks a shell command against blocked shell patterns.
func (f *Filter) CheckShell(cmd string) FilterResult {
	for _, re := range f.blockedShell {
		if re.MatchString(cmd) {
			return FilterResult{
				Allowed: false,
				Reason:  "shell command is blocked by safety rules",
				Rule:    fmt.Sprintf("blocked_shell:%s", re.String()),
			}
		}
	}
	return FilterResult{Allowed: true}
}

// CheckPath checks whether a directory path is inside a protected area.
// It uses the raw protected path strings for prefix matching so that, for
// example, "~/.ssh" is blocked when "~/.ssh/authorized_keys" is protected.
func (f *Filter) CheckPath(path string) FilterResult {
	// Normalize: ensure path does not have a trailing slash for consistent
	// prefix comparison, but keep a version with trailing slash too.
	clean := strings.TrimRight(path, "/")
	withSlash := clean + "/"

	for _, raw := range f.rawPaths {
		// Strip the regex escape sequences for path comparison.
		// The raw strings are path-like (e.g., "/etc/", "~/.ssh/authorized_keys").
		literal := strings.ReplaceAll(raw, `\.`, ".")
		literalClean := strings.TrimRight(literal, "/")

		// Block if:
		// 1. The path exactly matches the protected path (without trailing slash).
		// 2. The path starts with the protected path (path is inside protected area).
		// 3. The protected path starts with the given path (path is a parent of
		//    something protected, e.g., "~/.ssh" contains "~/.ssh/authorized_keys").
		if clean == literalClean ||
			strings.HasPrefix(clean, literalClean+"/") ||
			strings.HasPrefix(withSlash, literal) ||
			strings.HasPrefix(literal, withSlash) {
			return FilterResult{
				Allowed: false,
				Reason:  "path is protected by safety rules",
				Rule:    fmt.Sprintf("protected_path:%s", raw),
			}
		}
	}
	return FilterResult{Allowed: true}
}
