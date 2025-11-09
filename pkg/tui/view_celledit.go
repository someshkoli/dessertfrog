package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// renderCellEditPopup renders the cell edit popup overlay
func (m Model) renderCellEditPopup(mainView string) string {
	// Calculate popup dimensions
	popupWidth := m.width - 20
	if popupWidth < 60 {
		popupWidth = 60
	}
	if popupWidth > 100 {
		popupWidth = 100
	}

	// Get column name
	var columnName string
	if m.cellEditColIdx < len(m.tableColumns) {
		columnName = m.tableColumns[m.cellEditColIdx]
	}

	// Build popup content
	var content string

	// Title with buffer count indicator
	title := fmt.Sprintf("Edit Cell: %s", columnName)
	if m.cellEditBufferCount > 0 {
		title += fmt.Sprintf("  [%d pending]", m.cellEditBufferCount)
	}
	content += titleStyle.Render(title) + "\n\n"

	// Editable input with cursor
	runes := []rune(m.cellEditValue)
	cursorPos := m.cellEditCursor
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}
	beforeCursor := string(runes[:cursorPos])
	afterCursor := string(runes[cursorPos:])

	// Calculate content width: popup width - outer border (2) - outer padding (4) - inner border (2) - inner padding (2)
	contentWidth := popupWidth - 10
	if contentWidth < 30 {
		contentWidth = 30
	}

	// Multi-line text area style - smaller height
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Width(contentWidth).
		Height(3)

	inputText := beforeCursor + "█" + afterCursor
	content += inputBox.Render(inputText) + "\n"

	// Help text or command mode
	if m.cellEditCommandMode {
		cmdText := commandLineStyle.Render(m.cellEditCommand + "█")
		content += cmdText + "\n"
		if m.cellEditBufferCount > 0 {
			content += helpStyle.Render("Enter: execute | Esc: cancel | :w saves all pending edits")
		} else {
			content += helpStyle.Render("Enter: execute | Esc: cancel")
		}
	} else {
		content += helpStyle.Render("Enter: save to buffer  Ctrl+Enter: newline  :w commit all  Esc: cancel")
	}

	// Create popup box - don't set fixed width or height, let content determine it
	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	popup := popupStyle.Render(content)

	// Overlay on main view
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
