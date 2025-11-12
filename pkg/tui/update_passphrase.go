package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/someshkoli/dessertfrog/pkg/encryption"
	"github.com/someshkoli/dessertfrog/pkg/helpers"
)

// handlePassphrasePromptKeys handles key presses in passphrase prompt mode
func (m Model) handlePassphrasePromptKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "ctrl+c":
		// Cancel passphrase prompt
		m.passphrasePromptMode = false
		m.passphraseInput = ""
		m.passphraseCursor = 0
		m.passphraseKeyName = ""
		m.passphraseKeyPath = ""
		// Continue without encryption
		m.connectionStatus = ConnectionFailed
		return m, tea.Quit

	case "enter":
		// Submit passphrase
		if m.passphraseInput == "" {
			// Empty passphrase - treat as cancel
			m.passphrasePromptMode = true
			m.passphraseInput = ""
			m.passphraseCursor = 0
			m.connectionStatus = ConnectionFailed
			return m, connectToDatabase(m.driver)
		}

		// Save passphrase to keychain
		passphrase := m.passphraseInput
		keyPath := m.passphraseKeyPath

		// Close prompt immediately
		m.passphrasePromptMode = false
		m.passphraseInput = ""
		m.passphraseCursor = 0

		// Save passphrase and continue
		return m, savePassphraseAndContinue(keyPath, passphrase)

	case "left":
		// Move cursor left
		if m.passphraseCursor > 0 {
			m.passphraseCursor--
		}
		return m, nil

	case "right":
		// Move cursor right
		runes := []rune(m.passphraseInput)
		if m.passphraseCursor < len(runes) {
			m.passphraseCursor++
		}
		return m, nil

	case "home", "ctrl+a":
		// Move cursor to start
		m.passphraseCursor = 0
		return m, nil

	case "end", "ctrl+e":
		// Move cursor to end
		runes := []rune(m.passphraseInput)
		m.passphraseCursor = len(runes)
		return m, nil

	case "backspace":
		// Delete character before cursor
		if m.passphraseCursor > 0 {
			newInput, newCursor := helpers.DeleteRune(m.passphraseInput, m.passphraseCursor-1)
			m.passphraseInput = newInput
			m.passphraseCursor = newCursor
		}
		return m, nil

	case "delete", "ctrl+d":
		// Delete character at cursor
		runes := []rune(m.passphraseInput)
		if m.passphraseCursor < len(runes) {
			newInput, _ := helpers.DeleteRune(m.passphraseInput, m.passphraseCursor)
			m.passphraseInput = newInput
		}
		return m, nil

	case "ctrl+u":
		// Delete from cursor to start
		runes := []rune(m.passphraseInput)
		m.passphraseInput = string(runes[m.passphraseCursor:])
		m.passphraseCursor = 0
		return m, nil

	case "ctrl+k":
		// Delete from cursor to end
		runes := []rune(m.passphraseInput)
		m.passphraseInput = string(runes[:m.passphraseCursor])
		return m, nil

	default:
		// Add character to input (if it's a printable character)
		if len(key) == 1 {
			m.passphraseInput = helpers.InsertRune(m.passphraseInput, rune(key[0]), m.passphraseCursor)
			m.passphraseCursor++
		}
		return m, nil
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
