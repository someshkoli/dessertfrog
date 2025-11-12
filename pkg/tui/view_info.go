package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// renderStatusLine renders the bottom status line with DB info and connection status
func (m Model) renderStatusLine() string {
	var statusText string
	var statusStyle lipgloss.Style

	switch m.connectionStatus {
	case Connected:
		statusText = "● Connected"
		statusStyle = m.styles.ConnectedStyle
	case Connecting:
		statusText = "● Connecting..."
		statusStyle = m.styles.ConnectingStyle
	case ConnectionFailed:
		statusText = "● Disconnected"
		statusStyle = m.styles.DisconnectedStyle
	case Disconnected:
		statusText = "● Disconnected"
		statusStyle = m.styles.DisconnectedStyle
	}

	// Build status line: DB info on left, timing in middle (if available), status on right
	dbInfo := fmt.Sprintf("%s@%s:%d/%s",
		m.dbConfig.Username,
		m.dbConfig.Host,
		m.dbConfig.Port,
		m.dbConfig.Database,
	)
	if m.dbConfig.Schema != "" {
		dbInfo += fmt.Sprintf(" (schema: %s)", m.dbConfig.Schema)
	}

	// Add timing info if available (when viewing table data)
	var timingInfo string
	if m.tableViewMode && (m.queryTime != "" || m.fetchTime != "") {
		timingInfo = fmt.Sprintf("  query: %s | fetch: %s", m.queryTime, m.fetchTime)
	}

	// Calculate available width for content (account for screen border and padding)
	// m.width - 2 (border left/right) - 2 (screen padding left/right) = m.width - 4
	contentWidth := m.width - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	// Style the left, middle, and right parts
	leftPart := m.styles.StatusLineLeftStyle.Inline(true).Render(dbInfo)
	middlePart := m.styles.StatusLineLeftStyle.Inline(true).Render(timingInfo)
	rightPart := statusStyle.Inline(true).Render(statusText)

	// Calculate actual widths after styling
	leftWidth := lipgloss.Width(leftPart)
	middleWidth := lipgloss.Width(middlePart)
	rightWidth := lipgloss.Width(rightPart)

	// Create spacing to fill the gap
	spacingWidth := contentWidth - leftWidth - middleWidth - rightWidth
	if spacingWidth < 1 {
		spacingWidth = 1
	}
	spacing := lipgloss.NewStyle().Width(spacingWidth).Inline(true).Render("")

	// Combine everything
	content := leftPart + middlePart + spacing + rightPart

	return m.styles.StatusLineStyle.Width(contentWidth).Render(content)
}

// renderConnectionError renders the connection error prominently on main screen
func (m Model) renderConnectionError() string {
	errorContent := fmt.Sprintf("Connection Failed\n\n%s\n\nDetails:\nDriver: %s\nHost: %s:%d\nDatabase: %s\nUser: %s",
		m.connectionError,
		m.dbConfig.Driver,
		m.dbConfig.Host,
		m.dbConfig.Port,
		m.dbConfig.Database,
		m.dbConfig.Username,
	)
	return m.styles.ErrorBoxStyle.Render(errorContent)
}
