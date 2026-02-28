package tui

import (
	"encoding/json"
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

	// Handle help popup mode
	if m.helpPopupMode {
		// handleHelpPopupKeys handles keyboard input in help popup mode
		return m.handleHelpPopupKeys(msg)
	}

	// Handle passphrase prompt mode first - highest priority for encryption
	if m.passphrasePromptMode {
		return m.handlePassphrasePromptKeys(msg)
	}

	// Handle key selector mode second - for first-run setup
	if m.keySelectorMode {
		return m.handleKeySelectorKeys(msg)
	}

	// Handle search mode - high priority to allow typing
	if m.searchMode {
		return m.handleSearchModeKeys(msg)
	}

	// Handle connection manager mode
	if m.connManagerMode {
		return m.handleConnectionManagerKeys(msg)
	}

	// Handle connection input mode
	if m.connInputMode {
		return m.handleConnectionInputKeys(msg)
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

func (m Model) handleHelpPopupKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	maxScroll := m.getHelpMaxScroll()

	switch key {
	case "?", "esc":
		m.helpPopupMode = false
		m.helpPopupScroll = 0 // Reset scroll when closing
	case "j", "down":
		m.helpPopupScroll++
		if m.helpPopupScroll > maxScroll {
			m.helpPopupScroll = maxScroll
		}
	case "k", "up":
		m.helpPopupScroll--
		if m.helpPopupScroll < 0 {
			m.helpPopupScroll = 0
		}
	case "g":
		m.helpPopupScroll = 0
	case "G":
		m.helpPopupScroll = maxScroll
	}
	return m, nil
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
				// Fetch both data and schema in parallel
				return m, tea.Batch(
					fetchTableData(m.driver, selectedEntity.SchemaName, selectedEntity.TableName, 0),
					fetchTableSchema(m.driver, selectedEntity.SchemaName, selectedEntity.TableName),
				)
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
			// Batch update all cells in buffer AND/OR batch delete rows
			m.commandMode = false
			m.commandBuffer = ""

			// Check if we have both updates and deletes
			hasUpdates := m.cellEditBufferCount > 0
			hasDeletes := len(m.currentDeletedRows()) > 0

			if hasDeletes {
				return m.batchDeleteRows()
			} else if hasUpdates {
				return m.batchUpdateCells()
			}

			// Nothing to do
			return m, nil
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
			// Fetch both data and schema in parallel
			return m, tea.Batch(
				fetchTableData(m.driver, selectedTable.SchemaName, selectedTable.TableName, 0),
				fetchTableSchema(m.driver, selectedTable.SchemaName, selectedTable.TableName),
			)
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
		if !inTableView {
			// On home screen: switch between table list and schema panel
			// If there's an autocomplete suggestion and schema panel is not focused, accept it
			if m.inlineSearchSuggestion != "" && !m.schemaPanelFocused {
				m.inlineSearchQuery += m.inlineSearchSuggestion
				m.inlineSearchSuggestion = getAutocompleteSuggestion(m.inlineSearchQuery)
			} else {
				// Otherwise, switch focus to schema panel
				m.schemaPanelFocused = !m.schemaPanelFocused
			}
		} else {
			// In table view: accept autocomplete suggestion
			if m.inlineSearchSuggestion != "" {
				m.inlineSearchQuery += m.inlineSearchSuggestion
				m.inlineSearchSuggestion = getAutocompleteSuggestion(m.inlineSearchQuery)
			}
		}

	case "backspace":
		// Don't modify filter if schema panel is focused (unless in table view)
		if len(m.inlineSearchQuery) > 0 && (!m.schemaPanelFocused || inTableView) {
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
				// Trigger schema fetch for first filtered table
				displayTables := m.tables
				if m.inlineSearchQuery != "" {
					displayTables = filterTables(m.tables, m.inlineSearchQuery)
				}
				if len(displayTables) > 0 {
					m.schemaInfoLoading = true
					m.schemaInfo = nil
					return m, fetchSchemaInfo(m.driver, displayTables[0].SchemaName, displayTables[0].TableName)
				}
			}
		}

	case "down":
		if !inTableView && m.schemaPanelFocused {
			// Navigate within schema panel
			if m.schemaPanelLineCount > 0 && m.schemaPanelSelected < m.schemaPanelLineCount-1 {
				m.schemaPanelSelected++
			}
		} else {
			// Move down in filtered table list
			displayTables := m.tables
			if m.inlineSearchQuery != "" {
				displayTables = filterTables(m.tables, m.inlineSearchQuery)
			}
			if m.selectedRow < len(displayTables)-1 {
				m.selectedRow++
				// Trigger async schema info fetch for new table
				if !inTableView {
					m.schemaInfoLoading = true
					m.schemaInfo = nil
					selectedTable := displayTables[m.selectedRow]
					return m, fetchSchemaInfo(m.driver, selectedTable.SchemaName, selectedTable.TableName)
				}
			}
		}

	case "up":
		if !inTableView && m.schemaPanelFocused {
			// Navigate within schema panel
			if m.schemaPanelSelected > 0 {
				m.schemaPanelSelected--
			}
		} else {
			// Move up in filtered table list
			if m.selectedRow > 0 {
				m.selectedRow--
				// Trigger async schema info fetch for new table
				displayTables := m.tables
				if m.inlineSearchQuery != "" {
					displayTables = filterTables(m.tables, m.inlineSearchQuery)
				}
				if !inTableView {
					m.schemaInfoLoading = true
					m.schemaInfo = nil
					selectedTable := displayTables[m.selectedRow]
					return m, fetchSchemaInfo(m.driver, selectedTable.SchemaName, selectedTable.TableName)
				}
			}
		}

	case "left":
		// h/left: scroll left in schema panel (no-op for now, could be used for horizontal scroll)
		// Or navigate to table list if in schema panel
		if !inTableView && m.schemaPanelFocused {
			// Could implement horizontal scroll here if needed
		}

	case "right":
		// l/right: scroll right in schema panel (no-op for now)
		// Or could be used for expanding/collapsing items
		if !inTableView && m.schemaPanelFocused {
			// Could implement horizontal scroll here if needed
		}

	case "ctrl+c":
		return m, tea.Quit

	default:
		// Append to inline search query (only single characters)
		// Don't append if schema panel is focused (navigating schema)
		if len(msg.String()) == 1 && (!m.schemaPanelFocused || inTableView) {
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
				// Trigger schema fetch for first filtered table
				displayTables := filterTables(m.tables, m.inlineSearchQuery)
				if len(displayTables) > 0 {
					m.schemaInfoLoading = true
					m.schemaInfo = nil
					return m, fetchSchemaInfo(m.driver, displayTables[0].SchemaName, displayTables[0].TableName)
				}
			}
		}
	}

	return m, nil
}

// handleNormalModeKeys handles keyboard input in normal mode
func (m Model) handleNormalModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// If disconnected, treat this as connection manager in main view (not popup)
	// Use the same insert/normal mode behavior as the connection manager popup
	if m.connectionStatus == Disconnected {
		// Temporarily set connManagerMode to true and delegate to connection manager handler
		m.connManagerMode = true
		m2, cmd := m.handleConnectionManagerKeys(msg)
		// If connection manager was closed by the handler, keep it open since we're in main view
		if !m2.connManagerMode {
			// User pressed 'q' or 'esc' in normal mode - interpret as quit
			if key == "q" || (key == "esc" && !m.connManagerInsertMode) {
				return m2, tea.Quit
			}
			// Otherwise, keep the connection manager conceptually "open" (it's the main view)
			m2.connManagerMode = false // Keep it as main view, not popup
			return m2, cmd
		}
		m2.connManagerMode = false // Reset since we're not in popup mode
		return m2, cmd
	}

	// Handle Tab to switch between panels
	if key == "tab" {
		m.schemaPanelFocused = !m.schemaPanelFocused
		return m, nil
	}

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
				// Fetch both data and schema in parallel
				return m, tea.Batch(
					fetchTableData(m.driver, selectedTable.SchemaName, selectedTable.TableName, 0),
					fetchTableSchema(m.driver, selectedTable.SchemaName, selectedTable.TableName),
				)
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
			m.sqlQueryInput.SetValue("")
			m.sqlQueryInput.Focus()
			m = m.updateSQLHistorySuggestions()

		case CommandOpenConnectionManager:
			// Open connection manager popup
			m.connManagerMode = true
			m.connManagerFilter = m.makeConnectionManagerFilter()
			m.connManagerSelected = 0
			m.connManagerScroll = 0
			m.connManagerInsertMode = true // Start in insert mode
			if m.connHistory != nil {
				m.filteredConnections = m.connHistory.GetAll()
			}

		case CommandQuit:
			return m, tea.Quit

		case CommandOpenCommandMode:
			m.commandMode = true
			m.commandBuffer = ":"

		case CommandNavigateDown:
			if m.schemaPanelFocused {
				// Schema panel navigation - move cursor down
				// Bounds will be checked in render, but we track line count
				if m.schemaPanelLineCount > 0 && m.schemaPanelSelected < m.schemaPanelLineCount-1 {
					m.schemaPanelSelected++
				}
			} else {
				// Move down in table list
				if m.selectedRow < len(m.tables)-1 {
					m.selectedRow++
					// Reset schema panel cursor when switching tables
					m.schemaPanelSelected = 0
					m.schemaPanelScroll = 0
					m = m.updateSchemaPanelLineCount()

					// Trigger async schema info fetch for new table
					m.schemaInfoLoading = true
					m.schemaInfo = nil
					selectedTable := m.tables[m.selectedRow]
					return m, fetchSchemaInfo(m.driver, selectedTable.SchemaName, selectedTable.TableName)
				}
			}

		case CommandNavigateUp:
			if m.schemaPanelFocused {
				// Schema panel navigation - move cursor up
				if m.schemaPanelSelected > 0 {
					m.schemaPanelSelected--
				}
			} else {
				// Move up in table list
				if m.selectedRow > 0 {
					m.selectedRow--
					// Reset schema panel cursor when switching tables
					m.schemaPanelSelected = 0
					m.schemaPanelScroll = 0
					m = m.updateSchemaPanelLineCount()

					// Trigger async schema info fetch for new table
					m.schemaInfoLoading = true
					m.schemaInfo = nil
					selectedTable := m.tables[m.selectedRow]
					return m, fetchSchemaInfo(m.driver, selectedTable.SchemaName, selectedTable.TableName)
				}
			}

		case CommandGoToTop:
			if m.schemaPanelFocused {
				m.schemaPanelSelected = 0
				m.schemaPanelScroll = 0
			} else {
				m.selectedRow = 0
				m.scrollOffset = 0
			}

		case CommandGoToBottom:
			if m.schemaPanelFocused {
				// Will be bounded by render function
				m.schemaPanelSelected = 1000
			} else {
				if len(m.tables) > 0 {
					m.selectedRow = len(m.tables) - 1
				}
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

	// Handle ? key for help popup
	if key == "?" {
		m.helpPopupMode = !m.helpPopupMode
		return m, nil
	}

	// Handle : key for command mode (to allow :w for batch updates)
	if key == ":" {
		m.commandMode = true
		m.commandBuffer = ":"
		return m, nil
	}

	// Handle dd key sequence (vim-style row deletion)
	if key == "d" {
		if m.lastKeyPress == 'd' {
			if len(m.tableData) > 0 && m.selectedDataRow < len(m.tableData) {
				tableKey := m.currentTableKey()
				if m.deletedRows[tableKey] == nil {
					m.deletedRows[tableKey] = make(map[int]bool)
				}
				if m.deletedRows[tableKey][m.selectedDataRow] {
					delete(m.deletedRows[tableKey], m.selectedDataRow)
				} else {
					m.deletedRows[tableKey][m.selectedDataRow] = true
				}
			}
			m.lastKeyPress = 0
			return m, nil
		}
		// First 'd' pressed - store it
		m.lastKeyPress = 'd'
		return m, nil
	}

	// Reset lastKeyPress if any other key is pressed
	if m.lastKeyPress == 'd' {
		m.lastKeyPress = 0
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
			if m.isCustomQuery && m.executedSQLQuery != "" {
				m.sqlQueryMode = true
				m.sqlQueryInput.SetValue(m.executedSQLQuery)
			} else if m.currentViewTable != nil {
				query := fmt.Sprintf("SELECT * FROM \"%s\".\"%s\" LIMIT 500 OFFSET %d",
					m.currentViewTable.SchemaName,
					m.currentViewTable.TableName,
					m.tableDataOffset)
				m.sqlQueryMode = true
				m.sqlQueryInput.SetValue(query)
			} else {
				m.sqlQueryMode = true
				m.sqlQueryInput.SetValue("")
			}
			m.sqlQueryInput.Focus()
			m = m.updateSQLHistorySuggestions()

		case CommandFilterContent:
			m.inlineSearchMode = true
			m.inlineSearchQuery = m.tableContentFilter

		case CommandOpenConnectionManager:
			// Open connection manager popup
			m.connManagerMode = true
			m.connManagerFilter = m.makeConnectionManagerFilter()
			m.connManagerSelected = 0
			m.connManagerScroll = 0
			m.connManagerInsertMode = true // Start in insert mode
			if m.connHistory != nil {
				m.filteredConnections = m.connHistory.GetAll()
			}

		case CommandOpenSearch:
			m.searchMode = true
			m.searchQuery = ""
			m.searchSuggestion = ""
			m.filteredTables = m.allEntities
			m.searchSelected = 0

		case CommandSortAsc:
			if m.currentViewTable != nil && m.selectedDataCol < len(m.tableColumns) {
				col := m.tableColumns[m.selectedDataCol]
				m.restoreCursor = true
				m.savedCursorRow = m.selectedDataRow
				m.savedCursorCol = m.selectedDataCol
				m.tableDataOffset = 0
				m.tableDataLoading = true
				cur := m.currentSort()
				if cur.column == col && cur.order == SortAsc {
					m.clearSort()
					return m, fetchTableData(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, 0)
				}
				m.setSort(col, SortAsc)
				return m, fetchTableDataSorted(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, col, SortAsc, 0)
			}

		case CommandSortDesc:
			if m.currentViewTable != nil && m.selectedDataCol < len(m.tableColumns) {
				col := m.tableColumns[m.selectedDataCol]
				m.restoreCursor = true
				m.savedCursorRow = m.selectedDataRow
				m.savedCursorCol = m.selectedDataCol
				m.tableDataOffset = 0
				m.tableDataLoading = true
				cur := m.currentSort()
				if cur.column == col && cur.order == SortDesc {
					m.clearSort()
					return m, fetchTableData(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, 0)
				}
				m.setSort(col, SortDesc)
				return m, fetchTableDataSorted(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, col, SortDesc, 0)
			}

		case CommandNextPage:
			if m.currentViewTable != nil && len(m.tableData) == 500 {
				m.tableDataOffset += 500
				m.tableDataLoading = true
				cur := m.currentSort()
				if cur.order != SortNone {
					return m, fetchTableDataSorted(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, cur.column, cur.order, m.tableDataOffset)
				}
				return m, fetchTableData(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, m.tableDataOffset)
			}

		case CommandPreviousPage:
			if m.currentViewTable != nil && m.tableDataOffset > 0 {
				m.tableDataOffset -= 500
				if m.tableDataOffset < 0 {
					m.tableDataOffset = 0
				}
				m.tableDataLoading = true
				cur := m.currentSort()
				if cur.order != SortNone {
					return m, fetchTableDataSorted(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, cur.column, cur.order, m.tableDataOffset)
				}
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

		case CommandNavigateLastColumn:
			// Move cursor up
			if m.selectedDataRow > 0 {
				m.selectedDataCol = len(m.tableColumns) - 1
				m = m.adjustTableDataHorizontalScroll()
			}

		case CommandNavigateFirstColumn:
			// Move cursor up
			if m.selectedDataRow > 0 {
				m.selectedDataCol = 0
				m = m.adjustTableDataHorizontalScroll()
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
					m.cellValuePopupStack = []CellValuePopupSnapshot{} // Clear stack when opening from table view
					m = m.openCellValuePopup(row[m.selectedDataCol])
				}
			}

		case CommandEditCell:
			if len(m.tableData) > 0 && m.selectedDataRow < len(m.tableData) {
				row := m.tableData[m.selectedDataRow]
				if m.selectedDataCol < len(row) {
					m.cellEditMode = true
					m.cellEditTextarea = m.makeCellEditTextarea()
					m.cellEditTextarea.SetValue(row[m.selectedDataCol])
					m.cellEditTextarea.Focus()
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
			// Check if there's a parent popup to return to
			if len(m.cellValuePopupStack) > 0 {
				// Pop the previous state from the stack
				prevState := m.cellValuePopupStack[len(m.cellValuePopupStack)-1]
				m.cellValuePopupStack = m.cellValuePopupStack[:len(m.cellValuePopupStack)-1]

				// Restore the previous popup state
				m.cellValuePopupContent = prevState.content
				m.cellValuePopupIsJSON = prevState.isJSON
				m.cellValuePopupTree = prevState.tree
				m.cellValuePopupScroll = prevState.scroll
				m.cellValuePopupSelected = prevState.selected
			} else {
				// No parent popup, close completely
				m.cellValuePopupMode = false
				m.cellValuePopupContent = ""
				m.cellValuePopupIsJSON = false
				m.cellValuePopupTree = nil
				m.cellValuePopupScroll = 0
				m.cellValuePopupSelected = 0
			}

		case CommandOpenCellPopup:
			// Open a new cell value popup for the selected JSON node
			if m.cellValuePopupIsJSON && len(m.cellValuePopupTree) > 0 {
				if m.cellValuePopupSelected < len(m.cellValuePopupTree) {
					node := m.cellValuePopupTree[m.cellValuePopupSelected]

					// Convert value to string, handling JSON objects/arrays properly
					var valueStr string
					switch node.Type {
					case "object", "array":
						// Marshal nested JSON back to string format
						jsonBytes, err := json.Marshal(node.Value)
						if err != nil {
							valueStr = fmt.Sprintf("%v", node.Value)
						} else {
							valueStr = string(jsonBytes)
						}
					default:
						// For primitives, use fmt.Sprintf
						valueStr = fmt.Sprintf("%v", node.Value)
					}

					// Save current state to stack before opening new popup
					snapshot := CellValuePopupSnapshot{
						content:  m.cellValuePopupContent,
						isJSON:   m.cellValuePopupIsJSON,
						tree:     m.cellValuePopupTree,
						scroll:   m.cellValuePopupScroll,
						selected: m.cellValuePopupSelected,
					}
					m.cellValuePopupStack = append(m.cellValuePopupStack, snapshot)

					// Open new popup with the node value
					m = m.openCellValuePopup(valueStr)
				}
			}

		case CommandCopyCellValue:
			// Copy the appropriate value based on context
			var valueToCopy string
			if m.cellValuePopupIsJSON && len(m.cellValuePopupTree) > 0 {
				// JSON popup: copy the value of the selected node
				if m.cellValuePopupSelected < len(m.cellValuePopupTree) {
					node := m.cellValuePopupTree[m.cellValuePopupSelected]
					valueToCopy = fmt.Sprintf("%v", node.Value)
				}
			} else {
				// Non-JSON popup: copy the entire cell value
				valueToCopy = m.cellValuePopupContent
			}
			m = m.copyValueToClipboard(valueToCopy)
			return m, clearClipboardMessage()

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
				m.cellValuePopupStack = []CellValuePopupSnapshot{} // Clear stack when opening from record view
				m = m.openCellValuePopup(m.recordViewData[m.recordViewSelected])
			}

		case CommandCopyCellValue:
			// Copy the value of the selected field
			if m.recordViewSelected < len(m.recordViewData) {
				m = m.copyValueToClipboard(m.recordViewData[m.recordViewSelected])
				return m, clearClipboardMessage()
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

	// Handle up/down for history suggestions navigation
	if key == "up" || key == "ctrl+k" {
		if m.sqlHistorySuggestionsVisible && len(m.sqlHistorySuggestions) > 0 {
			m.sqlHistorySelected--
			if m.sqlHistorySelected < 0 {
				m.sqlHistorySelected = len(m.sqlHistorySuggestions) - 1
			}
			return m, nil
		}
	} else if key == "down" || key == "ctrl+j" {
		if m.sqlHistorySuggestionsVisible && len(m.sqlHistorySuggestions) > 0 {
			m.sqlHistorySelected++
			if m.sqlHistorySelected >= len(m.sqlHistorySuggestions) {
				m.sqlHistorySelected = 0
			}
			return m, nil
		}
	} else if key == "ctrl+n" {
		if m.sqlHistory != nil {
			m.sqlHistorySuggestionsVisible = !m.sqlHistorySuggestionsVisible
			if m.sqlHistorySuggestionsVisible {
				m.sqlHistorySuggestions = m.sqlHistory.SearchEntries(m.sqlQueryInput.Value())
				if len(m.sqlHistorySuggestions) > 0 {
					m.sqlHistorySelected = 0
				} else {
					m.sqlHistorySelected = -1
				}
			}
			return m, nil
		}
	}

	// Check key bindings for commands
	if cmd, ok := getCommand(key, m.keyBindings.SQLQuery); ok {
		switch cmd {
		case CommandConfirm:
			if m.sqlHistorySuggestionsVisible && m.sqlHistorySelected >= 0 && m.sqlHistorySelected < len(m.sqlHistorySuggestions) {
				m.sqlQueryInput.SetValue(m.sqlHistorySuggestions[m.sqlHistorySelected].Query)
				m.sqlHistorySuggestionsVisible = false
				m.sqlHistorySelected = -1
				return m, nil
			}

			if m.sqlQueryInput.Value() != "" {
				m.sqlQueryMode = false
				m.sqlHistorySuggestionsVisible = false
				m.sqlHistorySelected = -1
				m.tableViewMode = true
				m.tableDataLoading = true
				return m, executeSQLQuery(m.driver, m.sqlQueryInput.Value())
			}
			return m, nil

		case CommandCancel:
			if m.sqlHistorySuggestionsVisible {
				m.sqlHistorySuggestionsVisible = false
				m.sqlHistorySelected = -1
			} else {
				m.sqlQueryMode = false
				m.sqlQueryInput.SetValue("")
			}

		case CommandQuit:
			return m, tea.Quit
		}
	} else {
		// Delegate to textinput for all other keys
		oldValue := m.sqlQueryInput.Value()
		var cmd tea.Cmd
		m.sqlQueryInput, cmd = m.sqlQueryInput.Update(msg)

		// Update suggestions if the value changed
		if m.sqlQueryInput.Value() != oldValue {
			m = m.updateSQLHistorySuggestions()
		}

		return m, cmd
	}

	return m, nil
}

func (m Model) updateSQLHistorySuggestions() Model {
	if m.sqlHistory == nil {
		return m
	}

	m.sqlHistorySuggestions = m.sqlHistory.SearchEntries(m.sqlQueryInput.Value())

	if len(m.sqlHistorySuggestions) > 0 {
		m.sqlHistorySuggestionsVisible = true
		if m.sqlHistorySelected < 0 || m.sqlHistorySelected >= len(m.sqlHistorySuggestions) {
			m.sqlHistorySelected = 0
		}
	} else {
		m.sqlHistorySuggestionsVisible = false
		m.sqlHistorySelected = -1
	}

	return m
}

// handleCellEditModeKeys handles keyboard input in cell edit mode
func (m Model) handleCellEditModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// If in command mode within edit mode
	if m.cellEditCommandMode {
		if cmd, ok := getCommand(key, m.keyBindings.CommandMode); ok {
			switch cmd {
			case CommandConfirm:
				if m.cellEditCommand == ":w" || m.cellEditCommand == ":w " {
					return m.batchUpdateCells()
				} else if m.cellEditCommand == ":q" || m.cellEditCommand == ":q " {
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

	// Handle special keys
	if key == "enter" && !msg.Alt {
		return m.saveCellToBuffer()
	}

	if cmd, ok := getCommand(key, m.keyBindings.CellEdit); ok {
		switch cmd {
		case CommandCancel:
			m.cellEditMode = false
			m.cellEditCommandMode = false
			m.cellEditCommand = ""

		case CommandOpenCommandMode:
			m.cellEditCommandMode = true
			m.cellEditCommand = ":"

		case CommandQuit:
			return m, tea.Quit
		}
		return m, nil
	}

	// Delegate to textarea for all other keys
	var cmd tea.Cmd
	m.cellEditTextarea, cmd = m.cellEditTextarea.Update(msg)
	return m, cmd
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
