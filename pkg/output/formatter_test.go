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
