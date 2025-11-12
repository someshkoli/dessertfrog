package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/someshkoli/dessertfrog/pkg/helpers"
)

// renderKeySelectorPopup renders the encryption key selector popup
func (m Model) renderKeySelectorPopup(baseView string) string {
	popupWidth := helpers.Min(100, m.width-4)
	popupHeight := helpers.Min(30, m.height-4)

	// Build popup content
	var content strings.Builder

	// Title
	title := m.styles.ConnManagerTitleStyle.Render("Select Encryption Key")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Filter input
	filterPrompt := "Filter: "
	ghostText := ""
	if m.keySelectorFilter == "" {
		ghostText = m.styles.GhostTextStyle.Render("Type to filter keys...")
	}
	filterInput := m.styles.ConnManagerFilterStyle.Render(filterPrompt + m.keySelectorFilter + ghostText)
	content.WriteString(filterInput)
	content.WriteString("\n\n")

	// Key list
	keys := m.filteredKeys
	if len(keys) == 0 {
		if m.keySelectorFilter != "" {
			content.WriteString(m.styles.ErrorStyle.Render("No keys match filter"))
		} else {
			content.WriteString(m.styles.ErrorStyle.Render("No SSH/GPG keys found"))
			content.WriteString("\n\n")
			content.WriteString("Press 'g' to generate a new SSH key")
		}
	} else {
		// Calculate visible range for scrolling
		maxVisible := popupHeight - 12 // Account for title, filter, help
		startIdx := m.keySelectorScroll
		endIdx := helpers.Min(startIdx+maxVisible, len(keys))

		// Show scroll indicators
		if startIdx > 0 {
			content.WriteString(m.styles.ScrollIndicatorStyle.Render(fmt.Sprintf("↑ %d more above", startIdx)))
			content.WriteString("\n")
		}

		// Render visible keys
		for i := startIdx; i < endIdx; i++ {
			key := keys[i]
			line := fmt.Sprintf("  [%s] %s", key.Type, key.Name)

			// Highlight selected
			if i == m.keySelectorSelected {
				line = m.styles.SelectedRowStyle.Render(line)
			} else {
				line = m.styles.TableRowStyle.Render(line)
			}

			content.WriteString(line)
			content.WriteString("\n")
		}

		// Show scroll indicator for below
		if endIdx < len(keys) {
			remaining := len(keys) - endIdx
			content.WriteString(m.styles.ScrollIndicatorStyle.Render(fmt.Sprintf("↓ %d more below", remaining)))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")

	// Help text with mode indicator on the right
	var helpText string
	var modeIndicator string
	if m.keySelectorInsertMode {
		if len(keys) == 0 {
			helpText = "g: generate SSH key  esc: cancel"
		} else {
			helpText = "type to filter  enter: select  esc: normal mode"
		}
		modeIndicator = " INSERT "
	} else {
		if len(keys) == 0 {
			helpText = "g: generate SSH key  i: insert mode  esc: cancel"
		} else {
			helpText = "hjkl: navigate  g/G: top/bottom  i: insert  enter: select  esc: cancel"
		}
		modeIndicator = " NORMAL "
	}

	// Calculate spacing to push mode indicator to the right
	helpTextPlain := m.styles.HelpStyle.Render(helpText)
	availableWidth := popupWidth - 4 // Account for padding

	// Render help text and mode indicator side by side
	var modeStyle lipgloss.Style
	if m.keySelectorInsertMode {
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
