package tui

// View renders the UI
func (m Model) View() string {
	// Calculate available dimensions
	availableWidth := m.width - 6   // Account for screen border padding (1+1) and border (1+1) and margin
	availableHeight := m.height - 6 // Account for screen border, title, status line

	// Build the inner content
	var content string

	// Check if in table view mode
	if m.tableViewMode {
		// Render table data view
		title := m.renderTableDataTitle()
		content += title + "\n"

		tableDataView := m.renderTableDataView()
		content += tableDataView + "\n"

		// SQL query input or help text for table view
		if m.sqlQueryMode {
			content += "\n" // Add spacing before SQL input
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
			sqlInput := commandLineStyle.Render(sqlPrompt + beforeCursor + "█" + afterCursor)
			content += sqlInput
			content += "\n"
			help := helpStyle.Render("Enter: execute | Esc: cancel")
			content += help
		} else {
			// Show help text with 's' to view/edit query
			var helpText string
			if m.isCustomQuery {
				helpText = "i: edit | v: view | V: record | y: copy | Y: row | w/b: cell | hjkl: move | s: query | q: quit | o: back"
			} else {
				helpText = "i: edit | v: view | V: record | y: copy | Y: row | w/b: cell | hjkl: move | n/p: page | s: query | q: quit | o: back"
			}
			help := helpStyle.Render(helpText)
			content += help
		}
	} else {
		// Normal table list view
		title := titleStyle.Render("dessertfrog - Database Browser")
		content += title + "\n\n"

		// If connection failed, show error prominently instead of tables
		if m.connectionStatus == ConnectionFailed {
			errorBox := m.renderConnectionError()
			content += errorBox + "\n\n"
		} else if !m.sqlQueryMode {
			// Only show tables when not in SQL query mode
			// Inline search bar
			if m.inlineSearchMode {
				searchBar := m.renderInlineSearchBar()
				content += searchBar + "\n\n"
			}

			// Tables box - adjust height if inline search is active
			adjustedHeight := availableHeight
			if m.inlineSearchMode {
				adjustedHeight = availableHeight - 4 // Account for search bar height
			}
			tablesBox := m.renderTablesBox(availableWidth, adjustedHeight)
			content += tablesBox + "\n\n"
		} else {
			// When SQL query mode is active, add some spacing
			content += "\n"
		}

		// Command line, SQL query, or help text
		if m.commandMode {
			cmdLine := commandLineStyle.Render(m.commandBuffer)
			content += cmdLine
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
			sqlInput := commandLineStyle.Render(sqlPrompt + beforeCursor + "█" + afterCursor)
			content += sqlInput
			content += "\n"
			help := helpStyle.Render("Enter: execute | Esc: cancel")
			content += help
		} else if m.inlineSearchMode {
			help := helpStyle.Render("↑/↓: navigate | Tab: autocomplete | Esc: clear filter | Enter: open table")
			content += help
		} else {
			help := helpStyle.Render("/: filter | Ctrl+P: search | s: SQL query | j/k: scroll | Enter: view | g/G: top/bot | q: quit")
			content += help
		}
	}

	content += "\n\n"

	// Status line at the bottom
	statusLine := m.renderStatusLine()
	content += statusLine

	// Wrap everything in the screen border
	screenBorder := screenBorderStyle
	if m.height > 0 {
		screenBorder = screenBorder.Height(m.height - 4)
	}

	mainView := screenBorder.Render(content)

	// Render search popup overlay if active
	if m.searchMode {
		return m.renderSearchPopup(mainView)
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
		return m.renderCellValuePopup(mainView)
	}

	return mainView
}

// renderInlineSearchBar renders the inline search bar on the main view
func (m Model) renderInlineSearchBar() string {
	searchPrompt := "Filter: "
	cursorChar := "█"

	// Build input with ghost text suggestion
	var inputText string
	if m.inlineSearchSuggestion != "" {
		inputText = searchPrompt + m.inlineSearchQuery + ghostTextStyle.Render(m.inlineSearchSuggestion) + cursorChar
	} else {
		inputText = searchPrompt + m.inlineSearchQuery + cursorChar
	}

	return searchInputStyle.Render(inputText)
}
