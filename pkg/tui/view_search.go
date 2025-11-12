package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/someshkoli/dessertfrog/pkg/driver"
)

// renderSearchPopup renders the search popup overlay
func (m Model) renderSearchPopup(baseView string) string {
	// Search input with autocomplete ghost text
	searchPrompt := "Search tables: "
	cursorChar := "█"

	// Build input with ghost text suggestion
	var inputText string
	if m.searchSuggestion != "" {
		// Show query + ghost text + cursor
		inputText = searchPrompt + m.searchQuery + m.styles.GhostTextStyle.Render(m.searchSuggestion) + cursorChar
	} else {
		// Just show query + cursor
		inputText = searchPrompt + m.searchQuery + cursorChar
	}

	searchInput := m.styles.SearchInputStyle.Render(inputText)

	// Build filtered results
	var resultsContent string
	maxResults := 10
	displayResults := m.filteredTables
	if len(displayResults) > maxResults {
		displayResults = displayResults[:maxResults]
	}

	if len(m.filteredTables) == 0 {
		resultsContent = "\nNo matches found"
	} else {
		resultsContent = fmt.Sprintf("\n\nResults (%d):\n", len(m.filteredTables))
		for i, entity := range displayResults {
			// Format with entity type prefix
			entityDisplay := formatEntityDisplay(entity)

			// Determine column count or info
			var info string
			if entity.EntityType == driver.EntityTable || entity.EntityType == driver.EntityView || entity.EntityType == driver.EntityMaterializedView {
				info = fmt.Sprintf("%d cols", len(entity.Columns))
			} else {
				info = truncate(entity.Comment, 20)
			}

			resultLine := fmt.Sprintf("  %-50s %-15s %s",
				truncate(entityDisplay, 50),
				truncate(entity.SchemaName, 15),
				info,
			)
			if i == m.searchSelected {
				resultLine = m.styles.SelectedRowStyle.Render(resultLine)
			}
			resultsContent += resultLine + "\n"
		}
		if len(m.filteredTables) > maxResults {
			resultsContent += fmt.Sprintf("\n  ... and %d more", len(m.filteredTables)-maxResults)
		}
	}

	helpText := "\n↑/↓: navigate | Tab: autocomplete | Prefix: table/, view/, function/, trigger/ | Enter: select | Esc: close"
	popupContent := searchInput + resultsContent + helpText

	popup := m.styles.PopupStyle.Render(popupContent)

	// Center the popup on screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		popup,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#333333")),
	)
}
