package repl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	historyFileName = ".kselect_history"
	maxHistorySize  = 1000
)

// HistoryEntry represents a single history entry
type HistoryEntry struct {
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
}

// History manages query history
type History struct {
	entries  []HistoryEntry
	filePath string
}

// NewHistory creates a new History instance
func NewHistory() (*History, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	filePath := filepath.Join(homeDir, historyFileName)

	h := &History{
		entries:  []HistoryEntry{},
		filePath: filePath,
	}

	// Load existing history
	if err := h.load(); err != nil {
		// If file doesn't exist, that's okay
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load history: %w", err)
		}
	}

	return h, nil
}

// Add adds a query to history
func (h *History) Add(query string) {
	// Don't add empty queries or duplicates of the last entry
	if query == "" {
		return
	}
	if len(h.entries) > 0 && h.entries[len(h.entries)-1].Query == query {
		return
	}

	entry := HistoryEntry{
		Query:     query,
		Timestamp: time.Now(),
	}

	h.entries = append(h.entries, entry)

	// Trim history if too large
	if len(h.entries) > maxHistorySize {
		h.entries = h.entries[len(h.entries)-maxHistorySize:]
	}

	// Save to file
	h.save()
}

// GetAll returns all history entries as strings
func (h *History) GetAll() []string {
	result := make([]string, len(h.entries))
	for i, entry := range h.entries {
		result[i] = entry.Query
	}
	return result
}

// GetLast returns the last n queries
func (h *History) GetLast(n int) []string {
	if n <= 0 {
		return []string{}
	}

	start := len(h.entries) - n
	if start < 0 {
		start = 0
	}

	result := make([]string, len(h.entries)-start)
	for i, entry := range h.entries[start:] {
		result[i] = entry.Query
	}
	return result
}

// Clear clears all history
func (h *History) Clear() error {
	h.entries = []HistoryEntry{}
	return h.save()
}

// load loads history from file
func (h *History) load() error {
	data, err := os.ReadFile(h.filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &h.entries); err != nil {
		return fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return nil
}

// save saves history to file
func (h *History) save() error {
	data, err := json.Marshal(h.entries)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(h.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}
