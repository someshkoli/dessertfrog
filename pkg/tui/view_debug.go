package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderDebugOverlay renders the debug overlay on top of the main view
func (m Model) renderDebugOverlay() string {
	if !m.debugMode {
		return ""
	}

	// Debug panel takes up full width
	debugWidth := m.width
	if debugWidth < 80 {
		debugWidth = 80
	}

	debugHeight := m.height / 3 // Take up bottom third of screen
	if debugHeight < 10 {
		debugHeight = 10
	}

	// Create styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1).
		Width(debugWidth - 4)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("220"))

	logStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("246"))

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("220")).
		Bold(true)

	focusIndicatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	// Build debug content in two columns

	// Calculate column widths (50/50 split)
	leftWidth := (debugWidth - 8) / 2
	rightWidth := debugWidth - 8 - leftWidth

	// Build left column (App State)
	var leftContent strings.Builder
	stateSectionHeader := "App State:"
	if m.debugPanelFocused && m.debugSelectedSection == 0 {
		stateSectionHeader = focusIndicatorStyle.Render("▶ ") + stateSectionHeader
	}
	leftContent.WriteString(sectionStyle.Render(stateSectionHeader))
	leftContent.WriteString("\n")

	stateLines := strings.Split(m.getDebugState(), "\n")
	maxStateLines := debugHeight - 6
	if maxStateLines < 5 {
		maxStateLines = 5
	}

	for i, line := range stateLines {
		if i >= maxStateLines {
			leftContent.WriteString(fmt.Sprintf("... (%d more)\n", len(stateLines)-maxStateLines))
			break
		}
		if line != "" {
			// Truncate to fit column width
			if len(line) > leftWidth-4 {
				line = line[:leftWidth-7] + "..."
			}
			leftContent.WriteString(fmt.Sprintf("  %s\n", line))
		}
	}

	// Build right column (Debug Logs)
	var rightContent strings.Builder
	logsSectionHeader := "Debug Logs:"
	if m.debugPanelFocused && m.debugSelectedSection == 1 {
		logsSectionHeader = focusIndicatorStyle.Render("▶ ") + logsSectionHeader
	}
	rightContent.WriteString(sectionStyle.Render(logsSectionHeader))
	rightContent.WriteString("\n")

	maxLogs := debugHeight - 6
	if maxLogs < 5 {
		maxLogs = 5
	}

	// Use scroll offset instead of always showing last N logs
	startIdx := m.debugLogScrollOffset
	endIdx := startIdx + maxLogs

	// Clamp to valid range
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(m.debugLogs) {
		endIdx = len(m.debugLogs)
	}

	if len(m.debugLogs) == 0 {
		rightContent.WriteString(logStyle.Render("  (no logs)"))
		rightContent.WriteString("\n")
	} else {
		// Show scroll indicator if there are logs above
		if startIdx > 0 {
			rightContent.WriteString(logStyle.Render(fmt.Sprintf("  ... (%d more above)", startIdx)))
			rightContent.WriteString("\n")
		}

		for i := startIdx; i < endIdx; i++ {
			logLine := m.debugLogs[i]
			// Truncate long lines to fit column
			if len(logLine) > rightWidth-4 {
				logLine = logLine[:rightWidth-7] + "..."
			}

			// Highlight selected log
			if m.debugPanelFocused && m.debugSelectedSection == 1 && i == m.debugSelectedLog {
				rightContent.WriteString(selectedStyle.Render(fmt.Sprintf("  %s", logLine)))
			} else {
				rightContent.WriteString(logStyle.Render(fmt.Sprintf("  %s", logLine)))
			}
			rightContent.WriteString("\n")
		}

		// Show scroll indicator if there are logs below
		if endIdx < len(m.debugLogs) {
			rightContent.WriteString(logStyle.Render(fmt.Sprintf("  ... (%d more below)", len(m.debugLogs)-endIdx)))
			rightContent.WriteString("\n")
		}
	}

	// Style the columns
	leftColumnStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(debugHeight - 6)

	rightColumnStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(debugHeight - 6)

	leftColumn := leftColumnStyle.Render(leftContent.String())
	rightColumn := rightColumnStyle.Render(rightContent.String())

	// Join columns side by side
	columnsView := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	// Build final content with title and help
	var finalContent strings.Builder

	// Title with focus indicator
	title := "DEBUG PANEL"
	if m.debugPanelFocused {
		title = focusIndicatorStyle.Render("▶ ") + title + focusIndicatorStyle.Render(" (FOCUSED)")
	}
	finalContent.WriteString(titleStyle.Render(title))
	finalContent.WriteString("\n\n")
	finalContent.WriteString(columnsView)
	finalContent.WriteString("\n\n")

	helpText := "F12: Toggle | F11: Clear | F10: Focus"
	if m.debugPanelFocused {
		helpText = "Tab: Switch Section | j/k: Navigate | Enter: Details | Esc: Unfocus"
	}
	finalContent.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(helpText))

	// Apply border
	debugPanel := borderStyle.Render(finalContent.String())

	return debugPanel
}

// renderDebugDetailPopup renders the detail popup for expanded debug content
func (m Model) renderDebugDetailPopup() string {
	if !m.debugDetailMode {
		return ""
	}

	// Popup takes up 80% of screen
	popupWidth := m.width * 80 / 100
	if popupWidth < 60 {
		popupWidth = 60
	}
	if popupWidth > 120 {
		popupWidth = 120
	}

	popupHeight := m.height * 80 / 100
	if popupHeight < 20 {
		popupHeight = 20
	}

	// Create styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(popupWidth - 6).
		Height(popupHeight - 4)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(popupWidth - 6).
		Align(lipgloss.Center)

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	// Build content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render(m.debugDetailTitle))
	content.WriteString("\n\n")

	// Content (with wrapping)
	lines := strings.Split(m.debugDetailContent, "\n")
	maxContentLines := popupHeight - 8
	if maxContentLines < 5 {
		maxContentLines = 5
	}

	displayedLines := 0
	for _, line := range lines {
		if displayedLines >= maxContentLines {
			content.WriteString(contentStyle.Render("... (truncated)"))
			content.WriteString("\n")
			break
		}

		// Word wrap long lines
		if len(line) > popupWidth-10 {
			// Simple word wrap
			for len(line) > popupWidth-10 {
				content.WriteString(contentStyle.Render(line[:popupWidth-10]))
				content.WriteString("\n")
				line = line[popupWidth-10:]
				displayedLines++
				if displayedLines >= maxContentLines {
					break
				}
			}
			if len(line) > 0 && displayedLines < maxContentLines {
				content.WriteString(contentStyle.Render(line))
				content.WriteString("\n")
				displayedLines++
			}
		} else {
			content.WriteString(contentStyle.Render(line))
			content.WriteString("\n")
			displayedLines++
		}
	}

	content.WriteString("\n")
	content.WriteString(helpStyle.Render("Esc/Enter: Close"))

	// Apply border
	popup := borderStyle.Render(content.String())

	// Center on screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		popup,
	)
}
