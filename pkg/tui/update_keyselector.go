package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/someshkoli/dessertfrog/pkg/encryption"
	"github.com/someshkoli/dessertfrog/pkg/helpers"
)

// handleKeySelectorKeys handles key presses in key selector mode
func (m Model) handleKeySelectorKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	// Handle escape key - different behavior based on mode
	if key == "esc" {
		if m.keySelectorInsertMode {
			// In insert mode, escape goes to normal mode
			m.keySelectorInsertMode = false
			return m, nil
		} else {
			// In normal mode, escape closes the popup
			m.keySelectorMode = false
			m.keySelectorFilter = ""
			m.keySelectorSelected = 0
			m.keySelectorScroll = 0
			m.keySelectorInsertMode = true // Reset for next time
			return m, nil
		}
	}

	// Handle 'i' key - enter insert mode from normal mode
	if key == "i" && !m.keySelectorInsertMode {
		m.keySelectorInsertMode = true
		return m, nil
	}

	// Handle 'g' key - generate SSH key when no keys available
	if key == "g" {
		if len(m.availableKeys) == 0 {
			// Trigger SSH key generation
			return m, generateSSHKeyCmd("")
		} else if !m.keySelectorInsertMode {
			// In normal mode with keys, 'g' means go to top
			m.keySelectorSelected = 0
			m.keySelectorScroll = 0
			return m, nil
		}
	}

	// In insert mode, handle text input
	if m.keySelectorInsertMode {
		return m.handleKeySelectorInsertMode(msg)
	}

	// In normal mode, handle navigation
	return m.handleKeySelectorNormalMode(msg)
}

// handleKeySelectorInsertMode handles keys when in insert mode
func (m Model) handleKeySelectorInsertMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "backspace":
		// Remove last character from filter
		if len(m.keySelectorFilter) > 0 {
			m.keySelectorFilter = m.keySelectorFilter[:len(m.keySelectorFilter)-1]
			// Update filtered keys
			m.filteredKeys = filterKeys(m.availableKeys, m.keySelectorFilter)
			// Reset selection
			m.keySelectorSelected = 0
			m.keySelectorScroll = 0
		}
		return m, nil

	case "enter":
		// Select key
		if len(m.filteredKeys) > 0 && m.keySelectorSelected < len(m.filteredKeys) {
			selectedKey := m.filteredKeys[m.keySelectorSelected]

			// Close key selector
			m.keySelectorMode = false
			m.keySelectorFilter = ""
			m.keySelectorInsertMode = true // Reset for next time

			// Save the selected key and continue with encryption setup
			return m, selectEncryptionKey(&selectedKey)
		}
		return m, nil

	default:
		// Add character to filter (if it's a printable character)
		if len(msg.String()) == 1 {
			m.keySelectorFilter += msg.String()
			// Update filtered keys
			m.filteredKeys = filterKeys(m.availableKeys, m.keySelectorFilter)
			// Reset selection
			m.keySelectorSelected = 0
			m.keySelectorScroll = 0
		}
		return m, nil
	}
}

// handleKeySelectorNormalMode handles navigation keys when in normal mode
func (m Model) handleKeySelectorNormalMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		// Navigate down
		if len(m.filteredKeys) > 0 {
			m.keySelectorSelected++
			if m.keySelectorSelected >= len(m.filteredKeys) {
				m.keySelectorSelected = len(m.filteredKeys) - 1
			}

			// Auto-scroll
			maxVisible := helpers.Min(20, m.height-16)
			if m.keySelectorSelected >= m.keySelectorScroll+maxVisible {
				m.keySelectorScroll = m.keySelectorSelected - maxVisible + 1
			}
		}
		return m, nil

	case "k", "up":
		// Navigate up
		if len(m.filteredKeys) > 0 {
			m.keySelectorSelected--
			if m.keySelectorSelected < 0 {
				m.keySelectorSelected = 0
			}

			// Auto-scroll
			if m.keySelectorSelected < m.keySelectorScroll {
				m.keySelectorScroll = m.keySelectorSelected
			}
		}
		return m, nil

	case "G":
		// Go to bottom
		if len(m.filteredKeys) > 0 {
			m.keySelectorSelected = len(m.filteredKeys) - 1
			maxVisible := helpers.Min(20, m.height-16)
			m.keySelectorScroll = helpers.Max(0, len(m.filteredKeys)-maxVisible)
		}
		return m, nil

	case "enter":
		// Select key
		if len(m.filteredKeys) > 0 && m.keySelectorSelected < len(m.filteredKeys) {
			selectedKey := m.filteredKeys[m.keySelectorSelected]

			// Close key selector
			m.keySelectorMode = false
			m.keySelectorFilter = ""
			m.keySelectorInsertMode = true // Reset for next time

			// Save the selected key and continue with encryption setup
			return m, selectEncryptionKey(&selectedKey)
		}
		return m, nil

	default:
		return m, nil
	}
}

// filterKeys filters keys based on query string
func filterKeys(keys []encryption.Key, query string) []encryption.Key {
	if query == "" {
		return keys
	}

	filtered := make([]encryption.Key, 0)
	for _, key := range keys {
		sig := key.Signature()
		if helpers.FuzzyMatch(query, sig) || helpers.FuzzyMatch(query, key.Name) {
			filtered = append(filtered, key)
		}
	}

	return filtered
}

// generateSSHKeyCmd triggers SSH key generation
func generateSSHKeyCmd(passphrase string) tea.Cmd {
	return func() tea.Msg {
		key, err := encryption.GenerateSSHKey(passphrase)
		if err != nil {
			return sshKeyGenerationMsg{
				success: false,
				err:     err,
			}
		}

		return sshKeyGenerationMsg{
			success: true,
			key:     key,
		}
	}
}

// selectEncryptionKey saves the selected encryption key
func selectEncryptionKey(key *encryption.Key) tea.Cmd {
	return func() tea.Msg {
		// Save encryption config
		config := &encryption.Config{
			KeyPath: key.Path,
			KeyType: key.Type,
		}

		if err := encryption.SaveConfig(config); err != nil {
			return encryptionSetupErrorMsg{err: err}
		}

		return encryptionKeySelectedMsg{key: key}
	}
}

// Message types for encryption flow
type sshKeyGenerationMsg struct {
	success bool
	key     *encryption.Key
	err     error
}

type encryptionKeySelectedMsg struct {
	key *encryption.Key
}

type encryptionSetupRequiredMsg struct{}

type encryptionSetupCompleteMsg struct {
	key *encryption.Key
}

type encryptionSetupErrorMsg struct {
	err error
}
