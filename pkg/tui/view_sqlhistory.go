package tui

import (
	"fmt"
	"strings"
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
	title := m.styles.SQLHistoryTitleStyle.Render("SQL History")
	if len(m.sqlHistorySuggestions) > maxVisible {
		title += " " + m.styles.SQLHistoryCountStyle.Render(fmt.Sprintf("(%d/%d)", maxVisible, len(m.sqlHistorySuggestions)))
	} else {
		title += " " + m.styles.SQLHistoryCountStyle.Render(fmt.Sprintf("(%d)", len(m.sqlHistorySuggestions)))
	}

	lines = append(lines, title)
	lines = append(lines, "") // Empty line for spacing

	// Suggestion lines

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
		ghostTimestamp := m.styles.GhostTextStyle.Render(timestamp)

		// Highlight selected item
		var line string
		if i == m.sqlHistorySelected {
			line = ghostTimestamp + " " + m.styles.SQLHistorySelectedStyle.Render(fmt.Sprintf(" > %s ", displayQuery))
		} else {
			line = ghostTimestamp + " " + m.styles.SQLHistoryNormalStyle.Render(fmt.Sprintf("   %s", displayQuery))
		}

		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	return m.styles.SQLHistoryBorderStyle.Render(content)
}
