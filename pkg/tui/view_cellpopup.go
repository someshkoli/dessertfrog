package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderCellValuePopup renders the cell value popup overlay
func (m Model) renderCellValuePopup(mainView string) string {
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

	if m.cellValuePopupIsJSON {
		// Render JSON tree
		content.WriteString(m.renderJSONTree(popupHeight - 6))
	} else {
		// Render plain text
		content.WriteString(m.renderPlainText(popupHeight - 6))
	}

	// Create popup box with title
	title := " Cell Value "
	if m.cellValuePopupIsJSON {
		title = " Cell Value (JSON) "
	}

	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(popupWidth - 4).
		Height(popupHeight - 4)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	helpText := "j/k: navigate | h/l: collapse/expand | q/Esc/v: close"
	if !m.cellValuePopupIsJSON {
		helpText = "j/k: scroll | q/Esc/v: close"
	}

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

// renderJSONTree renders the JSON tree view
func (m Model) renderJSONTree(maxHeight int) string {
	if len(m.cellValuePopupTree) == 0 {
		return "Empty JSON"
	}

	var lines []string

	for i, node := range m.cellValuePopupTree {
		line := m.formatJSONNode(node, i == m.cellValuePopupSelected)
		lines = append(lines, line)
	}

	// Apply scrolling if needed
	visibleLines := maxHeight
	if visibleLines > len(lines) {
		visibleLines = len(lines)
	}

	// Auto-scroll to keep selected node visible
	startIdx := 0
	if m.cellValuePopupSelected >= visibleLines {
		startIdx = m.cellValuePopupSelected - visibleLines + 1
	}
	endIdx := startIdx + visibleLines
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	return strings.Join(lines[startIdx:endIdx], "\n")
}

// formatJSONNode formats a single JSON node for display
func (m Model) formatJSONNode(node JSONNode, isSelected bool) string {
	indent := strings.Repeat("  ", node.Depth)

	var line string
	var expandIndicator string

	if node.HasChildren {
		if node.Expanded {
			expandIndicator = "▼ "
		} else {
			expandIndicator = "▶ "
		}
	} else {
		expandIndicator = "  "
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("228"))
	typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)

	key := node.Key
	if key == "" {
		key = "root"
	}

	switch node.Type {
	case "object":
		count := len(node.Value.(map[string]interface{}))
		line = fmt.Sprintf("%s%s%s %s",
			indent,
			expandIndicator,
			keyStyle.Render(key),
			typeStyle.Render(fmt.Sprintf("{%d}", count)))

	case "array":
		count := len(node.Value.([]interface{}))
		line = fmt.Sprintf("%s%s%s %s",
			indent,
			expandIndicator,
			keyStyle.Render(key),
			typeStyle.Render(fmt.Sprintf("[%d]", count)))

	case "string":
		line = fmt.Sprintf("%s%s%s: %s",
			indent,
			expandIndicator,
			keyStyle.Render(key),
			valueStyle.Render(fmt.Sprintf("\"%v\"", node.Value)))

	case "number":
		line = fmt.Sprintf("%s%s%s: %s",
			indent,
			expandIndicator,
			keyStyle.Render(key),
			valueStyle.Render(fmt.Sprintf("%v", node.Value)))

	case "boolean":
		line = fmt.Sprintf("%s%s%s: %s",
			indent,
			expandIndicator,
			keyStyle.Render(key),
			valueStyle.Render(fmt.Sprintf("%v", node.Value)))

	case "null":
		line = fmt.Sprintf("%s%s%s: %s",
			indent,
			expandIndicator,
			keyStyle.Render(key),
			typeStyle.Render("null"))
	}

	if isSelected {
		selectedStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Bold(true)
		line = selectedStyle.Render(line)
	}

	return line
}

// renderPlainText renders plain text value with scrolling
func (m Model) renderPlainText(maxHeight int) string {
	lines := strings.Split(m.cellValuePopupContent, "\n")

	// Apply scrolling if needed
	visibleLines := maxHeight
	if visibleLines > len(lines) {
		visibleLines = len(lines)
	}

	startIdx := m.cellValuePopupScroll
	if startIdx > len(lines)-visibleLines {
		startIdx = len(lines) - visibleLines
	}
	if startIdx < 0 {
		startIdx = 0
	}

	endIdx := startIdx + visibleLines
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	content := strings.Join(lines[startIdx:endIdx], "\n")

	// Add scroll indicators if needed
	scrollInfo := ""
	if len(lines) > visibleLines {
		scrollInfo = fmt.Sprintf("\n\n[Lines %d-%d of %d]", startIdx+1, endIdx, len(lines))
	}

	return valueStyle.Render(content) + scrollInfo
}
