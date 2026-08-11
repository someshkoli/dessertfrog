package tui

import "fmt"

// renderTablesContent renders the tables list content
func (m Model) renderTablesContent(visibleRows, contentWidth int) string {
	if m.tablesLoading {
		return "Loading tables..."
	}

	if m.tablesError != "" {
		return fmt.Sprintf("Error loading tables: %s", m.tablesError)
	}

	if len(m.tables) == 0 {
		return "No tables found in database"
	}

	// Determine which tables to display (filter persists after leaving search mode)
	displayTables := m.displayTables()
	if len(displayTables) == 0 {
		return "No tables match filter"
	}

	// Calculate dynamic column widths
	tableNameWidth := contentWidth - 28 // Reserve 28 chars for Columns and Rows columns
	if tableNameWidth < 20 {
		tableNameWidth = 20
	}

	// Build table list header with right-aligned Columns and Rows
	content := fmt.Sprintf("%-*s %10s %12s\n", tableNameWidth, "Table Name", "Columns", "Rows")

	// Generate separator line based on content width (account for padding)
	separatorWidth := contentWidth - 4 // Account for border padding
	if separatorWidth < 10 {
		separatorWidth = 10
	}
	separator := ""
	for i := 0; i < separatorWidth; i++ {
		separator += "─"
	}
	content += separator + "\n"

	// Check if we need to show "more above" indicator
	hasMoreAbove := m.scrollOffset > 0

	// When showing "more above", skip the first row to make room for the indicator
	startIndex := m.scrollOffset
	if hasMoreAbove {
		startIndex = m.scrollOffset + 1
	}

	// Render only visible rows
	endIndex := m.scrollOffset + visibleRows
	if endIndex > len(displayTables) {
		endIndex = len(displayTables)
	}

	// Check if we need to show "more below" indicator (after calculating endIndex)
	hasMoreBelow := endIndex < len(displayTables)

	for i := startIndex; i < endIndex; i++ {
		table := displayTables[i]
		columnCount := len(table.Columns)

		// Format row count
		var rowCountStr string
		if table.RowCount >= 0 {
			rowCountStr = fmt.Sprintf("%d", table.RowCount)
		} else {
			rowCountStr = "N/A"
		}

		// Format with dynamic table name width and right-aligned numeric columns
		rowText := fmt.Sprintf("%-*s %10d %12s",
			tableNameWidth,
			truncate(table.TableName, tableNameWidth),
			columnCount,
			rowCountStr,
		)

		// Highlight selected row
		if i == m.selectedRow {
			rowText = m.styles.SelectedRowStyle.Render(rowText)
		}

		content += rowText + "\n"
	}

	// Add scroll indicators
	if hasMoreAbove {
		content += "\n↑ More above (k to scroll up)"
	}
	if hasMoreBelow {
		content += "\n↓ More below (j to scroll down)"
	}

	return content
}

// renderTablesBox renders the complete tables box with borders
func (m Model) renderTablesBox(availableWidth, availableHeight int) string {
	// Use the width passed in (already calculated by caller)
	tablesBoxWidth := availableWidth
	tablesBoxHeight := availableHeight - 6 // Leave room for layout spacing

	if tablesBoxWidth < 40 {
		tablesBoxWidth = 40
	}

	if tablesBoxHeight < 5 {
		tablesBoxHeight = 5
	}
	//
	// Calculate visible rows (subtract 2 for header and separator line)
	visibleRows := tablesBoxHeight - 2
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Determine which tables to use for scroll calculations (filter persists after leaving search mode)
	displayTables := m.displayTables()

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

	tablesContent := m.renderTablesContent(visibleRows, tablesBoxWidth)

	// Choose border style based on focus
	boxBorderStyle := m.styles.InactiveBorderStyle
	if !m.schemaPanelFocused {
		boxBorderStyle = m.styles.ActiveBorderStyle
	}

	// Create styled tables box with maximum size
	return boxBorderStyle.
		Width(tablesBoxWidth).
		Height(tablesBoxHeight).
		Render(tablesContent)
}
