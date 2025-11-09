package tui

// adjustTableDataScroll adjusts the scroll offset to keep the selected row visible
func (m Model) adjustTableDataScroll() Model {
	// Calculate visible rows based on available height
	availableHeight := m.height - 14 // Same calculation as in renderTableDataView
	if availableHeight < 5 {
		availableHeight = 5
	}
	visibleRows := availableHeight - 2
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Adjust scroll offset to keep selected row visible
	if m.selectedDataRow < m.tableDataScrollY {
		m.tableDataScrollY = m.selectedDataRow
	} else if m.selectedDataRow >= m.tableDataScrollY+visibleRows {
		m.tableDataScrollY = m.selectedDataRow - visibleRows + 1
	}

	return m
}

// adjustTableDataHorizontalScroll adjusts horizontal scroll to keep selected cell visible
func (m Model) adjustTableDataHorizontalScroll() Model {
	if len(m.tableColumns) == 0 {
		return m
	}

	// Calculate column widths (same logic as in renderTableDataView)
	columnWidths := make([]int, len(m.tableColumns))
	for i, col := range m.tableColumns {
		columnWidths[i] = len(col)
		// Check data rows for max width
		for _, row := range m.tableData {
			if i < len(row) && len(row[i]) > columnWidths[i] {
				columnWidths[i] = len(row[i])
			}
		}
		// Cap column width at 30 characters for display
		if columnWidths[i] > 30 {
			columnWidths[i] = 30
		}
		// Minimum width of 10
		if columnWidths[i] < 10 {
			columnWidths[i] = 10
		}
	}

	// Calculate the position of the selected column
	selectedColStart := 0
	for i := 0; i < m.selectedDataCol; i++ {
		selectedColStart += columnWidths[i] + 3 // +3 for separator
	}
	selectedColWidth := columnWidths[m.selectedDataCol] + 3 // Include separator
	selectedColEnd := selectedColStart + selectedColWidth

	availableWidth := m.width - 12
	if availableWidth < 40 {
		availableWidth = 40
	}

	// Adjust scroll to show selected column
	if selectedColStart < m.tableDataScrollX {
		// Selected column is before visible area, scroll left
		// Position so the selected column starts at the left edge
		m.tableDataScrollX = selectedColStart
	} else if selectedColEnd > m.tableDataScrollX+availableWidth {
		// Selected column is after visible area, scroll right
		// Position so the selected column starts at the left edge of visible area
		m.tableDataScrollX = selectedColStart
	}

	return m
}

// openCellValuePopup opens the popup to display a cell value
func (m Model) openCellValuePopup(value string) Model {
	m.cellValuePopupMode = true
	m.cellValuePopupContent = value
	m.cellValuePopupScroll = 0
	m.cellValuePopupSelected = 0

	// Try to parse as JSON
	m.cellValuePopupIsJSON = false
	m.cellValuePopupTree = nil

	if value != "" && (value[0] == '{' || value[0] == '[') {
		// Attempt to parse as JSON
		tree, isJSON := buildJSONTree(value)
		if isJSON {
			m.cellValuePopupIsJSON = true
			m.cellValuePopupTree = tree
		}
	}

	return m
}

// expandJSONNode expands the selected JSON node
func (m Model) expandJSONNode() Model {
	if m.cellValuePopupSelected >= len(m.cellValuePopupTree) {
		return m
	}

	node := &m.cellValuePopupTree[m.cellValuePopupSelected]
	if !node.HasChildren {
		return m
	}

	if node.Expanded {
		// Already expanded, do nothing
		return m
	}

	// Mark as expanded
	node.Expanded = true

	// Rebuild the visible tree
	m.cellValuePopupTree = rebuildVisibleJSONTree(m.cellValuePopupTree)

	return m
}

// collapseJSONNode collapses the selected JSON node
func (m Model) collapseJSONNode() Model {
	if m.cellValuePopupSelected >= len(m.cellValuePopupTree) {
		return m
	}

	node := &m.cellValuePopupTree[m.cellValuePopupSelected]
	if !node.HasChildren {
		return m
	}

	if !node.Expanded {
		// Already collapsed, do nothing
		return m
	}

	// Mark as collapsed
	node.Expanded = false

	// Rebuild the visible tree
	m.cellValuePopupTree = rebuildVisibleJSONTree(m.cellValuePopupTree)

	return m
}

// openRecordViewPopup opens the popup to display an entire record as key-value pairs
func (m Model) openRecordViewPopup() Model {
	m.recordViewMode = true
	m.recordViewData = m.tableData[m.selectedDataRow]
	m.recordViewColumns = m.tableColumns
	m.recordViewSelected = 0
	m.recordViewScroll = 0

	return m
}
