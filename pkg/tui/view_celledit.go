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
	content += m.styles.TitleStyle.Render(title) + "\n\n"

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
	inputBox := m.styles.CellEditInputBoxStyle.Copy().
		Width(contentWidth)

	inputText := beforeCursor + "█" + afterCursor
	content += inputBox.Render(inputText) + "\n"

	// Help text or command mode
	if m.cellEditCommandMode {
		cmdText := m.styles.CommandLineStyle.Render(m.cellEditCommand + "█")
		content += cmdText + "\n"
		if m.cellEditBufferCount > 0 {
			content += m.styles.HelpStyle.Render("Enter: execute | Esc: cancel | :w saves all pending edits")
		} else {
			content += m.styles.HelpStyle.Render("Enter: execute | Esc: cancel")
		}
	} else {
		content += m.styles.HelpStyle.Render("Enter: save to buffer  Ctrl+Enter: newline  :w commit all  Esc: cancel")
	}

	// Create popup box - don't set fixed width or height, let content determine it
	popup := m.styles.CellEditPopupStyle.Render(content)

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
