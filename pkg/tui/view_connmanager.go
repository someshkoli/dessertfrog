package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/someshkoli/dessertfrog/pkg/helpers"
)

// renderConnectionManagerPopup renders the connection manager popup overlay
func (m Model) renderConnectionManagerPopup(baseView string) string {
	popupWidth := helpers.Min(100, m.width-4)
	popupHeight := helpers.Min(30, m.height-4)

	// Build popup content
	var content strings.Builder

	// Title
	title := m.styles.ConnManagerTitleStyle.Render("Connection Manager")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Filter input
	filterPrompt := "Filter: "
	filterInput := m.styles.ConnManagerFilterStyle.Render(filterPrompt + m.connManagerFilter.View())
	content.WriteString(filterInput)
	content.WriteString("\n\n")

	// Connection list
	connections := m.filteredConnections
	if len(connections) == 0 {
		if m.connManagerFilter.Value() != "" {
			content.WriteString(m.styles.ErrorStyle.Render("No connections match filter"))
		} else {
			content.WriteString(m.styles.ErrorStyle.Render("No saved connections"))
		}
		content.WriteString("\n\n")
		content.WriteString("Connect to a database to save it to history")
	} else {
		// Calculate visible range for scrolling
		maxVisible := popupHeight - 10 // Account for title, filter, help
		startIdx := m.connManagerScroll
		endIdx := helpers.Min(startIdx+maxVisible, len(connections))

		// Show scroll indicators
		if startIdx > 0 {
			content.WriteString(m.styles.ScrollIndicatorStyle.Render(fmt.Sprintf("↑ %d more above", startIdx)))
			content.WriteString("\n")
		}

		// Render visible connections
		for i := startIdx; i < endIdx; i++ {
			conn := connections[i]
			line := fmt.Sprintf("  %s", conn.Signature())

			// Highlight selected
			if i == m.connManagerSelected {
				line = m.styles.SelectedRowStyle.Render(line)
			} else {
				line = m.styles.TableRowStyle.Render(line)
			}

			content.WriteString(line)
			content.WriteString("\n")
		}

		// Show scroll indicator for below
		if endIdx < len(connections) {
			remaining := len(connections) - endIdx
			content.WriteString(m.styles.ScrollIndicatorStyle.Render(fmt.Sprintf("↓ %d more below", remaining)))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")

	// Help text with mode indicator on the right
	var helpText string
	var modeIndicator string
	if m.connManagerInsertMode {
		helpText = "type to filter  enter: connect  esc: normal mode"
		modeIndicator = " INSERT "
	} else {
		helpText = "hjkl: navigate  c: new connection  enter: connect  esc: close  q: quit"
		modeIndicator = " NORMAL "
	}

	// Calculate spacing to push mode indicator to the right
	helpTextPlain := m.styles.HelpStyle.Render(helpText)
	availableWidth := popupWidth - 4 // Account for padding

	// Render help text and mode indicator side by side
	var modeStyle lipgloss.Style
	if m.connManagerInsertMode {
		modeStyle = m.styles.ConnManagerInsertModeStyle
	} else {
		modeStyle = m.styles.ConnManagerNormalModeStyle
	}

	helpLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		helpTextPlain,
		strings.Repeat(" ", helpers.Max(1, availableWidth-lipgloss.Width(helpTextPlain)-len(modeIndicator))),
		modeStyle.Render(modeIndicator),
	)
	content.WriteString(helpLine)

	// Create popup box
	popup := m.styles.ConnManagerPopupStyle.
		Width(popupWidth).
		MaxHeight(popupHeight).
		Render(content.String())

	// Overlay on base view
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
