package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// renderPassphrasePromptPopup renders the passphrase prompt popup overlay
func (m Model) renderPassphrasePromptPopup(baseView string) string {
	// Popup dimensions
	popupWidth := 70
	popupHeight := 10

	// Title - different for new key creation vs unlocking existing key
	var title string
	var keyInfo string
	if m.passphraseForNewKey {
		title = m.styles.PassphraseTitleStyle.Render("Create New SSH Key")
		keyInfo = m.styles.PassphraseKeyInfoStyle.Render("Key: " + m.passphraseKeyName + " (leave empty for no passphrase)")
	} else {
		title = m.styles.PassphraseTitleStyle.Render("Enter SSH Key Passphrase")
		keyInfo = m.styles.PassphraseKeyInfoStyle.Render("Key: " + m.passphraseKeyName)
	}

	// Input line with prompt
	inputPrompt := "Passphrase: "
	inputLine := m.styles.CommandLineStyle.Render(inputPrompt + m.passphraseInput.View())

	// Help text
	helpText := m.styles.HelpStyle.Render("Enter: submit | Esc: cancel | Ctrl+C: quit")

	// Build popup content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		keyInfo,
		"",
		inputLine,
		"",
		helpText,
	)

	popup := m.styles.PassphrasePromptStyle.Width(popupWidth).Height(popupHeight).Render(content)

	// Center the popup
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		popup,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")))
}
