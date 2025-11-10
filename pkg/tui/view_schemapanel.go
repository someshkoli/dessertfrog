package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/someshkoli/dessertfrog/pkg/driver"
)

// renderSchemaPanel renders the schema information panel on the right side
func (m Model) renderSchemaPanel(width, availableHeight int) string {
	// Calculate panel height same as tables box
	panelWidth := width - 2
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
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			Render("No table selected")

		return borderStyle.
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

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Underline(true)

	allLines = append(allLines, titleStyle.Render(titleText))
	allLines = append(allLines, "")

	// Schema info sections
	if m.schemaInfoLoading {
		loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		allLines = append(allLines, loadingStyle.Render("Loading detailed schema..."))
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

	// Render visible lines
	endIdx := m.schemaPanelScroll + visibleLines
	if endIdx > len(allLines) {
		endIdx = len(allLines)
	}

	var content strings.Builder

	// Build visible lines
	visibleCount := 0
	for i := m.schemaPanelScroll; i < endIdx && visibleCount < visibleLines; i++ {
		line := allLines[i]

		// Highlight selected line when schema panel is focused
		if m.schemaPanelFocused && i == m.schemaPanelSelected {
			line = selectedRowStyle.Render(line)
		}

		content.WriteString(line)
		content.WriteString("\n")
		visibleCount++
	}
	content.WriteString("\n")

	// Add scroll indicators (same as table pane)
	if m.schemaPanelScroll > 0 {
		content.WriteString("↑ More above (k to scroll up)\n")
	}
	if endIdx < len(allLines) {
		content.WriteString("↓ More below (j to scroll down)")
	}

	// Choose border style based on focus
	panelBorderStyle := inactiveBorderStyle
	if m.schemaPanelFocused {
		panelBorderStyle = activeBorderStyle
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

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))

	fieldStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	columnNameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("228"))

	typeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141"))

	// Type and Row count on same line
	infoLine := fmt.Sprintf("%s %s",
		sectionStyle.Render("Type:"),
		fieldStyle.Render(table.TableType))

	if table.RowCount > 0 {
		infoLine += fmt.Sprintf("  %s %s",
			sectionStyle.Render("Rows:"),
			fieldStyle.Render(fmt.Sprintf("%d", table.RowCount)))
	}
	lines = append(lines, infoLine)
	lines = append(lines, "")

	// Columns section
	if len(table.Columns) > 0 {
		lines = append(lines, sectionStyle.Render("Columns:"))

		for i, col := range table.Columns {
			if i >= 50 {
				// Limit to avoid too many lines
				remaining := len(table.Columns) - 50
				lines = append(lines, fieldStyle.Render(fmt.Sprintf("  ... and %d more", remaining)))
				break
			}

			// Column name and type
			columnLine := fmt.Sprintf("  %s %s",
				columnNameStyle.Render(col.Name),
				typeStyle.Render(col.DataType))

			// Add nullable indicator
			if !col.IsNullable {
				columnLine += fieldStyle.Render(" NOT NULL")
			}

			// Add primary key indicator
			if col.IsPrimaryKey {
				columnLine += lipgloss.NewStyle().
					Foreground(lipgloss.Color("205")).
					Bold(true).
					Render(" PK")
			}

			// Add foreign key indicator
			if col.IsForeignKey {
				fkInfo := fmt.Sprintf(" FK -> %s(%s)", col.ForeignTable, col.ForeignColumn)
				columnLine += lipgloss.NewStyle().
					Foreground(lipgloss.Color("214")).
					Render(fkInfo)
			}

			lines = append(lines, columnLine)
		}
		lines = append(lines, "")
	}

	// Indexes section
	if len(table.Indexes) > 0 {
		lines = append(lines, sectionStyle.Render("Indexes:"))

		for _, idx := range table.Indexes {
			indexLine := fmt.Sprintf("  %s", columnNameStyle.Render(idx.Name))

			// Add index type indicators
			if idx.IsPrimary {
				indexLine += lipgloss.NewStyle().
					Foreground(lipgloss.Color("205")).
					Bold(true).
					Render(" PRIMARY")
			} else if idx.IsUnique {
				indexLine += lipgloss.NewStyle().
					Foreground(lipgloss.Color("141")).
					Render(" UNIQUE")
			}

			// Add columns
			if len(idx.Columns) > 0 {
				colList := strings.Join(idx.Columns, ", ")
				indexLine += fieldStyle.Render(fmt.Sprintf(" (%s)", colList))
			}

			lines = append(lines, indexLine)
		}
		lines = append(lines, "")
	}

	// Comment (if available)
	if table.Comment != "" {
		lines = append(lines, sectionStyle.Render("Comment:"))
		lines = append(lines, fieldStyle.Render(table.Comment))
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
