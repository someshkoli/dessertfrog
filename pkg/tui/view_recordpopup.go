package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderRecordViewPopup renders the record view popup showing all fields as key-value pairs
func (m Model) renderRecordViewPopup(mainView string) string {
	// Calculate popup dimensions (80% of screen, centered)
	popupWidth := int(float64(m.width) * 0.8)
	popupHeight := int(float64(m.height) * 0.8)
	if popupWidth < 60 {
		popupWidth = 60
	}
	if popupHeight < 20 {
		popupHeight = 20
	}

	// Build popup content
	var content strings.Builder

	// Render all fields as key-value pairs
	content.WriteString(m.renderRecordFields(popupHeight - 6))

	// Create popup box with title
	title := fmt.Sprintf(" Record View (Row %d) ", m.selectedDataRow+m.tableDataOffset+1)

	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(popupWidth - 4).
		Height(popupHeight - 4)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	helpText := "j/k: navigate  v: view value  q/Esc/V: close"

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	popupContent := fmt.Sprintf("%s\n\n%s\n\n%s",
		titleStyle.Render(title),
		content.String(),
		helpStyle.Render(helpText))

	popup := popupStyle.Render(popupContent)

	// Overlay popup on main view
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		popup,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}

// renderRecordFields renders all fields as key-value pairs
func (m Model) renderRecordFields(maxHeight int) string {
	if len(m.recordViewData) == 0 {
		return "No data"
	}

	var lines []string

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	jsonIndicatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("228")).
		Italic(true)

	// Build all key-value pairs
	for i := 0; i < len(m.recordViewColumns) && i < len(m.recordViewData); i++ {
		key := m.recordViewColumns[i]
		value := m.recordViewData[i]

		// Truncate long values for display
		displayValue := value
		isJSON := false
		if len(value) > 0 && (value[0] == '{' || value[0] == '[') {
			// Check if it's JSON
			if _, ok := buildJSONTree(value); ok {
				isJSON = true
			}
		}

		// Truncate if too long
		maxValueLen := 80
		if len(displayValue) > maxValueLen {
			displayValue = displayValue[:maxValueLen] + "..."
		}

		// Replace newlines with spaces for display
		displayValue = strings.ReplaceAll(displayValue, "\n", " ")
		displayValue = strings.ReplaceAll(displayValue, "\r", "")

		var line string
		if isJSON {
			line = fmt.Sprintf("%-25s: %s %s",
				keyStyle.Render(key),
				valueStyle.Render(displayValue),
				jsonIndicatorStyle.Render("[JSON]"))
		} else {
			line = fmt.Sprintf("%-25s: %s",
				keyStyle.Render(key),
				valueStyle.Render(displayValue))
		}

		// Highlight selected field
		if i == m.recordViewSelected {
			selectedStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("230")).
				Bold(true)
			line = selectedStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// Apply scrolling if needed
	visibleLines := maxHeight
	if visibleLines > len(lines) {
		visibleLines = len(lines)
	}

	// Auto-scroll to keep selected field visible
	startIdx := 0
	if m.recordViewSelected >= visibleLines {
		startIdx = m.recordViewSelected - visibleLines + 1
	}
	endIdx := startIdx + visibleLines
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	return strings.Join(lines[startIdx:endIdx], "\n")
}
