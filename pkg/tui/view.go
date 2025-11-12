package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the UI
func (m Model) View() string {
	// Calculate available dimensions
	availableWidth := m.width - 6   // Account for screen border padding (1+1) and border (1+1) and margin
	availableHeight := m.height - 6 // Account for screen border, title, status line

	// Build the main content (without help and status line)
	var mainContent string
	var helpMessage string

	// Check if in table view mode
	if m.tableViewMode {
		// Render table data view
		title := m.renderTableDataTitle()
		mainContent += title + "\n"

		tableDataView := m.renderTableDataView()
		mainContent += tableDataView

		// Inline search bar for content filtering
		if m.inlineSearchMode {
			searchBar := m.renderInlineSearchBar()
			helpMessage = searchBar + "\n" + m.styles.HelpStyle.Render("Enter: apply filter | Esc: cancel")
		} else if m.commandMode {
			cmdLine := m.styles.CommandLineStyle.Render(m.commandBuffer + "█")
			helpMessage = cmdLine + "\n" + m.styles.HelpStyle.Render("Enter: execute | Esc: cancel")
		} else if m.sqlQueryMode {
			sqlPrompt := "SQL Query: "
			// Insert cursor at the correct position (using runes for UTF-8)
			runes := []rune(m.sqlQueryInput)
			cursorPos := m.sqlQueryCursor
			if cursorPos < 0 {
				cursorPos = 0
			}
			if cursorPos > len(runes) {
				cursorPos = len(runes)
			}
			beforeCursor := string(runes[:cursorPos])
			afterCursor := string(runes[cursorPos:])
			sqlInput := m.styles.CommandLineStyle.Render(sqlPrompt + beforeCursor + "█" + afterCursor)
			if m.sqlHistorySuggestionsVisible {
				helpMessage = sqlInput + "\n" + m.styles.HelpStyle.Render("Enter: select | ↑/↓: navigate | Esc: close suggestions | Ctrl+N: toggle")
			} else {
				helpMessage = sqlInput + "\n" + m.styles.HelpStyle.Render("Enter: execute | Esc: cancel | Ctrl+N: show history")
			}
		} else {
			// Show help text with 's' to view/edit query
			var helpText string
			if m.cellEditBufferCount > 0 {
				// Show :w hint when there are pending edits
				if m.isCustomQuery {
					helpText = "i: edit | :w: save all | v: view | y: copy | Y: row | /: filter | d: connections | hjkl: move | s: query | q: quit | o: back"
				} else {
					helpText = "i: edit | :w: save all | v: view | y: copy | Y: row | /: filter | d: connections | hjkl: move | n/p: page | s: query | q: quit"
				}
			} else {
				if m.isCustomQuery {
					helpText = "i: edit | v: view | V: record | y: copy | Y: row | /: filter | d: connections | hjkl: move | s: query | q: quit | o: back"
				} else {
					helpText = "i: edit | v: view | V: record | y: copy | Y: row | /: filter | d: connections | hjkl: move | n/p: page | s: query | q: quit"
				}
			}
			helpMessage = m.styles.HelpStyle.Render(helpText)
		}
	} else {
		// Normal table list view with split view (tables on left, schema on right)
		// Row 1: Title
		title := m.styles.TitleStyle.Render("dessertfrog - Database Browser")
		mainContent += title + "\n"

		// If disconnected and not in any other mode, show connection manager as main view
		if m.connectionStatus == Disconnected && !m.sqlQueryMode {
			// Render connection manager inline with filter support
			var filterLine string
			if m.connManagerInsertMode {
				filterPrompt := "Filter: "
				ghostText := ""
				if m.connManagerFilter == "" {
					ghostText = m.styles.GhostTextStyle.Render("Type to filter connections...")
				}
				filterLine = m.styles.ConnManagerFilterStyle.Render(filterPrompt + m.connManagerFilter + ghostText)
			}

			connections := m.filteredConnections
			if len(connections) == 0 && m.connHistory != nil {
				if m.connManagerFilter == "" {
					connections = m.connHistory.GetAll()
				} else {
					connections = m.connHistory.Filter(m.connManagerFilter)
				}
			}

			if m.connManagerInsertMode && m.connManagerFilter != "" {
				mainContent += filterLine + "\n\n"
			}

			if len(connections) == 0 {
				if m.connManagerFilter != "" {
					mainContent += m.styles.ErrorStyle.Render("No connections match filter") + "\n"
				} else {
					mainContent += m.styles.ErrorStyle.Render("No saved connections") + "\n"
				}
				mainContent += "\n"
				if !m.connManagerInsertMode {
					mainContent += "Press 'C' to create a new connection\n"
				}
			} else {
				mainContent += "Saved Connections:\n\n"
				for i, conn := range connections {
					line := fmt.Sprintf("  %s", conn.Signature())
					if i == m.connManagerSelected {
						line = m.styles.SelectedRowStyle.Render(line)
					} else {
						line = m.styles.TableRowStyle.Render(line)
					}
					mainContent += line + "\n"
				}
			}

			// Help text with mode indicator
			var helpText string
			var modeIndicator string
			if m.connManagerInsertMode {
				helpText = "type to filter  enter: connect  esc: normal mode"
				modeIndicator = " INSERT "
			} else {
				helpText = "hjkl: navigate  g/G: top/bottom  C: new connection  i: insert mode  enter: connect  q: quit"
				modeIndicator = " NORMAL "
			}

			var modeStyle lipgloss.Style
			if m.connManagerInsertMode {
				modeStyle = m.styles.ConnManagerInsertModeStyle
			} else {
				modeStyle = m.styles.ConnManagerNormalModeStyle
			}

			helpMessage = lipgloss.JoinHorizontal(
				lipgloss.Left,
				m.styles.HelpStyle.Render(helpText),
				"  ",
				modeStyle.Render(modeIndicator),
			)
		} else if m.connectionStatus == ConnectionFailed {
			// If connection failed, show error prominently instead of tables
			errorBox := m.renderConnectionError()
			mainContent += errorBox
		} else if !m.sqlQueryMode {
			// Row 2: Split view - tables list on left, schema panel on right
			splitWidth := availableWidth / 2

			// Calculate panel height - pass directly, let functions handle internally
			panelHeight := availableHeight
			if m.inlineSearchMode {
				panelHeight = availableHeight - 3 // Account for search bar
			}

			// Left side: tables list
			tablesBox := m.renderTablesBox(splitWidth-1, panelHeight)

			// Right side: schema panel
			schemaPanel := m.renderSchemaPanel(splitWidth-1, panelHeight)

			// Join left and right panels horizontally
			splitView := lipgloss.JoinHorizontal(lipgloss.Top, tablesBox, schemaPanel)
			mainContent += splitView

			// Row 3: Filter box (if active)
			if m.inlineSearchMode {
				mainContent += "\n"
				searchBar := m.renderInlineSearchBar()
				mainContent += searchBar
			}
		}

		// SQL history suggestions window (if visible)
		if m.sqlQueryMode && m.sqlHistorySuggestionsVisible && len(m.sqlHistorySuggestions) > 0 {
			suggestionsWindow := m.renderSQLHistorySuggestionsWindow()
			mainContent += "\n" + suggestionsWindow
		}

		// Command line, SQL query, or help text
		if m.commandMode {
			cmdLine := m.styles.CommandLineStyle.Render(m.commandBuffer)
			helpMessage = cmdLine
		} else if m.sqlQueryMode {
			sqlPrompt := "SQL Query: "
			// Insert cursor at the correct position (using runes for UTF-8)
			runes := []rune(m.sqlQueryInput)
			cursorPos := m.sqlQueryCursor
			if cursorPos < 0 {
				cursorPos = 0
			}
			if cursorPos > len(runes) {
				cursorPos = len(runes)
			}
			beforeCursor := string(runes[:cursorPos])
			afterCursor := string(runes[cursorPos:])
			sqlInput := m.styles.CommandLineStyle.Render(sqlPrompt + beforeCursor + "█" + afterCursor)
			if m.sqlHistorySuggestionsVisible {
				helpMessage = sqlInput + "\n" + m.styles.HelpStyle.Render("Enter: select | ↑/↓: navigate | Esc: cancel | Ctrl+N: toggle")
			} else {
				helpMessage = sqlInput + "\n" + m.styles.HelpStyle.Render("Enter: execute | Esc: cancel | Ctrl+N: show history")
			}
		} else if m.inlineSearchMode {
			helpMessage = m.styles.HelpStyle.Render("↑/↓: navigate | Tab: autocomplete | Esc: clear filter | Enter: open table")
		} else {
			helpMessage = m.styles.HelpStyle.Render("/: filter | Ctrl+P: search | s: SQL query | d: connections | Tab: switch panel | j/k: scroll | Enter: view | g/G: top/bot | q: quit")
		}
	}

	// Status line at the bottom
	statusLine := m.renderStatusLine()

	// Calculate remaining height to push help and status to bottom
	// We need to fill the space between mainContent and help/status
	// Account for inline search bar in table view mode
	extraLines := 0
	if m.tableViewMode && m.inlineSearchMode {
		extraLines = 2 // Search bar + help text
	}

	contentHeight := m.height - 7 - extraLines // Account for borders, help, status, spacing, and search bar
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Use height to ensure help and status are at bottom
	paddedMainContent := lipgloss.NewStyle().Height(contentHeight).Render(mainContent)

	// Build final content with help and status at bottom
	content := lipgloss.JoinVertical(lipgloss.Left,
		paddedMainContent,
		"",
		helpMessage,
		"",
		statusLine,
	)

	// Wrap everything in the screen border
	screenBorder := m.styles.ScreenBorderStyle
	if m.height > 0 {
		// Set border height to fill terminal (account for border lines)
		screenBorder = screenBorder.Height(m.height - 2)
	}

	mainView := screenBorder.Render(content)

	// Render search popup overlay if active
	if m.searchMode {
		return m.renderSearchPopup(mainView)
	}

	// Render passphrase prompt popup overlay if active (highest priority for encryption)
	if m.passphrasePromptMode {
		return m.renderPassphrasePromptPopup(mainView)
	}

	// Render key selector popup overlay if active (highest priority - must be set up first)
	if m.keySelectorMode {
		return m.renderKeySelectorPopup(mainView)
	}

	// Render connection manager popup overlay if active
	if m.connManagerMode {
		return m.renderConnectionManagerPopup(mainView)
	}

	// Render connection input popup overlay if active
	if m.connInputMode {
		return m.renderConnectionInputPopup(mainView)
	}

	// Render record view popup overlay if active
	if m.recordViewMode {
		mainView = m.renderRecordViewPopup(mainView)
	}

	// Render cell edit popup overlay if active (highest priority for editing)
	if m.cellEditMode {
		return m.renderCellEditPopup(mainView)
	}

	// Render cell value popup overlay if active (on top of record view if both active)
	if m.cellValuePopupMode {
		mainView = m.renderCellValuePopup(mainView)
	}

	// Render debug overlay if active (always on top)
	if m.debugMode {
		debugOverlay := m.renderDebugOverlay()
		mainView = mainView + "\n" + debugOverlay
	}

	// Render debug detail popup if active (highest priority)
	if m.debugDetailMode {
		return m.renderDebugDetailPopup()
	}

	// Remove any trailing newlines to eliminate extra padding
	return strings.TrimRight(mainView, "\n")
}

// renderInlineSearchBar renders the inline search bar on the main view
func (m Model) renderInlineSearchBar() string {
	searchPrompt := "Filter: "
	cursorChar := "█"

	// Build input with ghost text suggestion
	var inputText string
	if m.inlineSearchSuggestion != "" {
		inputText = searchPrompt + m.inlineSearchQuery + m.styles.GhostTextStyle.Render(m.inlineSearchSuggestion) + cursorChar
	} else {
		inputText = searchPrompt + m.inlineSearchQuery + cursorChar
	}

	// Choose style based on focus - active when not in schema panel
	inputStyle := m.styles.ActiveSearchInputStyle
	if m.schemaPanelFocused {
		inputStyle = m.styles.InactiveSearchInputStyle
	}

	return inputStyle.Render(inputText)
}
