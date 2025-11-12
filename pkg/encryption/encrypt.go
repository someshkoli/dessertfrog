package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/crypto/ssh"
)

// ErrPassphraseRequired is returned when an SSH key needs a passphrase
var ErrPassphraseRequired = errors.New("SSH key requires passphrase")

// Config holds encryption configuration
type Config struct {
	KeyPath string
	KeyType KeyType
}

// ConfigPath returns the path to the encryption config file
func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".config", "dessertfrog")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "encryption.conf"), nil
}

// LoadConfig loads the encryption configuration
func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No config yet
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	config := &Config{}

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "key_path":
			config.KeyPath = value
		case "key_type":
			config.KeyType = KeyType(value)
		}
	}

	if config.KeyPath == "" {
		return nil, nil
	}

	return config, nil
}

// SaveConfig saves the encryption configuration
func SaveConfig(config *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	content := fmt.Sprintf("key_path=%s\nkey_type=%s\n", config.KeyPath, config.KeyType)
	return os.WriteFile(path, []byte(content), 0600)
}

// Encrypt encrypts data using the specified key
func Encrypt(data []byte, key *Key) ([]byte, error) {
	return EncryptWithPassphrase(data, key, "")
}

// EncryptWithPassphrase encrypts data using the specified key and optional passphrase
func EncryptWithPassphrase(data []byte, key *Key, passphrase string) ([]byte, error) {
	// Derive encryption key from SSH/GPG key
	encKey, err := deriveEncryptionKeyWithPassphrase(key, passphrase)
	if err != nil {
		return nil, err
	}

	return encryptAES(data, encKey)
}

// Decrypt decrypts data using the specified key
func Decrypt(data []byte, key *Key) ([]byte, error) {
	return DecryptWithPassphrase(data, key, "")
}

// DecryptWithPassphrase decrypts data using the specified key and optional passphrase
func DecryptWithPassphrase(data []byte, key *Key, passphrase string) ([]byte, error) {
	// Derive encryption key from SSH/GPG key
	encKey, err := deriveEncryptionKeyWithPassphrase(key, passphrase)
	if err != nil {
		return nil, err
	}

	return decryptAES(data, encKey)
}

// TestKey tests if a key can be used without a passphrase
func TestKey(key *Key) error {
	_, err := deriveEncryptionKey(key)
	return err
}

// deriveEncryptionKey derives an AES key from an SSH/GPG key
func deriveEncryptionKey(key *Key) ([]byte, error) {
	return deriveEncryptionKeyWithPassphrase(key, "")
}

// deriveEncryptionKeyWithPassphrase derives an AES key from an SSH/GPG key with optional passphrase
func deriveEncryptionKeyWithPassphrase(key *Key, passphrase string) ([]byte, error) {
	switch key.Type {
	case KeyTypeSSH:
		return deriveFromSSHKeyWithPassphrase(key.Path, passphrase)
	case KeyTypeGPG:
		return deriveFromGPGKey(key.Path)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", key.Type)
	}
}

// deriveFromSSHKey reads an SSH private key and derives an encryption key
func deriveFromSSHKey(keyPath string) ([]byte, error) {
	return deriveFromSSHKeyWithPassphrase(keyPath, "")
}

// deriveFromSSHKeyWithPassphrase reads an SSH private key and derives an encryption key with optional passphrase
func deriveFromSSHKeyWithPassphrase(keyPath, passphrase string) ([]byte, error) {
	// Read the private key file
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSH key: %w", err)
	}

	// Try to parse the key (may require passphrase)
	// First try without passphrase
	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		// If it fails, it might need a passphrase
		// Try the provided passphrase first
		if passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(passphrase))
			if err != nil {
				return nil, fmt.Errorf("invalid passphrase: %w", err)
			}
		} else {
			// Try to get it from the OS keychain
			keychainPass, err := getKeychainPassword(keyPath)
			if err != nil {
				// Passphrase required but not available
				return nil, ErrPassphraseRequired
			}

			signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(keychainPass))
			if err != nil {
				// Keychain passphrase is wrong, need user input
				return nil, ErrPassphraseRequired
			}
			// Use keychain passphrase for derivation
			passphrase = keychainPass
		}
	}

	// Derive a key from the SSH key data + passphrase
	// This ensures different passphrases produce different encryption keys
	var derivationData []byte
	if passphrase != "" {
		// Combine key data and passphrase for derivation
		derivationData = append(keyData, []byte(passphrase)...)
	} else {
		// No passphrase - just use key data
		derivationData = keyData
	}

	// Also include the public key to ensure we're using the actual key material
	if signer != nil {
		pubKeyBytes := signer.PublicKey().Marshal()
		derivationData = append(derivationData, pubKeyBytes...)
	}

	hash := sha256.Sum256(derivationData)
	return hash[:], nil
}

// deriveFromGPGKey derives an encryption key from a GPG keyring
func deriveFromGPGKey(keyPath string) ([]byte, error) {
	// For GPG, we'll use the keyring file data to derive a key
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read GPG key: %w", err)
	}

	hash := sha256.Sum256(keyData)
	return hash[:], nil
}

// getKeychainPassword retrieves a password from the OS keychain
func getKeychainPassword(keyPath string) (string, error) {
	service := "dessertfrog"
	account := base64.StdEncoding.EncodeToString([]byte(keyPath))

	switch runtime.GOOS {
	case "darwin":
		// macOS Keychain
		cmd := exec.Command("security", "find-generic-password", "-s", service, "-a", account, "-w")
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("password not found in keychain")
		}
		return strings.TrimSpace(string(output)), nil

	case "linux":
		// Try secret-tool (part of libsecret)
		cmd := exec.Command("secret-tool", "lookup", "service", service, "account", account)
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("password not found in keychain")
		}
		return strings.TrimSpace(string(output)), nil

	default:
		return "", fmt.Errorf("keychain not supported on %s", runtime.GOOS)
	}
}

// SaveKeychainPassword saves a password to the OS keychain
func SaveKeychainPassword(keyPath, password string) error {
	service := "dessertfrog"
	account := base64.StdEncoding.EncodeToString([]byte(keyPath))

	switch runtime.GOOS {
	case "darwin":
		// macOS Keychain
		cmd := exec.Command("security", "add-generic-password", "-s", service, "-a", account, "-w", password, "-U")
		return cmd.Run()

	case "linux":
		// Try secret-tool (part of libsecret)
		cmd := exec.Command("secret-tool", "store", "--label", "Dessertfrog SSH Key", "service", service, "account", account)
		cmd.Stdin = strings.NewReader(password)
		return cmd.Run()

	default:
		return fmt.Errorf("keychain not supported on %s", runtime.GOOS)
	}
}

// encryptAES encrypts data using AES-GCM
func encryptAES(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// decryptAES decrypts data using AES-GCM
func decryptAES(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
