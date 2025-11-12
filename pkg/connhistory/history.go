package connhistory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ConnectionEntry represents a saved database connection
type ConnectionEntry struct {
	Driver    string    `json:"driver"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Database  string    `json:"database"`
	Schema    string    `json:"schema,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Signature returns a human-readable connection signature
// Format: driver://user@host:port/database/schema
func (c *ConnectionEntry) Signature() string {
	sig := fmt.Sprintf("%s://%s@%s:%d/%s",
		c.Driver, c.Username, c.Host, c.Port, c.Database)
	if c.Schema != "" {
		sig += "/" + c.Schema
	}
	return sig
}

// History manages connection history stored in cache directory
type History struct {
	entries  []ConnectionEntry
	filePath string
}

// NewHistory creates a new connection history manager
func NewHistory() (*History, error) {
	dir, err := historyDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine history directory: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	filePath := filepath.Join(dir, "connections.jsonl")

	h := &History{
		entries:  make([]ConnectionEntry, 0),
		filePath: filePath,
	}

	// Load existing history
	if err := h.load(); err != nil {
		return nil, err
	}

	return h, nil
}

// historyDir returns the connection history directory path
// Uses XDG_CACHE_HOME if set, otherwise ~/.cache/dessertfrog
func historyDir() (string, error) {
	if cacheHome := os.Getenv("XDG_CACHE_HOME"); cacheHome != "" {
		return filepath.Join(cacheHome, "dessertfrog"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".cache", "dessertfrog"), nil
}

// load reads connection history from the JSONL file
func (h *History) load() error {
	file, err := os.Open(h.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's fine
			return nil
		}
		return fmt.Errorf("failed to open connection history file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	entries := make([]ConnectionEntry, 0)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entry ConnectionEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Skip malformed lines
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading connection history: %w", err)
	}

	h.entries = entries
	return nil
}

// save writes connection history to the JSONL file
func (h *History) save() error {
	file, err := os.Create(h.filePath)
	if err != nil {
		return fmt.Errorf("failed to create connection history file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, entry := range h.entries {
		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		writer.WriteString(string(data) + "\n")
	}

	return nil
}

// Add adds a new connection to history
// If the connection already exists, it updates the timestamp and moves it to the front
func (h *History) Add(driver, host string, port int, username, password, database, schema string) error {
	newEntry := ConnectionEntry{
		Driver:    driver,
		Host:      host,
		Port:      port,
		Username:  username,
		Password:  password,
		Database:  database,
		Schema:    schema,
		Timestamp: time.Now(),
	}

	// Check if connection already exists
	existingIndex := -1
	for i, entry := range h.entries {
		if entry.Driver == driver &&
			entry.Host == host &&
			entry.Port == port &&
			entry.Username == username &&
			entry.Database == database &&
			entry.Schema == schema {
			existingIndex = i
			break
		}
	}

	// Remove existing entry if found
	if existingIndex >= 0 {
		h.entries = append(h.entries[:existingIndex], h.entries[existingIndex+1:]...)
	}

	// Prepend new entry (most recent first)
	h.entries = append([]ConnectionEntry{newEntry}, h.entries...)

	// Limit to 100 most recent connections
	if len(h.entries) > 100 {
		h.entries = h.entries[:100]
	}

	return h.save()
}

// GetAll returns all connection entries
func (h *History) GetAll() []ConnectionEntry {
	return h.entries
}

// Filter returns connections matching the query string
// Matches against the full signature (case-insensitive)
func (h *History) Filter(query string) []ConnectionEntry {
	if query == "" {
		return h.entries
	}

	query = strings.ToLower(query)
	filtered := make([]ConnectionEntry, 0)

	for _, entry := range h.entries {
		sig := strings.ToLower(entry.Signature())
		if strings.Contains(sig, query) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}
