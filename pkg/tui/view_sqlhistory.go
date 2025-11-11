package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderSQLHistorySuggestionsWindow renders the SQL history suggestions as an inline window above the input bar
func (m Model) renderSQLHistorySuggestionsWindow() string {
	if len(m.sqlHistorySuggestions) == 0 {
		return ""
	}

	// Show max 10 suggestions
	maxVisible := 10
	if len(m.sqlHistorySuggestions) < maxVisible {
		maxVisible = len(m.sqlHistorySuggestions)
	}

	var lines []string

	// Title line
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	title := titleStyle.Render("SQL History")
	if len(m.sqlHistorySuggestions) > maxVisible {
		title += " " + countStyle.Render(fmt.Sprintf("(%d/%d)", maxVisible, len(m.sqlHistorySuggestions)))
	} else {
		title += " " + countStyle.Render(fmt.Sprintf("(%d)", len(m.sqlHistorySuggestions)))
	}

	lines = append(lines, title)
	lines = append(lines, "") // Empty line for spacing

	// Suggestion lines
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("63")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	ghostTextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	for i := 0; i < maxVisible; i++ {
		entry := m.sqlHistorySuggestions[i]
		query := entry.Query

		// Format timestamp
		timestamp := entry.Timestamp.Format("01/02 15:04")

		// Calculate available space for query (reserve space for timestamp)
		timestampWidth := len(timestamp) + 2 // +2 for spacing
		maxQueryLen := m.width - timestampWidth - 15
		if maxQueryLen < 40 {
			maxQueryLen = 40
		}

		displayQuery := query
		if len(query) > maxQueryLen {
			// Replace newlines and multiple spaces with single space
			displayQuery = strings.ReplaceAll(query, "\n", " ")
			displayQuery = strings.ReplaceAll(displayQuery, "\t", " ")
			// Collapse multiple spaces
			for strings.Contains(displayQuery, "  ") {
				displayQuery = strings.ReplaceAll(displayQuery, "  ", " ")
			}

			if len(displayQuery) > maxQueryLen {
				// Truncate and add ellipsis
				displayQuery = displayQuery[:maxQueryLen-3] + "..."
			}
		} else {
			// Even if not truncating, clean up the display
			displayQuery = strings.ReplaceAll(displayQuery, "\n", " ")
			displayQuery = strings.ReplaceAll(displayQuery, "\t", " ")
		}

		// Build line with timestamp on the left in ghost text
		ghostTimestamp := ghostTextStyle.Render(timestamp)

		// Highlight selected item
		var line string
		if i == m.sqlHistorySelected {
			line = ghostTimestamp + " " + selectedStyle.Render(fmt.Sprintf(" > %s ", displayQuery))
		} else {
			line = ghostTimestamp + " " + normalStyle.Render(fmt.Sprintf("   %s", displayQuery))
		}

		lines = append(lines, line)
	}

	// Border style
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)

	content := strings.Join(lines, "\n")
	return borderStyle.Render(content)
}
