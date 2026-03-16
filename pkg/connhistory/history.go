package connhistory

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/someshkoli/dessertfrog/pkg/encryption"
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
	SSLMode   string    `json:"ssl_mode,omitempty"`
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
	entries       []ConnectionEntry
	filePath      string
	encryptionKey *encryption.Key
	passphrase    string // Cached passphrase for this session
}

// NewHistory creates a new connection history manager
func NewHistory() (*History, error) {
	return NewHistoryWithEncryption(nil)
}

// NewHistoryWithEncryption creates a new connection history manager with optional encryption
func NewHistoryWithEncryption(key *encryption.Key) (*History, error) {
	return NewHistoryWithEncryptionAndPassphrase(key, "")
}

// NewHistoryWithEncryptionAndPassphrase creates a new connection history manager with encryption and passphrase
func NewHistoryWithEncryptionAndPassphrase(key *encryption.Key, passphrase string) (*History, error) {
	dir, err := historyDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine history directory: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	filePath := filepath.Join(dir, "connections.jsonl")

	h := &History{
		entries:       make([]ConnectionEntry, 0),
		filePath:      filePath,
		encryptionKey: key,
		passphrase:    passphrase,
	}

	// If encryption key is set, test if it needs a passphrase
	if key != nil && passphrase == "" {
		if err := encryption.TestKey(key); err != nil {
			return nil, err
			// // Check if it's a passphrase error
			// if errors.Is(err, encryption.ErrPassphraseRequired) {
			// 	return nil, err
			// }
			// return nil, err
			// Other error - log but continue
		}
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

	// Read all data
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(file); err != nil {
		return fmt.Errorf("failed to read connection history: %w", err)
	}

	data := buf.Bytes()

	// If encryption is enabled, decrypt the data
	if h.encryptionKey != nil && len(data) > 0 {
		var decrypted []byte
		var err error

		// Try with passphrase if we have one cached
		if h.passphrase != "" {
			decrypted, err = encryption.DecryptWithPassphrase(data, h.encryptionKey, h.passphrase)
		} else {
			decrypted, err = encryption.Decrypt(data, h.encryptionKey)
		}

		if err != nil {
			// Check if it's a passphrase error
			if errors.Is(err, encryption.ErrPassphraseRequired) {
				// Re-throw passphrase error so caller can handle it
				return err
			}

			// If decryption fails, try loading as plaintext (migration case)
			if err := h.loadPlaintext(data); err != nil {
				return fmt.Errorf("failed to decrypt connection history: %w", err)
			}
			// Successfully loaded plaintext, save as encrypted
			return h.save()
		}
		data = decrypted
	}

	// Parse JSONL data
	scanner := bufio.NewScanner(bytes.NewReader(data))
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

// loadPlaintext loads plaintext JSONL data (used for migration)
func (h *History) loadPlaintext(data []byte) error {
	scanner := bufio.NewScanner(bytes.NewReader(data))
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
		return fmt.Errorf("error reading plaintext connection history: %w", err)
	}

	h.entries = entries
	return nil
}

// SetPassphrase sets the passphrase for encrypted operations
func (h *History) SetPassphrase(passphrase string) {
	h.passphrase = passphrase
}

// save writes connection history to the JSONL file
func (h *History) save() error {
	// Build JSONL data in memory
	var buf bytes.Buffer
	for _, entry := range h.entries {
		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		buf.WriteString(string(data) + "\n")
	}

	data := buf.Bytes()

	// If encryption is enabled, encrypt the data
	if h.encryptionKey != nil {
		var encrypted []byte
		var err error

		// Use passphrase if we have one cached
		if h.passphrase != "" {
			encrypted, err = encryption.EncryptWithPassphrase(data, h.encryptionKey, h.passphrase)
		} else {
			encrypted, err = encryption.Encrypt(data, h.encryptionKey)
		}

		if err != nil {
			return fmt.Errorf("failed to encrypt connection history: %w", err)
		}
		data = encrypted
	}

	// Write to file
	if err := os.WriteFile(h.filePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write connection history file: %w", err)
	}

	return nil
}

// Add adds a new connection to history
// If the connection already exists, it updates the timestamp and moves it to the front
func (h *History) Add(driver, host string, port int, username, password, database, schema, sslMode string) error {
	newEntry := ConnectionEntry{
		Driver:    driver,
		Host:      host,
		Port:      port,
		Username:  username,
		Password:  password,
		Database:  database,
		Schema:    schema,
		SSLMode:   sslMode,
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
