package output

import (
	"regexp"
	"strings"
)

// Additional ANSI color codes for syntax highlighting
const (
	colorCyan    = "\033[36m"
	colorMagenta = "\033[35m"
)

// JSON highlighting patterns
var (
	jsonKeyRe    = regexp.MustCompile(`^(\s*)"([^"]+)":`)
	jsonStringRe = regexp.MustCompile(`:\s*"((?:[^"\\]|\\.)*)"`)
	jsonNumberRe = regexp.MustCompile(`:\s*(-?\d+\.?\d*(?:[eE][+-]?\d+)?)([,\s\n\r\]}]|$)`)
	jsonBoolRe   = regexp.MustCompile(`:\s*(true|false)`)
	jsonNullRe   = regexp.MustCompile(`:\s*(null)`)
)

// HighlightJSON applies syntax highlighting to a JSON string.
func HighlightJSON(s string) string {
	if !colorEnabled {
		return s
	}

	var out strings.Builder
	for _, line := range strings.Split(s, "\n") {
		highlighted := line

		// Color keys: "key":
		highlighted = jsonKeyRe.ReplaceAllString(highlighted, `${1}`+colorCyan+`"${2}":`+colorReset)

		// Color string values: : "value"
		highlighted = jsonStringRe.ReplaceAllString(highlighted, `: `+colorGreen+`"${1}"`+colorReset)

		// Color numbers: : 123
		highlighted = jsonNumberRe.ReplaceAllString(highlighted, `: `+colorYellow+`${1}`+colorReset+`${2}`)

		// Color booleans: : true/false
		highlighted = jsonBoolRe.ReplaceAllString(highlighted, `: `+colorMagenta+`${1}`+colorReset)

		// Color null
		highlighted = jsonNullRe.ReplaceAllString(highlighted, `: `+colorRed+`${1}`+colorReset)

		out.WriteString(highlighted)
		out.WriteByte('\n')
	}

	return strings.TrimRight(out.String(), "\n")
}

// HighlightYAML applies syntax highlighting to a YAML string.
func HighlightYAML(s string) string {
	if !colorEnabled {
		return s
	}

	var out strings.Builder
	for _, line := range strings.Split(s, "\n") {
		highlighted := highlightYAMLLine(line)
		out.WriteString(highlighted)
		out.WriteByte('\n')
	}

	return strings.TrimRight(out.String(), "\n")
}

func highlightYAMLLine(line string) string {
	// Skip empty lines and document separators
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || trimmed == "---" || trimmed == "..." {
		return line
	}

	// List items: "- key: value"
	if strings.HasPrefix(trimmed, "- ") {
		// Find indent + "- " prefix
		idx := strings.Index(line, "- ")
		prefix := line[:idx+2]
		rest := line[idx+2:]
		return prefix + highlightYAMLKeyValue(rest)
	}

	return highlightYAMLKeyValue(line)
}

func highlightYAMLKeyValue(line string) string {
	// Find "key: value" pattern
	colonIdx := strings.Index(line, ": ")
	if colonIdx < 0 {
		// Check for key-only line (e.g. "key:")
		if strings.HasSuffix(strings.TrimSpace(line), ":") {
			trimmed := strings.TrimRight(line, " ")
			keyEnd := len(trimmed) - 1
			// Preserve leading whitespace
			return colorCyan + trimmed[:keyEnd] + colorReset + ":"
		}
		return line
	}

	key := line[:colonIdx]
	value := line[colonIdx+2:]

	coloredKey := colorCyan + key + colorReset
	coloredValue := colorizeYAMLValue(value)

	return coloredKey + ": " + coloredValue
}

func colorizeYAMLValue(value string) string {
	trimmed := strings.TrimSpace(value)

	// Null
	if trimmed == "null" || trimmed == "~" {
		return colorRed + value + colorReset
	}

	// Booleans
	if trimmed == "true" || trimmed == "false" {
		return colorMagenta + value + colorReset
	}

	// Numbers (int or float)
	if isYAMLNumber(trimmed) {
		return colorYellow + value + colorReset
	}

	// Quoted strings
	if (strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`)) ||
		(strings.HasPrefix(trimmed, `'`) && strings.HasSuffix(trimmed, `'`)) {
		return colorGreen + value + colorReset
	}

	// Plain string values
	if trimmed != "" {
		return colorGreen + value + colorReset
	}

	return value
}

func isYAMLNumber(s string) bool {
	if s == "" {
		return false
	}
	// Allow leading minus
	start := 0
	if s[0] == '-' || s[0] == '+' {
		start = 1
		if len(s) == 1 {
			return false
		}
	}
	hasDot := false
	for i := start; i < len(s); i++ {
		if s[i] == '.' {
			if hasDot {
				return false
			}
			hasDot = true
		} else if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
