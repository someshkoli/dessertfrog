package sqlhistory

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/someshkoli/dessertfrog/pkg/helpers"
)

// ConnectionSignature generates a unique signature for a database connection
func ConnectionSignature(driver, host string, port int, database, schema, user string) string {
	// Create a consistent string representation of the connection
	sigStr := fmt.Sprintf("%s://%s@%s:%d/%s/%s", driver, user, host, port, database, schema)

	// Hash to create a short, filesystem-safe identifier
	hash := sha256.Sum256([]byte(sigStr))
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes (16 hex chars)
}

// HistoryEntry represents a single SQL query in history
type HistoryEntry struct {
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
}

// History manages SQL query history for a specific connection using WAL approach
type History struct {
	entries   []HistoryEntry
	maxSize   int
	signature string
	filePath  string
}

// HistoryDir returns the directory where history files are stored
// Prefers XDG_CACHE_HOME or ~/.cache/dessertfrog, falls back to ~/.dessertfrog
func HistoryDir() (string, error) {
	var baseDir string

	// Try XDG_CACHE_HOME first
	if cacheHome := os.Getenv("XDG_CACHE_HOME"); cacheHome != "" {
		baseDir = filepath.Join(cacheHome, "dessertfrog")
	} else {
		// Try ~/.cache/dessertfrog
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}

		cacheDir := filepath.Join(homeDir, ".cache", "dessertfrog")
		// If .cache exists, use it; otherwise fall back to ~/.dessertfrog
		if _, err := os.Stat(filepath.Join(homeDir, ".cache")); err == nil {
			baseDir = cacheDir
		} else {
			baseDir = filepath.Join(homeDir, ".dessertfrog")
		}
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create history directory: %w", err)
	}

	return baseDir, nil
}

// NewHistory creates a new History instance for a specific connection
func NewHistory(driver, host string, port int, database, schema, user string, maxSize int) (*History, error) {
	if maxSize <= 0 {
		maxSize = 1000 // Default maximum history size
	}

	signature := ConnectionSignature(driver, host, port, database, schema, user)

	histDir, err := HistoryDir()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(histDir, fmt.Sprintf("sql_history_%s.jsonl", signature))

	h := &History{
		entries:   make([]HistoryEntry, 0),
		maxSize:   maxSize,
		signature: signature,
		filePath:  filePath,
	}

	// Load existing history from WAL file
	if err := h.load(); err != nil {
		// If file doesn't exist, that's fine - we'll create it on first append
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load history: %w", err)
		}
	}

	return h, nil
}

// load reads all entries from the WAL file
func (h *History) load() error {
	file, err := os.Open(h.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer size for long queries
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // Max 1MB per line

	entries := make([]HistoryEntry, 0)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry HistoryEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// Skip malformed lines
			continue
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading history file: %w", err)
	}

	h.entries = entries
	return nil
}

// append writes a single entry to the WAL file
func (h *History) append(entry HistoryEntry) error {
	// Ensure directory exists
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	// Open file in append mode
	file, err := os.OpenFile(h.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer file.Close()

	// Marshal entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	// Write entry with newline
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

// compact rewrites the WAL file with only the most recent maxSize entries
func (h *History) compact() error {
	// Trim entries in memory
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}

	// Write to temporary file
	tmpPath := h.filePath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	writer := bufio.NewWriter(file)
	for _, entry := range h.entries {
		data, err := json.Marshal(entry)
		if err != nil {
			file.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("failed to marshal entry: %w", err)
		}
		if _, err := writer.Write(data); err != nil {
			file.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("failed to write entry: %w", err)
		}
		if _, err := writer.WriteString("\n"); err != nil {
			file.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		file.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	if err := file.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic replace
	if err := os.Rename(tmpPath, h.filePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// Add adds a new query to history (appends to WAL)
func (h *History) Add(query string) error {
	// Trim whitespace
	query = helpers.Trim(query)

	// Don't add empty queries
	if query == "" {
		return nil
	}

	// Don't add duplicate consecutive queries
	if len(h.entries) > 0 && h.entries[len(h.entries)-1].Query == query {
		return nil
	}

	// Create new entry
	entry := HistoryEntry{
		Query:     query,
		Timestamp: time.Now(),
	}

	// Append to WAL file
	if err := h.append(entry); err != nil {
		return err
	}

	// Add to in-memory list
	h.entries = append(h.entries, entry)

	// Compact if we exceed maxSize by a significant margin (e.g., 20%)
	if len(h.entries) > h.maxSize+h.maxSize/5 {
		return h.compact()
	}

	return nil
}

// GetRecent returns the most recent N queries (in reverse chronological order)
func (h *History) GetRecent(n int) []string {
	if h == nil {
		return []string{}
	}

	if n <= 0 {
		n = len(h.entries)
	}

	start := len(h.entries) - n
	if start < 0 {
		start = 0
	}

	queries := make([]string, 0, len(h.entries)-start)
	// Return in reverse order (most recent first)
	for i := len(h.entries) - 1; i >= start; i-- {
		queries = append(queries, h.entries[i].Query)
	}

	return queries
}

// Search returns queries matching the given search term using fuzzy matching
func (h *History) Search(searchTerm string) []string {
	if searchTerm == "" {
		return h.GetRecent(50) // Return last 50 if no search term
	}

	matches := make([]string, 0)

	// Search in reverse order (most recent first) using fuzzy matching
	for i := len(h.entries) - 1; i >= 0; i-- {
		if helpers.FuzzyMatch(searchTerm, h.entries[i].Query) {
			matches = append(matches, h.entries[i].Query)
			if len(matches) >= 50 { // Limit to 50 matches
				break
			}
		}
	}

	return matches
}

// SearchEntries returns full history entries matching the given search term using fuzzy matching
func (h *History) SearchEntries(searchTerm string) []HistoryEntry {
	if h == nil {
		return []HistoryEntry{}
	}

	if searchTerm == "" {
		// Return last 50 entries with timestamps
		n := 50
		if n > len(h.entries) {
			n = len(h.entries)
		}

		start := len(h.entries) - n
		if start < 0 {
			start = 0
		}

		entries := make([]HistoryEntry, 0, len(h.entries)-start)
		// Return in reverse order (most recent first)
		for i := len(h.entries) - 1; i >= start; i-- {
			entries = append(entries, h.entries[i])
		}

		return entries
	}

	matches := make([]HistoryEntry, 0)

	// Search in reverse order (most recent first) using fuzzy matching
	for i := len(h.entries) - 1; i >= 0; i-- {
		if helpers.FuzzyMatch(searchTerm, h.entries[i].Query) {
			matches = append(matches, h.entries[i])
			if len(matches) >= 50 { // Limit to 50 matches
				break
			}
		}
	}

	return matches
}

// Clear removes all history entries
func (h *History) Clear() error {
	h.entries = make([]HistoryEntry, 0)
	// Remove the file
	if err := os.Remove(h.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove history file: %w", err)
	}
	return nil
}
