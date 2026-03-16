package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/someshkoli/dessertfrog/pkg/helpers"
)

// renderConnectionInputPopup renders the new connection input form popup
func (m Model) renderConnectionInputPopup(baseView string) string {
	popupWidth := helpers.Min(80, m.width-4)
	popupHeight := helpers.Min(25, m.height-4)

	// Build popup content
	var content strings.Builder

	// Title
	title := m.styles.ConnManagerTitleStyle.Render("Create New Connection")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Field labels
	labels := []string{
		"Driver:",
		"Host:",
		"Port:",
		"Username:",
		"Password:",
		"Database:",
		"Schema:",
		"SSL Mode:",
	}

	for i, label := range labels {
		// Render label
		labelText := fmt.Sprintf("%-12s", label)

		// Get the input view
		inputView := m.connInputs[i].View()

		// Build the line
		if i == m.connInputField {
			// Focused field
			line := m.styles.SelectedRowStyle.Render(labelText + " " + inputView)
			content.WriteString(line)
		} else {
			// Unfocused field
			line := m.styles.TableRowStyle.Render(labelText + " " + inputView)
			content.WriteString(line)
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Show status message
	if m.connectionStatus == Connecting {
		connectingMsg := m.styles.ConnManagerTitleStyle.Render("⏳ Connecting...")
		content.WriteString(connectingMsg)
		content.WriteString("\n\n")
	} else if m.connectionError != "" {
		errorMsg := m.styles.ErrorStyle.Render("✗ Error: " + m.connectionError)
		content.WriteString(errorMsg)
		content.WriteString("\n\n")
	}

	// Help text
	helpText := "tab/shift+tab: navigate fields  enter: connect  esc: cancel"
	help := m.styles.HelpStyle.Render(helpText)
	content.WriteString(help)

	// Create popup box
	popup := m.styles.ConnManagerPopupStyle.
		Width(popupWidth).
		MaxHeight(popupHeight).
		Render(content.String())

	// Overlay on base view
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		popup,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}
