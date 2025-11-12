package tui

import (
	"fmt"
	"strings"

	"github.com/someshkoli/dessertfrog/pkg/driver"
)

// renderSchemaPanel renders the schema information panel on the right side
func (m Model) renderSchemaPanel(width, availableHeight int) string {
	// Calculate panel height same as tables box
	panelWidth := width + 1
	if panelWidth < 30 {
		panelWidth = 30
	}

	panelHeight := availableHeight - 4 // Leave room for layout spacing
	if panelHeight < 5 {
		panelHeight = 5
	}

	// Determine which tables to use (filtered or all)
	displayTables := m.tables
	if m.inlineSearchMode && m.inlineSearchQuery != "" {
		displayTables = filterTables(m.tables, m.inlineSearchQuery)
	}

	if len(displayTables) == 0 || m.selectedRow >= len(displayTables) {
		emptyMsg := m.styles.SchemaEmptyStyle.Render("No table selected")

		return m.styles.BorderStyle.
			Width(panelWidth).
			Height(panelHeight).
			Render(emptyMsg)
	}

	selectedTable := displayTables[m.selectedRow]

	// Build all content lines
	var allLines []string

	// Title
	titleText := fmt.Sprintf("Table: %s", selectedTable.TableName)
	if selectedTable.SchemaName != "" {
		titleText = fmt.Sprintf("Table: %s.%s", selectedTable.SchemaName, selectedTable.TableName)
	}

	allLines = append(allLines, m.styles.SchemaTitleStyle.Render(titleText))
	allLines = append(allLines, "")

	// Schema info sections
	if m.schemaInfoLoading {
		allLines = append(allLines, m.styles.SchemaLoadingStyle.Render("Loading detailed schema..."))
	} else if m.schemaInfo != nil {
		// Show detailed schema info fetched asynchronously
		infoLines := m.renderBasicSchemaInfoLines(*m.schemaInfo)
		allLines = append(allLines, infoLines...)
	} else {
		// Show basic info from TableSchema (just column count)
		infoLines := m.renderBasicSchemaInfoLines(selectedTable)
		allLines = append(allLines, infoLines...)
	}

	// Bound selected index (ensure we have at least one line)
	if len(allLines) == 0 {
		allLines = append(allLines, "No data")
	}

	if m.schemaPanelSelected >= len(allLines) {
		m.schemaPanelSelected = len(allLines) - 1
	}
	if m.schemaPanelSelected < 0 {
		m.schemaPanelSelected = 0
	}

	// Calculate visible area (subtract 2 for header lines)
	visibleLines := panelHeight - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Adjust scroll to keep cursor visible
	if m.schemaPanelSelected < m.schemaPanelScroll {
		m.schemaPanelScroll = m.schemaPanelSelected
	} else if m.schemaPanelSelected >= m.schemaPanelScroll+visibleLines {
		m.schemaPanelScroll = m.schemaPanelSelected - visibleLines + 1
	}

	// Ensure scroll doesn't exceed content
	if m.schemaPanelScroll > len(allLines)-visibleLines && len(allLines) > visibleLines {
		m.schemaPanelScroll = len(allLines) - visibleLines
	}
	if m.schemaPanelScroll < 0 {
		m.schemaPanelScroll = 0
	}

	// Check if we need to show "more above" indicator
	hasMoreAbove := m.schemaPanelScroll > 0

	// When showing "more above", skip the first line to make room for the indicator
	startIdx := m.schemaPanelScroll
	if hasMoreAbove {
		startIdx = m.schemaPanelScroll + 1
	}

	// Render visible lines
	endIdx := m.schemaPanelScroll + visibleLines
	if endIdx > len(allLines) {
		endIdx = len(allLines)
	}

	// Check if we need to show "more below" indicator
	hasMoreBelow := endIdx < len(allLines)

	var content strings.Builder

	// Build visible lines
	visibleCount := 0
	for i := startIdx; i < endIdx && visibleCount < visibleLines; i++ {
		line := allLines[i]

		// Highlight selected line when schema panel is focused
		if m.schemaPanelFocused && i == m.schemaPanelSelected {
			line = m.styles.SelectedRowStyle.Render(line)
		}

		content.WriteString(line)
		content.WriteString("\n")
		visibleCount++
	}
	content.WriteString("\n")

	// Add scroll indicators (same as table pane)
	if hasMoreAbove {
		content.WriteString("↑ More above (k to scroll up)\n")
	}
	if hasMoreBelow {
		content.WriteString("↓ More below (j to scroll down)")
	}

	// Choose border style based on focus
	panelBorderStyle := m.styles.InactiveBorderStyle
	if m.schemaPanelFocused {
		panelBorderStyle = m.styles.ActiveBorderStyle
	}

	// The borderStyle with Height will ensure fixed height
	// No extra trimming - lipgloss will handle the fixed height
	return panelBorderStyle.
		Width(panelWidth).
		Height(panelHeight).
		Render(content.String())
}

// renderBasicSchemaInfoLines renders basic schema information as individual lines
func (m Model) renderBasicSchemaInfoLines(table driver.TableSchema) []string {
	var lines []string

	// Type and Row count on same line
	infoLine := fmt.Sprintf("%s %s",
		m.styles.SchemaSectionStyle.Render("Type:"),
		m.styles.SchemaFieldStyle.Render(table.TableType))

	if table.RowCount > 0 {
		infoLine += fmt.Sprintf("  %s %s",
			m.styles.SchemaSectionStyle.Render("Rows:"),
			m.styles.SchemaFieldStyle.Render(fmt.Sprintf("%d", table.RowCount)))
	}
	lines = append(lines, infoLine)
	lines = append(lines, "")

	// Columns section
	if len(table.Columns) > 0 {
		lines = append(lines, m.styles.SchemaSectionStyle.Render("Columns:"))

		for i, col := range table.Columns {
			if i >= 50 {
				// Limit to avoid too many lines
				remaining := len(table.Columns) - 50
				lines = append(lines, m.styles.SchemaFieldStyle.Render(fmt.Sprintf("  ... and %d more", remaining)))
				break
			}

			// Column name and type
			columnLine := fmt.Sprintf("  %s %s",
				m.styles.SchemaColumnNameStyle.Render(col.Name),
				m.styles.SchemaTypeStyle.Render(col.DataType))

			// Add nullable indicator
			if !col.IsNullable {
				columnLine += m.styles.SchemaFieldStyle.Render(" NOT NULL")
			}

			// Add primary key indicator
			if col.IsPrimaryKey {
				columnLine += m.styles.SchemaPrimaryKeyStyle.Render(" PK")
			}

			// Add foreign key indicator
			if col.IsForeignKey {
				fkInfo := fmt.Sprintf(" FK -> %s(%s)", col.ForeignTable, col.ForeignColumn)
				columnLine += m.styles.SchemaForeignKeyStyle.Render(fkInfo)
			}

			lines = append(lines, columnLine)
		}
		lines = append(lines, "")
	}

	// Indexes section
	if len(table.Indexes) > 0 {
		lines = append(lines, m.styles.SchemaSectionStyle.Render("Indexes:"))

		for _, idx := range table.Indexes {
			indexLine := fmt.Sprintf("  %s", m.styles.SchemaColumnNameStyle.Render(idx.Name))

			// Add index type indicators
			if idx.IsPrimary {
				indexLine += m.styles.SchemaPrimaryKeyStyle.Render(" PRIMARY")
			} else if idx.IsUnique {
				indexLine += m.styles.SchemaTypeStyle.Render(" UNIQUE")
			}

			// Add columns
			if len(idx.Columns) > 0 {
				colList := strings.Join(idx.Columns, ", ")
				indexLine += m.styles.SchemaFieldStyle.Render(fmt.Sprintf(" (%s)", colList))
			}

			lines = append(lines, indexLine)
		}
		lines = append(lines, "")
	}

	// Comment (if available)
	if table.Comment != "" {
		lines = append(lines, m.styles.SchemaSectionStyle.Render("Comment:"))
		lines = append(lines, m.styles.SchemaFieldStyle.Render(table.Comment))
	}

	return lines
}

// updateSchemaPanelLineCount calculates and updates the line count for the schema panel
func (m Model) updateSchemaPanelLineCount() Model {
	if len(m.tables) == 0 || m.selectedRow >= len(m.tables) {
		m.schemaPanelLineCount = 0
		return m
	}

	selectedTable := m.tables[m.selectedRow]

	// Count lines: title + empty line + info lines
	lineCount := 2 // Title and empty line

	if m.schemaInfoLoading {
		lineCount += 1
	} else if m.schemaInfo != nil {
		infoLines := m.renderBasicSchemaInfoLines(*m.schemaInfo)
		lineCount += len(infoLines)
	} else {
		infoLines := m.renderBasicSchemaInfoLines(selectedTable)
		lineCount += len(infoLines)
	}

	m.schemaPanelLineCount = lineCount
	return m
}
