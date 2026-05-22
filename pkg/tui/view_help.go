package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// getHelpMaxScroll calculates the maximum scroll value for the help popup
func (m Model) getHelpMaxScroll() int {
	// Get full help content
	fullContent := m.getHelpContent()
	allLines := strings.Split(fullContent, "\n")
	totalLines := len(allLines)

	// Calculate popup height
	popupHeight := m.height - 10
	if popupHeight < 20 {
		popupHeight = 20
	}

	// Calculate visible area (subtract borders and padding)
	visibleLines := popupHeight - 4
	if visibleLines < 10 {
		visibleLines = 10
	}

	// Calculate max scroll
	maxScroll := totalLines - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}

	return maxScroll
}

// getHelpContent returns the full help text content based on current mode
func (m Model) getHelpContent() string {
	var content strings.Builder

	// Title
	title := "Keyboard Shortcuts"
	content.WriteString(m.styles.TitleStyle.Render(title) + "\n\n")

	if m.tableViewMode {
		// Table view mode keybindings
		content.WriteString(m.styles.SchemaSectionStyle.Render("Navigation:") + "\n")
		content.WriteString("  hjkl / ←↓↑→   Move cursor\n")
		content.WriteString("  w / e          Move to next/previous column\n")
		content.WriteString("  0 / $          Jump to first/last column\n")
		content.WriteString("  g / G          Jump to first/last row\n")
		content.WriteString("  n / p          Next/previous page (500 rows)\n\n")

		content.WriteString(m.styles.SchemaSectionStyle.Render("View & Edit:") + "\n")
		content.WriteString("  v              View cell value in popup\n")
		content.WriteString("  V              View entire record as key-value pairs\n")
		content.WriteString("  i              Edit cell value\n")
		content.WriteString("  :w             Save all pending edits/deletes\n")
		content.WriteString("  r              Refresh data from database\n\n")

		content.WriteString(m.styles.SchemaSectionStyle.Render("Selection & Deletion:") + "\n")
		content.WriteString("  dd             Mark/unmark row for deletion\n")
		content.WriteString("  (Visual mode for multi-row selection - coming soon)\n\n")

		content.WriteString(m.styles.SchemaSectionStyle.Render("Copy:") + "\n")
		content.WriteString("  y              Copy cell value to clipboard\n")
		content.WriteString("  Y              Copy entire row as CSV\n\n")

		content.WriteString(m.styles.SchemaSectionStyle.Render("Search & Filter:") + "\n")
		content.WriteString("  /              Filter table content\n")
		content.WriteString("  Ctrl+P         Search all tables/views\n\n")

		content.WriteString(m.styles.SchemaSectionStyle.Render("Other:") + "\n")
		content.WriteString("  s              Edit/run SQL query\n")
		content.WriteString("  d              Open connections manager\n")
		content.WriteString("  o              Go back (for custom queries)\n")
		content.WriteString("  q              Quit / Go back\n")
		content.WriteString("  ?              Toggle this help\n")
	} else {
		// Main view keybindings
		content.WriteString(m.styles.SchemaSectionStyle.Render("Navigation:") + "\n")
		content.WriteString("  j / k / ↓ / ↑  Move cursor up/down\n")
		content.WriteString("  g / G          Jump to first/last table\n")
		content.WriteString("  Enter          Open selected table\n\n")

		content.WriteString(m.styles.SchemaSectionStyle.Render("Search:") + "\n")
		content.WriteString("  /              Filter tables (inline)\n")
		content.WriteString("  Ctrl+P         Search all entities (popup)\n")
		content.WriteString("  Tab            Switch between tables list and schema panel\n\n")

		content.WriteString(m.styles.SchemaSectionStyle.Render("Other:") + "\n")
		content.WriteString("  s              Open SQL query mode\n")
		content.WriteString("  d              Open connections manager\n")
		content.WriteString("  q              Quit\n")
		content.WriteString("  ?              Toggle this help\n")
	}

	content.WriteString("\n")
	content.WriteString(m.styles.HelpStyle.Render("Press ? or Esc to close | ↑↓ or jk to scroll"))

	return content.String()
}

// renderHelpPopup renders the help popup overlay with scrolling support
func (m Model) renderHelpPopup(mainView string) string {
	// Calculate popup dimensions
	popupWidth := m.width - 20
	if popupWidth < 70 {
		popupWidth = 70
	}
	if popupWidth > 100 {
		popupWidth = 100
	}

	popupHeight := m.height - 10
	if popupHeight < 20 {
		popupHeight = 20
	}

	// Get full help content
	fullContent := m.getHelpContent()
	allLines := strings.Split(fullContent, "\n")
	totalLines := len(allLines)

	// Calculate visible area (subtract borders and padding)
	visibleLines := popupHeight - 4 // Account for borders and padding
	if visibleLines < 10 {
		visibleLines = 10
	}

	// Calculate max scroll for display purposes (bounds are enforced in key handler)
	maxScroll := totalLines - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}

	// Get current scroll position (clamped to valid range for safety)
	scrollPos := m.helpPopupScroll
	if scrollPos < 0 {
		scrollPos = 0
	}
	if scrollPos > maxScroll {
		scrollPos = maxScroll
	}

	// Extract visible portion
	startLine := scrollPos
	endLine := startLine + visibleLines
	if endLine > totalLines {
		endLine = totalLines
	}

	visibleContent := strings.Join(allLines[startLine:endLine], "\n")

	// Add scroll indicators
	var scrollInfo string
	if scrollPos > 0 {
		scrollInfo += "↑ More above | "
	}
	if endLine < totalLines {
		scrollInfo += "↓ More below"
	}
	if scrollInfo != "" {
		visibleContent += "\n\n" + m.styles.HelpStyle.Render(scrollInfo)
	}

	// Wrap in popup style
	popup := m.styles.PopupStyle.
		Width(popupWidth).
		Height(popupHeight).
		Render(visibleContent)

	// Center the popup
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
