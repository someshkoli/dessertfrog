package tui

import "fmt"

// renderTablesContent renders the tables list content
func (m Model) renderTablesContent(visibleRows int) string {
	if m.tablesLoading {
		return "Loading tables..."
	}

	if m.tablesError != "" {
		return fmt.Sprintf("Error loading tables: %s", m.tablesError)
	}

	if len(m.tables) == 0 {
		return "No tables found in database"
	}

	// Determine which tables to display
	displayTables := m.tables
	if m.inlineSearchMode && m.inlineSearchQuery != "" {
		// Filter tables based on inline search query
		displayTables = filterTables(m.tables, m.inlineSearchQuery)
		if len(displayTables) == 0 {
			return "No tables match filter"
		}
	}

	// Build table list header
	content := fmt.Sprintf("%-45s %-18s %-10s %-12s\n", "Table Name", "Schema", "Columns", "Rows")
	content += "──────────────────────────────────────────────────────────────────────────────────────\n"

	// Render only visible rows
	endIndex := m.scrollOffset + visibleRows
	if endIndex > len(displayTables) {
		endIndex = len(displayTables)
	}

	for i := m.scrollOffset; i < endIndex; i++ {
		table := displayTables[i]
		columnCount := len(table.Columns)

		// Format row count
		var rowCountStr string
		if table.RowCount >= 0 {
			rowCountStr = fmt.Sprintf("%d", table.RowCount)
		} else {
			rowCountStr = "N/A"
		}

		rowText := fmt.Sprintf("%-45s %-18s %-10d %-12s",
			truncate(table.TableName, 45),
			truncate(table.SchemaName, 18),
			columnCount,
			rowCountStr,
		)

		// Highlight selected row
		if i == m.selectedRow {
			rowText = selectedRowStyle.Render(rowText)
		}

		content += rowText + "\n"
	}

	// Add scroll indicators
	if m.scrollOffset > 0 {
		content += "\n↑ More above (k to scroll up)"
	}
	if endIndex < len(displayTables) {
		content += "\n↓ More below (j to scroll down)"
	}

	return content
}

// renderTablesBox renders the complete tables box with borders
func (m Model) renderTablesBox(availableWidth, availableHeight int) string {
	// Calculate tables box dimensions to maximize space
	// availableWidth already accounts for screen padding, just need border space
	tablesBoxWidth := availableWidth - 2  // Account for border left/right
	if tablesBoxWidth < 40 {
		tablesBoxWidth = 40
	}

	tablesBoxHeight := availableHeight - 8 // Leave room for title, help, status line
	if tablesBoxHeight < 5 {
		tablesBoxHeight = 5
	}

	// Calculate visible rows (subtract 2 for header and separator line)
	visibleRows := tablesBoxHeight - 2
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Determine which tables to use for scroll calculations
	displayTables := m.tables
	if m.inlineSearchMode && m.inlineSearchQuery != "" {
		displayTables = filterTables(m.tables, m.inlineSearchQuery)
	}

	// Adjust scroll offset to keep selected row visible
	if m.selectedRow < m.scrollOffset {
		m.scrollOffset = m.selectedRow
	} else if m.selectedRow >= m.scrollOffset+visibleRows {
		m.scrollOffset = m.selectedRow - visibleRows + 1
	}

	// Ensure scroll offset doesn't exceed filtered table length
	if m.scrollOffset > len(displayTables)-visibleRows && len(displayTables) > visibleRows {
		m.scrollOffset = len(displayTables) - visibleRows
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	tablesContent := m.renderTablesContent(visibleRows)

	// Create styled tables box with maximum size
	return borderStyle.
		Width(tablesBoxWidth).
		Height(tablesBoxHeight).
		Render(tablesContent)
}
