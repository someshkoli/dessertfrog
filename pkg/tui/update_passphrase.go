package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/someshkoli/dessertfrog/pkg/encryption"
)

// handlePassphrasePromptKeys handles key presses in passphrase prompt mode
func (m Model) handlePassphrasePromptKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "ctrl+c":
		// Cancel passphrase prompt
		m.passphrasePromptMode = false
		m.passphraseInput = m.makePassphraseInput()
		m.passphraseKeyName = ""
		m.passphraseKeyPath = ""
		m.passphraseForNewKey = false // Reset flag
		// Continue without encryption
		m.connectionStatus = ConnectionFailed
		return m, tea.Quit

	case "enter":
		// Submit passphrase
		passphrase := m.passphraseInput.Value()
		if passphrase == "" {
			// Empty passphrase is allowed for new key generation
			if !m.passphraseForNewKey {
				// For existing keys, empty passphrase means cancel
				m.passphrasePromptMode = true
				m.passphraseInput = m.makePassphraseInput()
				m.connectionStatus = ConnectionFailed
				return m, connectToDatabase(m.driver)
			}
		}

		// Close prompt immediately
		m.passphrasePromptMode = false
		m.passphraseInput = m.makePassphraseInput()

		// Check if we're creating a new key
		if m.passphraseForNewKey {
			m.passphraseForNewKey = false // Reset flag
			// Generate new SSH key with passphrase
			return m, generateSSHKeyCmd(passphrase)
		}

		// Otherwise, save passphrase for existing key and continue
		keyPath := m.passphraseKeyPath
		return m, savePassphraseAndContinue(keyPath, passphrase)

	default:
		// Pass all other keys to the textinput
		var cmd tea.Cmd
		m.passphraseInput, cmd = m.passphraseInput.Update(msg)
		return m, cmd
	}
}

// savePassphraseAndContinue saves the passphrase to keychain and continues with connection
func savePassphraseAndContinue(keyPath, passphrase string) tea.Cmd {
	return func() tea.Msg {
		// Try to save passphrase to keychain
		var keychainErr error
		if err := encryption.SaveKeychainPassword(keyPath, passphrase); err != nil {
			// If saving fails, log but continue anyway
			// The passphrase will work for this session
			keychainErr = err
		}

		// Return message to continue with connection
		return passphraseSubmittedMsg{
			passphrase:    passphrase,
			keychainError: keychainErr,
		}
	}
}

// passphraseSubmittedMsg is sent when user submits a passphrase
type passphraseSubmittedMsg struct {
	passphrase    string
	keychainError error
}
