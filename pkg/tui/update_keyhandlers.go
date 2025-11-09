package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/someshkoli/dessertfrog/pkg/driver"
)

// handleKeyPress handles all keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle debug detail popup (highest priority when active)
	if m.debugDetailMode {
		return m.handleDebugDetailKeys(msg)
	}

	// Handle global key bindings (work in any mode)
	key := msg.String()
	if cmd, ok := getCommand(key, m.keyBindings.Global); ok {
		switch cmd {
		case CommandToggleDebug:
			m = m.toggleDebugMode()
			return m, nil
		case CommandClearDebugLogs:
			if m.debugMode {
				m = m.clearDebugLogs()
			}
			return m, nil
		case CommandToggleDebugFocus:
			if m.debugMode {
				m = m.toggleDebugFocus()
			}
			return m, nil
		case CommandQuit:
			return m, tea.Quit
		}
	}

	// Handle debug panel navigation when focused
	if m.debugMode && m.debugPanelFocused {
		return m.handleDebugPanelKeys(msg)
	}

	// Handle search mode first - highest priority to allow typing
	if m.searchMode {
		return m.handleSearchModeKeys(msg)
	}

	// Handle inline search mode
	if m.inlineSearchMode {
		return m.handleInlineSearchModeKeys(msg)
	}

	// Handle command mode
	if m.commandMode {
		return m.handleCommandModeKeys(msg)
	}

	// Handle cell value popup mode (from cell or record view)
	if m.cellValuePopupMode {
		return m.handleCellValuePopupKeys(msg)
	}

	// Handle record view popup mode
	if m.recordViewMode {
		return m.handleRecordViewKeys(msg)
	}

	// Handle cell edit mode
	if m.cellEditMode {
		return m.handleCellEditModeKeys(msg)
	}

	// Handle SQL query mode
	if m.sqlQueryMode {
		return m.handleSQLQueryModeKeys(msg)
	}

	// Handle table view mode
	if m.tableViewMode {
		return m.handleTableViewModeKeys(msg)
	}

	// Handle normal mode
	return m.handleNormalModeKeys(msg)
}

// handleSearchModeKeys handles keyboard input in search mode
func (m Model) handleSearchModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Select the currently highlighted table
		if len(m.filteredTables) > 0 && m.searchSelected < len(m.filteredTables) {
			selectedEntity := m.filteredTables[m.searchSelected]

			// Check if the selected entity is a table/view that can be opened
			if selectedEntity.EntityType == driver.EntityTable || selectedEntity.EntityType == driver.EntityView || selectedEntity.EntityType == driver.EntityMaterializedView {
				// Open table data view
				m.tableViewMode = true
				m.currentViewTable = &selectedEntity
				m.tableDataLoading = true
				m.tableDataError = ""
				m.tableDataOffset = 0
				// Close search popup
				m.searchMode = false
				m.searchQuery = ""
				m.searchSelected = 0
				return m, fetchTableData(m.driver, selectedEntity.SchemaName, selectedEntity.TableName, 0)
			} else {
				// Non-table entity: find in tables list and update selectedRow, stay on home screen
				for i, table := range m.tables {
					if table.TableName == selectedEntity.TableName && table.SchemaName == selectedEntity.SchemaName {
						m.selectedRow = i
						break
					}
				}
				// Close search popup
				m.searchMode = false
				m.searchQuery = ""
				m.searchSelected = 0
			}
		} else {
			// Close search popup
			m.searchMode = false
			m.searchQuery = ""
			m.searchSelected = 0
		}

	case "esc":
		// Close search popup
		m.searchMode = false
		m.searchQuery = ""
		m.searchSuggestion = ""
		m.searchSelected = 0

	case "tab":
		// Accept autocomplete suggestion
		if m.searchSuggestion != "" {
			m.searchQuery += m.searchSuggestion
			m.searchSuggestion = getAutocompleteSuggestion(m.searchQuery)
			m.filteredTables = filterTables(m.allEntities, m.searchQuery)
			m.searchSelected = 0
		}

	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.searchSuggestion = getAutocompleteSuggestion(m.searchQuery)
			m.filteredTables = filterTables(m.allEntities, m.searchQuery)
			m.searchSelected = 0
		}

	case "down":
		// Navigate down in search results
		if m.searchSelected < len(m.filteredTables)-1 {
			m.searchSelected++
		}

	case "up":
		// Navigate up in search results
		if m.searchSelected > 0 {
			m.searchSelected--
		}

	case "ctrl+c":
		return m, tea.Quit

	default:
		// Append to search query (only single characters)
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
			m.searchSuggestion = getAutocompleteSuggestion(m.searchQuery)
			m.filteredTables = filterTables(m.allEntities, m.searchQuery)
			m.searchSelected = 0
		}
	}

	return m, nil
}

// handleCommandModeKeys handles keyboard input in command mode
func (m Model) handleCommandModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Execute command
		if m.commandBuffer == ":q" || m.commandBuffer == ":quit" {
			return m, tea.Quit
		} else if m.commandBuffer == ":w" || m.commandBuffer == ":w " {
			// Batch update all cells in buffer
			m.commandMode = false
			m.commandBuffer = ""
			return m.batchUpdateCells()
		}
		// Reset command mode
		m.commandMode = false
		m.commandBuffer = ""

	case "esc":
		// Cancel command mode
		m.commandMode = false
		m.commandBuffer = ""

	case "backspace":
		if len(m.commandBuffer) > 0 {
			m.commandBuffer = m.commandBuffer[:len(m.commandBuffer)-1]
		}

	default:
		// Append to command buffer
		m.commandBuffer += msg.String()
	}

	return m, nil
}

// handleInlineSearchModeKeys handles keyboard input in inline search mode
func (m Model) handleInlineSearchModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if we're in table view mode (content filtering) or normal mode (table filtering)
	inTableView := m.tableViewMode

	switch msg.String() {
	case "enter":
		if inTableView {
			// Save the filter and close search mode (filter already applied in real-time)
			m.tableContentFilter = m.inlineSearchQuery
			m.inlineSearchMode = false
			m.inlineSearchQuery = ""
			return m, nil
		}

		// Open table view for selected table (normal mode)
		displayTables := m.tables
		if m.inlineSearchQuery != "" {
			displayTables = filterTables(m.tables, m.inlineSearchQuery)
		}

		if len(displayTables) > 0 && m.selectedRow < len(displayTables) {
			selectedTable := displayTables[m.selectedRow]
			m.tableViewMode = true
			m.currentViewTable = &selectedTable
			m.tableDataLoading = true
			m.tableDataError = ""
			m.tableDataOffset = 0
			// Close inline search mode
			m.inlineSearchMode = false
			m.inlineSearchQuery = ""
			m.inlineSearchSuggestion = ""
			return m, fetchTableData(m.driver, selectedTable.SchemaName, selectedTable.TableName, 0)
		}

		// If no table selected, just close inline search mode
		m.inlineSearchMode = false

	case "esc":
		// Clear filter and exit inline search mode
		m.inlineSearchMode = false
		m.inlineSearchQuery = ""
		m.inlineSearchSuggestion = ""
		if !inTableView {
			m.selectedRow = 0
			m.scrollOffset = 0
		}

	case "tab":
		// Accept autocomplete suggestion
		if m.inlineSearchSuggestion != "" {
			m.inlineSearchQuery += m.inlineSearchSuggestion
			m.inlineSearchSuggestion = getAutocompleteSuggestion(m.inlineSearchQuery)
		}

	case "backspace":
		if len(m.inlineSearchQuery) > 0 {
			m.inlineSearchQuery = m.inlineSearchQuery[:len(m.inlineSearchQuery)-1]
			m.inlineSearchSuggestion = getAutocompleteSuggestion(m.inlineSearchQuery)

			if inTableView {
				// Apply filter in real-time as user types (removes characters)
				m.tableData = filterTableData(m.allTableData, m.inlineSearchQuery)
				m.selectedDataRow = 0
				m.tableDataScrollY = 0
			} else {
				m.selectedRow = 0
				m.scrollOffset = 0
			}
		}

	case "down":
		// Move down in filtered table list
		displayTables := m.tables
		if m.inlineSearchQuery != "" {
			displayTables = filterTables(m.tables, m.inlineSearchQuery)
		}
		if m.selectedRow < len(displayTables)-1 {
			m.selectedRow++
		}

	case "up":
		// Move up in filtered table list
		if m.selectedRow > 0 {
			m.selectedRow--
		}

	case "ctrl+c":
		return m, tea.Quit

	default:
		// Append to inline search query (only single characters)
		if len(msg.String()) == 1 {
			m.inlineSearchQuery += msg.String()
			m.inlineSearchSuggestion = getAutocompleteSuggestion(m.inlineSearchQuery)

			if inTableView {
				// Apply filter in real-time as user types
				m.tableData = filterTableData(m.allTableData, m.inlineSearchQuery)
				m.selectedDataRow = 0
				m.tableDataScrollY = 0
			} else {
				m.selectedRow = 0
				m.scrollOffset = 0
			}
		}
	}

	return m, nil
}

// handleNormalModeKeys handles keyboard input in normal mode
func (m Model) handleNormalModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Check key bindings
	if cmd, ok := getCommand(key, m.keyBindings.Normal); ok {
		switch cmd {
		case CommandOpenTable:
			// Open table view for selected table
			if len(m.tables) > 0 && m.selectedRow < len(m.tables) {
				selectedTable := m.tables[m.selectedRow]
				m.tableViewMode = true
				m.currentViewTable = &selectedTable
				m.tableDataLoading = true
				m.tableDataError = ""
				m.tableDataOffset = 0 // Start from first page
				return m, fetchTableData(m.driver, selectedTable.SchemaName, selectedTable.TableName, 0)
			}

		case CommandHistoryBack:
			return m.navigateBack()

		case CommandHistoryForward:
			return m.navigateForward()

		case CommandInlineSearch:
			m.inlineSearchMode = true
			m.inlineSearchQuery = ""
			m.inlineSearchSuggestion = ""

		case CommandOpenSearch:
			m.searchMode = true
			m.searchQuery = ""
			m.searchSuggestion = ""
			m.filteredTables = m.allEntities
			m.searchSelected = 0

		case CommandOpenSQLQuery:
			m.sqlQueryMode = true
			m.sqlQueryInput = ""
			m.sqlQueryCursor = 0

		case CommandQuit:
			return m, tea.Quit

		case CommandOpenCommandMode:
			m.commandMode = true
			m.commandBuffer = ":"

		case CommandNavigateDown:
			// Move down in table list
			if m.selectedRow < len(m.tables)-1 {
				m.selectedRow++
			}

		case CommandNavigateUp:
			// Move up in table list
			if m.selectedRow > 0 {
				m.selectedRow--
			}

		case CommandGoToTop:
			m.selectedRow = 0
			m.scrollOffset = 0

		case CommandGoToBottom:
			if len(m.tables) > 0 {
				m.selectedRow = len(m.tables) - 1
			}

		case CommandPageDown:
			// Move down by half the visible rows (like vim Ctrl+D)
			availableHeight := m.height - 4
			if availableHeight < 10 {
				availableHeight = 10
			}
			tablesBoxHeight := availableHeight - 8
			if tablesBoxHeight < 5 {
				tablesBoxHeight = 5
			}
			visibleRows := tablesBoxHeight - 2
			if visibleRows < 1 {
				visibleRows = 1
			}
			jumpSize := visibleRows / 2
			if jumpSize < 1 {
				jumpSize = 1
			}

			m = m.debugLog("PageDown (home): visibleRows=%d, jumpSize=%d, currentRow=%d", visibleRows, jumpSize, m.selectedRow)

			// Move cursor down by jumpSize
			newRow := m.selectedRow + jumpSize
			if newRow >= len(m.tables) {
				newRow = len(m.tables) - 1
			}
			if newRow < 0 {
				newRow = 0
			}
			m.selectedRow = newRow

			m = m.debugLog("PageDown (home): newRow=%d, scrollOffset=%d", m.selectedRow, m.scrollOffset)

		case CommandPageUp:
			// Move up by half the visible rows (like vim Ctrl+U)
			availableHeight := m.height - 4
			if availableHeight < 10 {
				availableHeight = 10
			}
			tablesBoxHeight := availableHeight - 8
			if tablesBoxHeight < 5 {
				tablesBoxHeight = 5
			}
			visibleRows := tablesBoxHeight - 2
			if visibleRows < 1 {
				visibleRows = 1
			}
			jumpSize := visibleRows / 2
			if jumpSize < 1 {
				jumpSize = 1
			}

			m = m.debugLog("PageUp (home): visibleRows=%d, jumpSize=%d, currentRow=%d", visibleRows, jumpSize, m.selectedRow)

			// Move cursor up by jumpSize
			newRow := m.selectedRow - jumpSize
			if newRow < 0 {
				newRow = 0
			}
			m.selectedRow = newRow

			m = m.debugLog("PageUp (home): newRow=%d, scrollOffset=%d", m.selectedRow, m.scrollOffset)
		}
	}

	return m, nil
}

// handleTableViewModeKeys handles keyboard input in table view mode
func (m Model) handleTableViewModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle : key for command mode (to allow :w for batch updates)
	if key == ":" {
		m.commandMode = true
		m.commandBuffer = ":"
		return m, nil
	}

	// Check key bindings for commands
	if cmd, ok := getCommand(key, m.keyBindings.TableView); ok {
		switch cmd {
		case CommandHistoryBack:
			return m.navigateBack()

		case CommandHistoryForward:
			return m.navigateForward()

		case CommandBack:
			// Clear content filter if active, otherwise exit table view mode
			if m.tableContentFilter != "" {
				m.tableContentFilter = ""
				m.tableData = m.allTableData
				m.selectedDataRow = 0
				m.tableDataScrollY = 0
			} else {
				// Exit table view mode
				m.tableViewMode = false
				m.currentViewTable = nil
				m.tableColumns = nil
				m.tableData = nil
				m.allTableData = nil
				m.tableDataScrollX = 0
				m.tableDataScrollY = 0
				m.selectedDataRow = 0
				m.selectedDataCol = 0
				m.tableDataOffset = 0
				m.isCustomQuery = false
				m.executedSQLQuery = ""
				m.tableContentFilter = ""
			}

		case CommandQuit:
			return m, tea.Quit

		case CommandOpenSQLQuery:
			// Open SQL query mode - show current query or allow new query
			if m.isCustomQuery && m.executedSQLQuery != "" {
				m.sqlQueryMode = true
				m.sqlQueryInput = m.executedSQLQuery
				m.sqlQueryCursor = len([]rune(m.sqlQueryInput))
			} else if m.currentViewTable != nil {
				query := fmt.Sprintf("SELECT * FROM \"%s\".\"%s\" LIMIT 500 OFFSET %d",
					m.currentViewTable.SchemaName,
					m.currentViewTable.TableName,
					m.tableDataOffset)
				m.sqlQueryMode = true
				m.sqlQueryInput = query
				m.sqlQueryCursor = len([]rune(query))
			} else {
				m.sqlQueryMode = true
				m.sqlQueryInput = ""
				m.sqlQueryCursor = 0
			}

		case CommandFilterContent:
			m.inlineSearchMode = true
			m.inlineSearchQuery = m.tableContentFilter

		case CommandOpenSearch:
			m.searchMode = true
			m.searchQuery = ""
			m.searchSuggestion = ""
			m.filteredTables = m.allEntities
			m.searchSelected = 0

		case CommandNextPage:
			// Next page - load next 500 rows
			if m.currentViewTable != nil && len(m.tableData) == 500 {
				m.tableDataOffset += 500
				m.tableDataLoading = true
				return m, fetchTableData(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, m.tableDataOffset)
			}

		case CommandPreviousPage:
			// Previous page - load previous 500 rows
			if m.currentViewTable != nil && m.tableDataOffset > 0 {
				m.tableDataOffset -= 500
				if m.tableDataOffset < 0 {
					m.tableDataOffset = 0
				}
				m.tableDataLoading = true
				return m, fetchTableData(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, m.tableDataOffset)
			}

		case CommandNavigateLeft:
			// Move to previous column
			if m.selectedDataCol > 0 {
				m.selectedDataCol--
				m = m.adjustTableDataHorizontalScroll()
			}

		case CommandNavigateRight:
			// Move to next column
			if len(m.tableColumns) > 0 && m.selectedDataCol < len(m.tableColumns)-1 {
				m.selectedDataCol++
				m = m.adjustTableDataHorizontalScroll()
			}

		case CommandNavigateWordForward:
			// Move to next column (word forward)
			if len(m.tableColumns) > 0 && m.selectedDataCol < len(m.tableColumns)-1 {
				m.selectedDataCol++
				m = m.adjustTableDataHorizontalScroll()
			}

		case CommandNavigateWordBackward:
			// Move to previous column (word backward)
			if m.selectedDataCol > 0 {
				m.selectedDataCol--
				m = m.adjustTableDataHorizontalScroll()
			}

		case CommandNavigateDown:
			// Move cursor down
			if m.selectedDataRow < len(m.tableData)-1 {
				m.selectedDataRow++
				m = m.adjustTableDataScroll()
			}

		case CommandNavigateUp:
			// Move cursor up
			if m.selectedDataRow > 0 {
				m.selectedDataRow--
				m = m.adjustTableDataScroll()
			}

		case CommandGoToTop:
			m.selectedDataRow = 0
			m.tableDataScrollY = 0

		case CommandGoToBottom:
			if len(m.tableData) > 0 {
				m.selectedDataRow = len(m.tableData) - 1
				m = m.adjustTableDataScroll()
			}

		case CommandOpenCellPopup:
			// Show cell value in popup
			if len(m.tableData) > 0 && m.selectedDataRow < len(m.tableData) {
				row := m.tableData[m.selectedDataRow]
				if m.selectedDataCol < len(row) {
					m = m.openCellValuePopup(row[m.selectedDataCol])
				}
			}

		case CommandEditCell:
			// Edit cell value in popup
			if len(m.tableData) > 0 && m.selectedDataRow < len(m.tableData) {
				row := m.tableData[m.selectedDataRow]
				if m.selectedDataCol < len(row) {
					m.cellEditMode = true
					m.cellEditValue = row[m.selectedDataCol]
					m.cellEditCursor = len([]rune(m.cellEditValue))
					m.cellEditRowIdx = m.selectedDataRow
					m.cellEditColIdx = m.selectedDataCol
					m.cellEditCommandMode = false
					m.cellEditCommand = ""
				}
			}

		case CommandOpenRecordView:
			// Show entire record in popup
			if len(m.tableData) > 0 && m.selectedDataRow < len(m.tableData) {
				m = m.openRecordViewPopup()
			}

		case CommandCopyCellValue:
			// Copy current cell value to clipboard
			if len(m.tableData) > 0 && m.selectedDataRow < len(m.tableData) {
				row := m.tableData[m.selectedDataRow]
				if m.selectedDataCol < len(row) {
					m = m.copyCellToClipboard(row[m.selectedDataCol])
					return m, clearClipboardMessage()
				}
			}

		case CommandCopyRow:
			// Copy entire row to clipboard
			if len(m.tableData) > 0 && m.selectedDataRow < len(m.tableData) {
				m = m.copyRowToClipboard()
				return m, clearClipboardMessage()
			}

		case CommandPageDown:
			// Move down by half the visible rows (like vim Ctrl+D)
			availableHeight := m.height - 14
			if availableHeight < 5 {
				availableHeight = 5
			}
			visibleRows := availableHeight - 2
			if visibleRows < 1 {
				visibleRows = 1
			}
			jumpSize := visibleRows / 2
			if jumpSize < 1 {
				jumpSize = 1
			}

			m = m.debugLog("PageDown: visibleRows=%d, jumpSize=%d, currentRow=%d", visibleRows, jumpSize, m.selectedDataRow)

			// Move cursor down by jumpSize
			newRow := m.selectedDataRow + jumpSize
			if newRow >= len(m.tableData) {
				newRow = len(m.tableData) - 1
			}
			if newRow < 0 {
				newRow = 0
			}
			m.selectedDataRow = newRow
			m = m.adjustTableDataScroll()

			m = m.debugLog("PageDown: newRow=%d, scrollY=%d", m.selectedDataRow, m.tableDataScrollY)

		case CommandPageUp:
			// Move up by half the visible rows (like vim Ctrl+U)
			availableHeight := m.height - 14
			if availableHeight < 5 {
				availableHeight = 5
			}
			visibleRows := availableHeight - 2
			if visibleRows < 1 {
				visibleRows = 1
			}
			jumpSize := visibleRows / 2
			if jumpSize < 1 {
				jumpSize = 1
			}

			m = m.debugLog("PageUp: visibleRows=%d, jumpSize=%d, currentRow=%d", visibleRows, jumpSize, m.selectedDataRow)

			// Move cursor up by jumpSize
			newRow := m.selectedDataRow - jumpSize
			if newRow < 0 {
				newRow = 0
			}
			m.selectedDataRow = newRow
			m = m.adjustTableDataScroll()

			m = m.debugLog("PageUp: newRow=%d, scrollY=%d", m.selectedDataRow, m.tableDataScrollY)
		}
	}

	return m, nil
}

// handleCellValuePopupKeys handles keyboard input in cell value popup mode
func (m Model) handleCellValuePopupKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Check key bindings for commands
	if cmd, ok := getCommand(key, m.keyBindings.CellPopup); ok {
		switch cmd {
		case CommandCancel:
			// Close popup
			m.cellValuePopupMode = false
			m.cellValuePopupContent = ""
			m.cellValuePopupIsJSON = false
			m.cellValuePopupTree = nil
			m.cellValuePopupScroll = 0
			m.cellValuePopupSelected = 0

		case CommandNavigateDown:
			// Scroll down or navigate to next node
			if m.cellValuePopupIsJSON && len(m.cellValuePopupTree) > 0 {
				if m.cellValuePopupSelected < len(m.cellValuePopupTree)-1 {
					m.cellValuePopupSelected++
				}
			} else {
				m.cellValuePopupScroll++
			}

		case CommandNavigateUp:
			// Scroll up or navigate to previous node
			if m.cellValuePopupIsJSON && len(m.cellValuePopupTree) > 0 {
				if m.cellValuePopupSelected > 0 {
					m.cellValuePopupSelected--
				}
			} else {
				if m.cellValuePopupScroll > 0 {
					m.cellValuePopupScroll--
				}
			}

		case CommandNavigateLeft:
			// Collapse node in JSON tree
			if m.cellValuePopupIsJSON && len(m.cellValuePopupTree) > 0 {
				m = m.collapseJSONNode()
			}

		case CommandNavigateRight, CommandToggleJSONNode:
			// Expand node in JSON tree
			if m.cellValuePopupIsJSON && len(m.cellValuePopupTree) > 0 {
				m = m.expandJSONNode()
			}

		case CommandGoToTop:
			m.cellValuePopupSelected = 0
			m.cellValuePopupScroll = 0

		case CommandGoToBottom:
			if m.cellValuePopupIsJSON && len(m.cellValuePopupTree) > 0 {
				m.cellValuePopupSelected = len(m.cellValuePopupTree) - 1
			}

		case CommandQuit:
			return m, tea.Quit
		}
	}

	return m, nil
}

// handleRecordViewKeys handles keyboard input in record view popup mode
func (m Model) handleRecordViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Check key bindings for commands
	if cmd, ok := getCommand(key, m.keyBindings.RecordView); ok {
		switch cmd {
		case CommandCancel:
			// Close record view popup
			m.recordViewMode = false
			m.recordViewData = nil
			m.recordViewColumns = nil
			m.recordViewSelected = 0
			m.recordViewScroll = 0

		case CommandOpenCellPopup:
			// Open cell value popup for the selected field
			if m.recordViewSelected < len(m.recordViewData) {
				m = m.openCellValuePopup(m.recordViewData[m.recordViewSelected])
			}

		case CommandNavigateDown:
			// Navigate down in field list
			if m.recordViewSelected < len(m.recordViewData)-1 {
				m.recordViewSelected++
			}

		case CommandNavigateUp:
			// Navigate up in field list
			if m.recordViewSelected > 0 {
				m.recordViewSelected--
			}

		case CommandGoToTop:
			m.recordViewSelected = 0
			m.recordViewScroll = 0

		case CommandGoToBottom:
			if len(m.recordViewData) > 0 {
				m.recordViewSelected = len(m.recordViewData) - 1
			}

		case CommandQuit:
			return m, tea.Quit
		}
	}

	return m, nil
}

// handleSQLQueryModeKeys handles keyboard input in SQL query mode
func (m Model) handleSQLQueryModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	runes := []rune(m.sqlQueryInput)

	// Check key bindings for commands
	if cmd, ok := getCommand(key, m.keyBindings.SQLQuery); ok {
		switch cmd {
		case CommandConfirm:
			// Execute the SQL query
			if m.sqlQueryInput != "" {
				m.sqlQueryMode = false
				m.tableDataLoading = true
				return m, executeSQLQuery(m.driver, m.sqlQueryInput)
			}

		case CommandCancel:
			// Cancel SQL query mode
			m.sqlQueryMode = false
			m.sqlQueryInput = ""
			m.sqlQueryCursor = 0

		case CommandCursorLeft:
			if m.sqlQueryCursor > 0 {
				m.sqlQueryCursor--
			}

		case CommandCursorRight:
			if m.sqlQueryCursor < len(runes) {
				m.sqlQueryCursor++
			}

		case CommandCursorHome:
			m.sqlQueryCursor = 0

		case CommandCursorEnd:
			m.sqlQueryCursor = len(runes)

		case CommandBackspace:
			if m.sqlQueryCursor > 0 && m.sqlQueryCursor <= len(runes) {
				runes = append(runes[:m.sqlQueryCursor-1], runes[m.sqlQueryCursor:]...)
				m.sqlQueryInput = string(runes)
				m.sqlQueryCursor--
			}

		case CommandDeleteChar:
			if m.sqlQueryCursor >= 0 && m.sqlQueryCursor < len(runes) {
				runes = append(runes[:m.sqlQueryCursor], runes[m.sqlQueryCursor+1:]...)
				m.sqlQueryInput = string(runes)
			}

		case CommandQuit:
			return m, tea.Quit
		}
	} else {
		// Handle text input (default case)
		keyStr := msg.String()
		if len(keyStr) == 1 || keyStr == "space" || keyStr == "tab" {
			var charToInsert rune
			if keyStr == "space" {
				charToInsert = ' '
			} else if keyStr == "tab" {
				charToInsert = '\t'
			} else {
				charToInsert = []rune(keyStr)[0]
			}
			if m.sqlQueryCursor >= 0 && m.sqlQueryCursor <= len(runes) {
				runes = append(runes[:m.sqlQueryCursor], append([]rune{charToInsert}, runes[m.sqlQueryCursor:]...)...)
				m.sqlQueryInput = string(runes)
				m.sqlQueryCursor++
			}
		}
	}

	return m, nil
}

// handleCellEditModeKeys handles keyboard input in cell edit mode
func (m Model) handleCellEditModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// If in command mode within edit mode
	if m.cellEditCommandMode {
		// Check for command mode bindings
		if cmd, ok := getCommand(key, m.keyBindings.CommandMode); ok {
			switch cmd {
			case CommandConfirm:
				// Execute command
				if m.cellEditCommand == ":w" || m.cellEditCommand == ":w " {
					// Batch update all cells in buffer
					return m.batchUpdateCells()
				} else if m.cellEditCommand == ":q" || m.cellEditCommand == ":q " {
					// Quit without saving - clear buffer
					m.cellEditMode = false
					m.cellEditCommandMode = false
					m.cellEditCommand = ""
					m.cellEditBuffer = make(map[string]string)
					m.cellEditBufferCount = 0
				}
				m.cellEditCommandMode = false
				m.cellEditCommand = ""

			case CommandCancel:
				m.cellEditCommandMode = false
				m.cellEditCommand = ""

			case CommandQuit:
				return m, tea.Quit
			}
		} else {
			// Handle text input for command
			if key == "backspace" {
				if len(m.cellEditCommand) > 0 {
					m.cellEditCommand = m.cellEditCommand[:len(m.cellEditCommand)-1]
				}
			} else if len(key) == 1 || key == "space" {
				if key == "space" {
					m.cellEditCommand += " "
				} else {
					m.cellEditCommand += key
				}
			}
		}
		return m, nil
	}

	// Normal edit mode (not command mode)
	runes := []rune(m.cellEditValue)

	// Handle special keys first
	switch key {
	case "enter":
		// Enter: Save current cell to buffer and close popup
		return m.saveCellToBuffer()

	case "ctrl+enter":
		// Ctrl+Enter: Insert newline
		runes = append(runes[:m.cellEditCursor], append([]rune{'\n'}, runes[m.cellEditCursor:]...)...)
		m.cellEditValue = string(runes)
		m.cellEditCursor++
		return m, nil
	}

	// Check key bindings for commands
	if cmd, ok := getCommand(key, m.keyBindings.CellEdit); ok {
		switch cmd {
		case CommandCancel:
			// Esc: Close popup without saving current cell (buffer remains)
			m.cellEditMode = false
			m.cellEditValue = ""
			m.cellEditCursor = 0
			m.cellEditCommandMode = false
			m.cellEditCommand = ""

		case CommandOpenCommandMode:
			m.cellEditCommandMode = true
			m.cellEditCommand = ":"

		case CommandCursorLeft:
			if m.cellEditCursor > 0 {
				m.cellEditCursor--
			}

		case CommandCursorRight:
			if m.cellEditCursor < len(runes) {
				m.cellEditCursor++
			}

		case CommandCursorHome:
			m.cellEditCursor = 0

		case CommandCursorEnd:
			m.cellEditCursor = len(runes)

		case CommandBackspace:
			if m.cellEditCursor > 0 && m.cellEditCursor <= len(runes) {
				runes = append(runes[:m.cellEditCursor-1], runes[m.cellEditCursor:]...)
				m.cellEditValue = string(runes)
				m.cellEditCursor--
			}

		case CommandDeleteChar:
			if m.cellEditCursor >= 0 && m.cellEditCursor < len(runes) {
				runes = append(runes[:m.cellEditCursor], runes[m.cellEditCursor+1:]...)
				m.cellEditValue = string(runes)
			}

		case CommandQuit:
			return m, tea.Quit
		}
	} else {
		// Handle text input (default case)
		if len(key) == 1 || key == "space" || key == "tab" {
			var charToInsert rune
			if key == "space" {
				charToInsert = ' '
			} else if key == "tab" {
				charToInsert = '\t'
			} else {
				charToInsert = []rune(key)[0]
			}
			if m.cellEditCursor >= 0 && m.cellEditCursor <= len(runes) {
				runes = append(runes[:m.cellEditCursor], append([]rune{charToInsert}, runes[m.cellEditCursor:]...)...)
				m.cellEditValue = string(runes)
				m.cellEditCursor++
			}
		}
	}

	return m, nil
}

// handleDebugPanelKeys handles keyboard input when debug panel is focused
func (m Model) handleDebugPanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Check key bindings for commands
	if cmd, ok := getCommand(key, m.keyBindings.DebugPanel); ok {
		switch cmd {
		case CommandCancel:
			// Unfocus debug panel
			m.debugPanelFocused = false

		case CommandNavigateDown:
			m = m.debugNavigateDown()

		case CommandNavigateUp:
			m = m.debugNavigateUp()

		case CommandSwitchDebugSection:
			m = m.debugSwitchSection()

		case CommandConfirm:
			m = m.debugOpenDetail()

		case CommandGoToTop:
			m.debugSelectedLog = 0

		case CommandGoToBottom:
			if len(m.debugLogs) > 0 {
				m.debugSelectedLog = len(m.debugLogs) - 1
			}

		case CommandQuit:
			return m, tea.Quit
		}
	}

	return m, nil
}

// handleDebugDetailKeys handles keyboard input in debug detail popup
func (m Model) handleDebugDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Check for command in debug detail bindings
	if cmd, ok := getCommand(key, m.keyBindings.DebugDetail); ok {
		switch cmd {
		case CommandCancel:
			// Close detail popup
			m = m.debugCloseDetail()
			return m, nil

		case CommandQuit:
			return m, tea.Quit
		}
	}

	// Check global bindings (like ctrl+c for quit)
	if cmd, ok := getCommand(key, m.keyBindings.Global); ok {
		switch cmd {
		case CommandQuit:
			return m, tea.Quit
		}
	}

	return m, nil
}
