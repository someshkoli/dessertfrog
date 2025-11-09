package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
)

// copyCellToClipboard copies the current cell value to clipboard
func (m Model) copyCellToClipboard(value string) Model {
	err := clipboard.WriteAll(value)
	if err != nil {
		m.clipboardMessage = fmt.Sprintf("Copy failed: %v", err)
	} else {
		m.clipboardMessage = "Cell copied!"
	}

	return m
}

// copyRowToClipboard copies the entire current row to clipboard in CSV format
func (m Model) copyRowToClipboard() Model {
	if m.selectedDataRow >= len(m.tableData) {
		return m
	}

	row := m.tableData[m.selectedDataRow]
	csvRow := formatRowAsCSV(row)

	err := clipboard.WriteAll(csvRow)
	if err != nil {
		m.clipboardMessage = fmt.Sprintf("Copy failed: %v", err)
	} else {
		m.clipboardMessage = "Row copied as CSV!"
	}

	return m
}

// formatRowAsCSV formats a row as CSV with proper escaping
func formatRowAsCSV(row []string) string {
	var fields []string
	for _, field := range row {
		fields = append(fields, escapeCSVField(field))
	}
	return strings.Join(fields, ",")
}

// escapeCSVField escapes a field for CSV format
func escapeCSVField(field string) string {
	// Check if field needs quoting (contains comma, quote, newline, or carriage return)
	needsQuoting := strings.ContainsAny(field, ",\"\n\r")

	if needsQuoting {
		// Escape quotes by doubling them
		escaped := strings.ReplaceAll(field, "\"", "\"\"")
		// Wrap in quotes
		return fmt.Sprintf("\"%s\"", escaped)
	}

	return field
}
