package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatValueNil(t *testing.T) {
	result := FormatValue(nil)
	if result != "<none>" {
		t.Errorf("Expected '<none>', got '%s'", result)
	}
}

func TestFormatValueString(t *testing.T) {
	result := FormatValue("hello")
	if result != "hello" {
		t.Errorf("Expected 'hello', got '%s'", result)
	}
}

func TestFormatValueSlice(t *testing.T) {
	result := FormatValue([]interface{}{"a", "b", "c"})
	if result != "a,b,c" {
		t.Errorf("Expected 'a,b,c', got '%s'", result)
	}
}

func TestFormatValueMap(t *testing.T) {
	result := FormatValue(map[string]interface{}{"app": "nginx"})
	if result != "app=nginx" {
		t.Errorf("Expected 'app=nginx', got '%s'", result)
	}
}

func TestTruncate(t *testing.T) {
	short := truncate("hello", 10)
	if short != "hello" {
		t.Errorf("Expected 'hello', got '%s'", short)
	}

	long := truncate("this is a very long string that should be truncated", 20)
	if len(long) != 20 {
		t.Errorf("Expected length 20, got %d", len(long))
	}
	if !strings.HasSuffix(long, "...") {
		t.Error("Expected truncated string to end with '...'")
	}
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{format: FormatJSON, writer: &buf}

	results := []map[string]interface{}{
		{"name": "pod-1", "status": "Running"},
	}

	err := f.Print(results, []string{"name", "status"})
	if err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
	if len(parsed) != 1 {
		t.Errorf("Expected 1 result, got %d", len(parsed))
	}
}

func TestPrintCSV(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{format: FormatCSV, writer: &buf}

	results := []map[string]interface{}{
		{"name": "pod-1", "status": "Running"},
	}

	err := f.Print(results, []string{"name", "status"})
	if err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines (header + 1 row), got %d", len(lines))
	}
	if lines[0] != "name,status" {
		t.Errorf("Expected header 'name,status', got '%s'", lines[0])
	}
}

func TestPrintTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{format: FormatTable, writer: &buf}

	err := f.Print(nil, []string{"name"})
	if err != nil {
		t.Fatalf("Print failed: %v", err)
	}
	if !strings.Contains(buf.String(), "No resources found") {
		t.Error("Expected 'No resources found' message")
	}
}

func TestColorizeStatus(t *testing.T) {
	SetColorEnabled(true)
	defer SetColorEnabled(false)

	tests := []struct {
		value, field string
		wantColor    string
	}{
		{"Running", "status", colorGreen},
		{"Succeeded", "status", colorGreen},
		{"Pending", "status", colorYellow},
		{"Failed", "status", colorRed},
		{"CrashLoopBackOff", "status", colorRed},
		{"Normal", "type", colorGreen},
		{"Warning", "type", colorYellow},
		{"pod-1", "name", ""},     // no color for name field
		{"<none>", "status", ""},  // no color for <none>
	}

	for _, tt := range tests {
		result := colorize(tt.value, tt.field)
		if tt.wantColor == "" {
			if strings.Contains(result, "\033[") {
				t.Errorf("colorize(%q, %q) = %q, expected no ANSI codes", tt.value, tt.field, result)
			}
		} else {
			if !strings.HasPrefix(result, tt.wantColor) {
				t.Errorf("colorize(%q, %q) = %q, expected prefix %q", tt.value, tt.field, result, tt.wantColor)
			}
			if !strings.HasSuffix(result, colorReset) {
				t.Errorf("colorize(%q, %q) should end with reset code", tt.value, tt.field)
			}
		}
	}
}

func TestVisibleLen(t *testing.T) {
	plain := "Running"
	colored := colorGreen + "Running" + colorReset

	if visibleLen(plain) != 7 {
		t.Errorf("visibleLen(%q) = %d, want 7", plain, visibleLen(plain))
	}
	if visibleLen(colored) != 7 {
		t.Errorf("visibleLen(%q) = %d, want 7", colored, visibleLen(colored))
	}
}

func TestTruncateColored(t *testing.T) {
	// Short string — no truncation
	short := colorGreen + "Running" + colorReset
	result := truncateColored(short, 50)
	if visibleLen(result) != 7 {
		t.Errorf("truncateColored short: visibleLen = %d, want 7", visibleLen(result))
	}

	// Long colored string — should truncate by visible chars
	long := colorRed + "CrashLoopBackOff-very-long-status-text-here-extra" + colorReset
	result = truncateColored(long, 20)
	if visibleLen(result) != 20 {
		t.Errorf("truncateColored long: visibleLen = %d, want 20", visibleLen(result))
	}
	if !strings.HasSuffix(result, colorReset) {
		t.Error("truncateColored should end with reset code")
	}
}

func TestColorDisabled(t *testing.T) {
	SetColorEnabled(false)

	result := colorize("Running", "status")
	if strings.Contains(result, "\033[") {
		t.Error("colorize should not add ANSI codes when disabled")
	}
	if result != "Running" {
		t.Errorf("Expected 'Running', got %q", result)
	}
}

func TestPrintTable(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{format: FormatTable, writer: &buf}

	results := []map[string]interface{}{
		{"name": "pod-1", "status": "Running"},
		{"name": "pod-2", "status": "Pending"},
	}

	err := f.Print(results, []string{"name", "status"})
	if err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "NAME") {
		t.Error("Expected header 'NAME' in table output")
	}
	if !strings.Contains(output, "pod-1") {
		t.Error("Expected 'pod-1' in table output")
	}
	if !strings.Contains(output, "2 resource(s) found") {
		t.Error("Expected '2 resource(s) found' in table output")
	}
}
