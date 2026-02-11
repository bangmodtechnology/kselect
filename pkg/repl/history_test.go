package repl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHistory(t *testing.T) {
	// Create temporary history file
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, ".kselect_history")

	// Create history with custom file path
	h := &History{
		entries:  []HistoryEntry{},
		filePath: historyFile,
	}

	// Test: Add entries
	h.Add("name FROM pod")
	h.Add("name,status FROM deployment")
	h.Add("name FROM service WHERE namespace=default")

	if len(h.entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(h.entries))
	}

	// Test: Don't add empty queries
	h.Add("")
	if len(h.entries) != 3 {
		t.Errorf("Empty query should not be added, got %d entries", len(h.entries))
	}

	// Test: Don't add duplicate of last entry
	h.Add("name FROM service WHERE namespace=default")
	if len(h.entries) != 3 {
		t.Errorf("Duplicate of last entry should not be added, got %d entries", len(h.entries))
	}

	// Test: GetAll
	all := h.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 entries from GetAll, got %d", len(all))
	}
	if all[0] != "name FROM pod" {
		t.Errorf("Expected first entry 'name FROM pod', got '%s'", all[0])
	}

	// Test: GetLast
	last2 := h.GetLast(2)
	if len(last2) != 2 {
		t.Errorf("Expected 2 entries from GetLast(2), got %d", len(last2))
	}
	if last2[0] != "name,status FROM deployment" {
		t.Errorf("Expected 'name,status FROM deployment', got '%s'", last2[0])
	}

	// Test: GetLast with n > len(entries)
	last10 := h.GetLast(10)
	if len(last10) != 3 {
		t.Errorf("Expected 3 entries from GetLast(10), got %d", len(last10))
	}

	// Test: Save
	if err := h.save(); err != nil {
		t.Errorf("Failed to save history: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		t.Errorf("History file was not created")
	}

	// Test: Load
	h2 := &History{
		entries:  []HistoryEntry{},
		filePath: historyFile,
	}
	if err := h2.load(); err != nil {
		t.Errorf("Failed to load history: %v", err)
	}

	if len(h2.entries) != 3 {
		t.Errorf("Expected 3 entries after load, got %d", len(h2.entries))
	}

	// Test: Clear
	if err := h2.Clear(); err != nil {
		t.Errorf("Failed to clear history: %v", err)
	}

	if len(h2.entries) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(h2.entries))
	}

	// Test: Max history size
	h3 := &History{
		entries:  []HistoryEntry{},
		filePath: historyFile,
	}

	// Add more than max entries
	for i := 0; i < maxHistorySize+100; i++ {
		h3.Add("query " + string(rune(i)))
	}

	if len(h3.entries) > maxHistorySize {
		t.Errorf("Expected max %d entries, got %d", maxHistorySize, len(h3.entries))
	}
}
