package encryption

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// KeyType represents the type of encryption key
type KeyType string

const (
	KeyTypeSSH KeyType = "ssh"
	KeyTypeGPG KeyType = "gpg"
)

// Key represents an encryption key
type Key struct {
	Type        KeyType
	Path        string
	Name        string
	Fingerprint string
}

// Signature returns a human-readable signature for the key
func (k *Key) Signature() string {
	return fmt.Sprintf("[%s] %s", k.Type, k.Name)
}

// DiscoverKeys finds SSH and GPG keys on the system
func DiscoverKeys() ([]Key, error) {
	keys := make([]Key, 0)

	// Discover SSH keys
	sshKeys, err := discoverSSHKeys()
	if err == nil {
		keys = append(keys, sshKeys...)
	}

	// Discover GPG keys
	gpgKeys, err := discoverGPGKeys()
	if err == nil {
		keys = append(keys, gpgKeys...)
	}

	return keys, nil
}

// discoverSSHKeys finds SSH private keys in ~/.ssh
func discoverSSHKeys() ([]Key, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		return []Key{}, nil
	}

	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return nil, err
	}

	keys := make([]Key, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Look for private keys (not .pub files, not known_hosts, config, etc.)
		if strings.HasSuffix(name, ".pub") ||
		   name == "known_hosts" ||
		   name == "config" ||
		   name == "authorized_keys" {
			continue
		}

		// Common SSH key patterns
		if strings.HasPrefix(name, "id_") || name == "identity" {
			path := filepath.Join(sshDir, name)
			keys = append(keys, Key{
				Type: KeyTypeSSH,
				Path: path,
				Name: name,
			})
		}
	}

	return keys, nil
}

// discoverGPGKeys finds GPG keys in ~/.gnupg
func discoverGPGKeys() ([]Key, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	gnupgDir := filepath.Join(homeDir, ".gnupg")
	if _, err := os.Stat(gnupgDir); os.IsNotExist(err) {
		return []Key{}, nil
	}

	// Check for GPG keyring files
	keys := make([]Key, 0)

	// Look for secring.gpg (older GPG versions)
	secringPath := filepath.Join(gnupgDir, "secring.gpg")
	if _, err := os.Stat(secringPath); err == nil {
		keys = append(keys, Key{
			Type: KeyTypeGPG,
			Path: secringPath,
			Name: "secring.gpg",
		})
	}

	// Look for pubring.kbx (newer GPG versions)
	pubringPath := filepath.Join(gnupgDir, "pubring.kbx")
	if _, err := os.Stat(pubringPath); err == nil {
		keys = append(keys, Key{
			Type: KeyTypeGPG,
			Path: pubringPath,
			Name: "pubring.kbx",
		})
	}

	return keys, nil
}
