package tui

import (
	"fmt"
	"strings"

	"github.com/someshkoli/dessertfrog/pkg/helpers"
)

// renderTableDataView renders the table data view with scrollable content
func (m Model) renderTableDataView() string {
	if m.tableDataLoading {
		return "Loading table data..."
	}

	if m.tableDataError != "" {
		// Calculate available width for error message
		availableWidth := m.width - 14 // Screen padding + border
		if availableWidth < 40 {
			availableWidth = 40
		}

		// Wrap the error message to fit within available width
		wrappedError := wrapText(m.tableDataError, availableWidth-4)
		return fmt.Sprintf("Error loading table data:\n\n%s", wrappedError)
	}

	if len(m.tableData) == 0 {
		// Check if this is a non-SELECT query (UPDATE, DELETE, INSERT, etc.)
		if m.isCustomQuery && m.executedSQLQuery != "" {
			queryUpper := helpers.ToUpperFirst(m.executedSQLQuery)
			if helpers.StartsWithAny(queryUpper, []string{"UPDATE", "DELETE", "INSERT", "CREATE", "DROP", "ALTER", "TRUNCATE"}) {
				return "Query executed successfully.\n\nNo rows returned (this is expected for UPDATE/DELETE/INSERT queries)."
			}
		}
		return "No data in table"
	}

	// Calculate available space for content (accounting for border)
	availableWidth := m.width - 12 // Screen padding + border
	if availableWidth < 40 {
		availableWidth = 40
	}
	availableHeight := m.height - 14 // Title, help, status, padding, extra spacing
	if availableHeight < 5 {
		availableHeight = 5
	}

	// Column widths - calculate based on content
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

	// Calculate which columns to show based on horizontal scroll and available width
	// Account for border (2 chars) and padding (2 chars) = 4 chars total
	contentWidth := availableWidth - 4
	if contentWidth < 20 {
		contentWidth = 20
	}

	visibleColumns := []int{}
	totalWidth := 0
	cumulativeX := 0
	hasMoreLeft := false
	hasMoreRight := false

	for i := range m.tableColumns {
		columnWidth := columnWidths[i] + 3 // +3 for " │ " separator

		// Check if this column is within the visible horizontal scroll window
		if cumulativeX+columnWidth > m.tableDataScrollX {
			// This column starts within or after the scroll position
			if totalWidth+columnWidth <= contentWidth {
				visibleColumns = append(visibleColumns, i)
				totalWidth += columnWidth
			} else {
				// Can't fit more columns
				hasMoreRight = true
				break
			}
		} else {
			// This column is before the scroll position
			hasMoreLeft = true
		}

		cumulativeX += columnWidth
	}

	// Check if there are more columns to the right
	if len(visibleColumns) > 0 && visibleColumns[len(visibleColumns)-1] < len(m.tableColumns)-1 {
		hasMoreRight = true
	}

	// If no columns are visible (scrolled too far), reset scroll
	if len(visibleColumns) == 0 {
		// Show first columns that fit
		totalWidth = 0
		for i := range m.tableColumns {
			columnWidth := columnWidths[i] + 3
			if totalWidth+columnWidth <= contentWidth {
				visibleColumns = append(visibleColumns, i)
				totalWidth += columnWidth
			} else {
				hasMoreRight = true
				break
			}
		}
	}

	// Build header row
	var content strings.Builder
	for _, colIdx := range visibleColumns {
		content.WriteString(fmt.Sprintf("%-*s │ ", columnWidths[colIdx], truncate(m.tableColumns[colIdx], columnWidths[colIdx])))
	}
	content.WriteString("\n")

	// Add separator
	for _, colIdx := range visibleColumns {
		content.WriteString(strings.Repeat("─", columnWidths[colIdx]))
		content.WriteString("─┼─")
	}
	content.WriteString("\n")

	// Calculate visible rows based on vertical scroll
	visibleRows := availableHeight - 2 // Subtract header and separator
	if visibleRows < 1 {
		visibleRows = 1
	}

	startRow := m.tableDataScrollY
	endRow := startRow + visibleRows
	if endRow > len(m.tableData) {
		endRow = len(m.tableData)
	}

	// Render data rows
	for rowIdx := startRow; rowIdx < endRow; rowIdx++ {
		row := m.tableData[rowIdx]
		for _, colIdx := range visibleColumns {
			var cellText string
			if colIdx < len(row) {
				cellText = fmt.Sprintf("%-*s │ ", columnWidths[colIdx], truncate(row[colIdx], columnWidths[colIdx]))
			} else {
				cellText = fmt.Sprintf("%-*s │ ", columnWidths[colIdx], "")
			}

			// Check if this cell has pending edits
			bufferKey := fmt.Sprintf("%d:%d", rowIdx, colIdx)
			hasPendingEdit := false
			if _, exists := m.cellEditBuffer[bufferKey]; exists {
				hasPendingEdit = true
			}

			// Highlight selected cell
			if rowIdx == m.selectedDataRow && colIdx == m.selectedDataCol {
				cellText = m.styles.SelectedRowStyle.Render(cellText)
			} else if hasPendingEdit {
				// Show pending edit indicator (yellow/warning color)
				cellText = m.styles.CellPendingEditStyle.Render(cellText)
			}

			content.WriteString(cellText)
		}
		content.WriteString("\n")
	}

	// Add scroll indicators and pagination info
	currentPage := (m.tableDataOffset / 500) + 1
	rowStart := m.tableDataOffset + startRow + 1
	rowEnd := m.tableDataOffset + endRow

	scrollInfo := fmt.Sprintf("\nRows %d-%d (Page %d)", rowStart, rowEnd, currentPage)

	// Add pagination hints
	if m.tableDataOffset > 0 {
		scrollInfo += " | p: prev page"
	}
	if len(m.tableData) == 500 {
		scrollInfo += " | n: next page"
	}

	// Add column scroll indicators
	if hasMoreLeft {
		scrollInfo += " | ← more columns"
	}
	if hasMoreRight {
		scrollInfo += " | more columns →"
	}
	content.WriteString(scrollInfo)

	return m.styles.BorderStyle.
		Width(contentWidth).
		Height(availableHeight).
		Render(content.String())
}

// renderTableDataTitle renders the title for table data view
func (m Model) renderTableDataTitle() string {
	var title string

	// Show SQL query if this is a custom query result
	if m.isCustomQuery && m.executedSQLQuery != "" {
		// Show first 100 characters of query
		queryDisplay := m.executedSQLQuery
		if len(queryDisplay) > 100 {
			queryDisplay = queryDisplay[:100] + "..."
		}
		title = fmt.Sprintf("Query: %s", queryDisplay)
	} else if m.currentViewTable == nil {
		title = "Table Data View"
	} else {
		title = fmt.Sprintf("Table: %s.%s", m.currentViewTable.SchemaName, m.currentViewTable.TableName)
	}

	// Add clipboard notification if present
	if m.clipboardMessage != "" {
		title += "  " + m.styles.TableClipboardStyle.Render(m.clipboardMessage)
	}

	// Add filter indicator if active
	if m.tableContentFilter != "" {
		title += "  " + m.styles.TableFilterStyle.Render(fmt.Sprintf("[Filter: %s]", m.tableContentFilter))
	}

	return m.styles.TitleStyle.Render(title)
}

// wrapText wraps text to fit within a given width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for lineIdx, line := range lines {
		if lineIdx > 0 {
			result.WriteString("\n")
		}

		// If line is already short enough, keep it as is
		if len(line) <= width {
			result.WriteString(line)
			continue
		}

		// Wrap long lines
		words := strings.Fields(line)
		currentLine := ""

		for _, word := range words {
			// If adding this word would exceed width, start a new line
			if len(currentLine)+len(word)+1 > width {
				if currentLine != "" {
					result.WriteString(currentLine)
					result.WriteString("\n")
					currentLine = word
				} else {
					// Word itself is longer than width, split it
					for len(word) > width {
						result.WriteString(word[:width])
						result.WriteString("\n")
						word = word[width:]
					}
					currentLine = word
				}
			} else {
				// Add word to current line
				if currentLine != "" {
					currentLine += " " + word
				} else {
					currentLine = word
				}
			}
		}

		// Write remaining text
		if currentLine != "" {
			result.WriteString(currentLine)
		}
	}

	return result.String()
}
