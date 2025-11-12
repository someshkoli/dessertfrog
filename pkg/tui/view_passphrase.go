package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/someshkoli/dessertfrog/pkg/helpers"
)

// renderPassphrasePromptPopup renders the passphrase prompt popup overlay
func (m Model) renderPassphrasePromptPopup(baseView string) string {
	// Popup dimensions
	popupWidth := 70
	popupHeight := 10

	// Title
	title := m.styles.PassphraseTitleStyle.Render("Enter SSH Key Passphrase")

	// Key info
	keyInfo := m.styles.PassphraseKeyInfoStyle.Render("Key: " + m.passphraseKeyName)

	// Masked input
	maskedInput := helpers.MaskString(m.passphraseInput, m.passphraseCursor)
	inputPrompt := "Passphrase: "

	// Insert cursor
	runes := []rune(maskedInput)
	cursorPos := m.passphraseCursor
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}
	beforeCursor := string(runes[:cursorPos])
	afterCursor := string(runes[cursorPos:])

	inputLine := m.styles.CommandLineStyle.Render(inputPrompt + beforeCursor + "█" + afterCursor)

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
