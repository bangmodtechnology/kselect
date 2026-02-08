package output

import (
	"os"
	"regexp"
	"strings"

	"golang.org/x/term"
)

// ANSI color codes
const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
)

var colorEnabled = true

// SetColorEnabled globally enables or disables color output.
func SetColorEnabled(enabled bool) {
	colorEnabled = enabled
}

// IsColorEnabled returns whether color output is enabled.
func IsColorEnabled() bool {
	return colorEnabled
}

// DetectColor returns true if stdout is a TTY and color should be used.
func DetectColor() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Status values grouped by color
var greenStatuses = map[string]bool{
	"running":   true,
	"succeeded": true,
	"active":    true,
	"bound":     true,
	"ready":     true,
	"complete":  true,
	"available": true,
}

var yellowStatuses = map[string]bool{
	"pending":      true,
	"terminating":  true,
	"unknown":      true,
	"creating":     true,
	"init":         true,
	"waiting":      true,
	"containercreating": true,
}

var redStatuses = map[string]bool{
	"failed":             true,
	"error":              true,
	"crashloopbackoff":   true,
	"imagepullbackoff":   true,
	"errimagepull":       true,
	"oomkilled":          true,
	"evicted":            true,
	"backoff":            true,
}

// colorize wraps a value with ANSI color codes based on field name and value.
func colorize(value, fieldName string) string {
	if !colorEnabled || value == "" || value == "<none>" {
		return value
	}

	lower := strings.ToLower(value)

	switch fieldName {
	case "status":
		if greenStatuses[lower] {
			return colorGreen + value + colorReset
		}
		if yellowStatuses[lower] {
			return colorYellow + value + colorReset
		}
		if redStatuses[lower] {
			return colorRed + value + colorReset
		}
	case "type":
		switch lower {
		case "normal":
			return colorGreen + value + colorReset
		case "warning":
			return colorYellow + value + colorReset
		}
	}

	return value
}

// ansiRegexp matches ANSI escape sequences.
var ansiRegexp = regexp.MustCompile(`\033\[[0-9;]*m`)

// visibleLen returns the length of a string excluding ANSI escape codes.
func visibleLen(s string) int {
	return len(ansiRegexp.ReplaceAllString(s, ""))
}

// truncateColored truncates a string by visible character count,
// preserving ANSI codes and ensuring proper reset.
func truncateColored(s string, max int) string {
	if visibleLen(s) <= max {
		return s
	}

	// Strip ANSI, truncate, then re-apply color
	plain := ansiRegexp.ReplaceAllString(s, "")
	if len(plain) <= max {
		return s
	}

	truncated := plain[:max-3] + "..."

	// Check if original had color — if so, re-wrap the truncated text
	if strings.Contains(s, "\033[") {
		// Find the first color code
		loc := ansiRegexp.FindString(s)
		if loc != "" {
			return loc + truncated + colorReset
		}
	}

	return truncated
}
