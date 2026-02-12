package encryption

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GenerateSSHKeyResult holds the result of SSH key generation
type GenerateSSHKeyResult struct {
	Key     *Key
	Success bool
	Error   error
}

// GenerateSSHKey creates a new Ed25519 SSH key using ssh-keygen
func GenerateSSHKey(passphrase string) (*Key, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	keyPath := filepath.Join(sshDir, "id_dessertfrog")

	// Check if key already exists
	if _, err := os.Stat(keyPath); err == nil {
		return &Key{
			Type: KeyTypeSSH,
			Path: keyPath,
			Name: "id_dessertfrog",
		}, nil
	}

	// Check if ssh-keygen is available
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		return nil, fmt.Errorf("ssh-keygen not found. Please install OpenSSH: %w", err)
	}

	// Generate Ed25519 key with ssh-keygen
	args := []string{
		"-t", "ed25519",
		"-f", keyPath,
		"-N", passphrase, // Passphrase (empty string for no passphrase)
		"-C", "dessertfrog-encryption-key",
		"-q", // Quiet mode
	}

	cmd := exec.Command("ssh-keygen", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ssh-keygen failed: %w\nOutput: %s", err, string(output))
	}

	// Verify the key was created
	if _, err := os.Stat(keyPath); err != nil {
		return nil, fmt.Errorf("key file not created: %w", err)
	}

	key := &Key{
		Type: KeyTypeSSH,
		Path: keyPath,
		Name: "id_dessertfrog",
	}

	// If passphrase was provided, save it to keychain
	if passphrase != "" {
		if err := SaveKeychainPassword(keyPath, passphrase); err != nil {
			// Non-fatal error, just means user will need to enter passphrase again
			fmt.Fprintf(os.Stderr, "Warning: Could not save passphrase to keychain: %v\n", err)
		}
	}

	return key, nil
}

// IsSSHKeygenAvailable checks if ssh-keygen command is available
func IsSSHKeygenAvailable() bool {
	_, err := exec.LookPath("ssh-keygen")
	return err == nil
}
