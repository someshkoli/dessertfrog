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

	// Field labels and inputs
	fields := []struct {
		label       string
		placeholder string
		index       int
		isPassword  bool
	}{
		{"Driver", "postgres, mariadb, clickhouse", 0, false},
		{"Host", "localhost", 1, false},
		{"Port", "5432 (default based on driver)", 2, false},
		{"Username", "postgres (default based on driver)", 3, false},
		{"Password", "", 4, true},
		{"Database", "postgres (default based on driver)", 5, false},
		{"Schema", "public (postgres only)", 6, false},
	}

	for _, field := range fields {
		// Get current value
		var value string
		switch field.index {
		case 0:
			value = m.connInputDriver
		case 1:
			value = m.connInputHost
		case 2:
			value = m.connInputPort
		case 3:
			value = m.connInputUsername
		case 4:
			value = m.connInputPassword
		case 5:
			value = m.connInputDatabase
		case 6:
			value = m.connInputSchema
		}

		// Prepare display value
		var displayValue string
		if field.isPassword && value != "" {
			// Show dots for password
			displayValue = strings.Repeat("•", len(value))
		} else {
			displayValue = value
		}

		// Add cursor if this field is focused
		if m.connInputField == field.index {
			runes := []rune(displayValue)
			cursorPos := m.connInputCursor
			if cursorPos < 0 {
				cursorPos = 0
			}
			if cursorPos > len(runes) {
				cursorPos = len(runes)
			}
			beforeCursor := string(runes[:cursorPos])
			afterCursor := string(runes[cursorPos:])
			displayValue = beforeCursor + "█" + afterCursor
		}

		// Add placeholder if empty and not focused
		if displayValue == "" && m.connInputField != field.index {
			displayValue = m.styles.GhostTextStyle.Render(field.placeholder)
		}

		// Build the line with label and input
		var line string
		if m.connInputField == field.index {
			line = fmt.Sprintf("%-12s %s", field.label+":", displayValue)
			content.WriteString(m.styles.SelectedRowStyle.Render(line))
		} else {
			line = fmt.Sprintf("%-12s %s", field.label+":", displayValue)
			content.WriteString(m.styles.TableRowStyle.Render(line))
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
