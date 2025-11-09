package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/someshkoli/dessertfrog/pkg/driver"
)

// Connection messages
type connectionSuccessMsg struct{}
type connectionFailedMsg struct {
	err error
}

// Tables messages
type tablesLoadedMsg struct {
	tables      []driver.TableSchema
	allEntities []driver.TableSchema
}
type tablesLoadFailedMsg struct {
	err error
}

// Table data messages
type tableDataLoadedMsg struct {
	columns []string
	rows    [][]string
}
type tableDataLoadFailedMsg struct {
	err error
}

// Clipboard messages
type clearClipboardMsgType struct{}

// SQL query messages
type sqlQueryResultMsg struct {
	columns []string
	rows    [][]string
	query   string
}
type sqlQueryFailedMsg struct {
	err   error
	query string
}

type cellUpdateSuccessMsg struct {
	newValue string
}
type cellUpdateFailedMsg struct {
	err error
}

// connectToDatabase attempts to connect to the database
func connectToDatabase(drv driver.Driver) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := drv.Connect(ctx); err != nil {
			return connectionFailedMsg{err: err}
		}
		return connectionSuccessMsg{}
	}
}

// fetchTables fetches the list of tables and all entities from the database
func fetchTables(drv driver.Driver) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Fetch only tables for main view
		tables, err := drv.GetTables(ctx)
		if err != nil {
			return tablesLoadFailedMsg{err: err}
		}

		// Fetch all entities for search
		allEntities, err := drv.GetAllEntities(ctx)
		if err != nil {
			return tablesLoadFailedMsg{err: err}
		}

		return tablesLoadedMsg{
			tables:      tables,
			allEntities: allEntities,
		}
	}
}

// fetchTableData fetches data from a specific table
func fetchTableData(drv driver.Driver, schemaName, tableName string, offset int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Fetch table data with limit of 500 rows and given offset
		columns, rows, err := drv.GetTableData(ctx, schemaName, tableName, 500, offset)
		if err != nil {
			return tableDataLoadFailedMsg{err: err}
		}

		return tableDataLoadedMsg{
			columns: columns,
			rows:    rows,
		}
	}
}

// executeSQLQuery executes a custom SQL query
func executeSQLQuery(drv driver.Driver, query string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Execute the custom SQL query
		columns, rows, err := drv.ExecuteQuery(ctx, query)
		if err != nil {
			return sqlQueryFailedMsg{err: err, query: query}
		}

		return sqlQueryResultMsg{
			columns: columns,
			rows:    rows,
			query:   query,
		}
	}
}

// updateCellValue updates a cell value in the database and refreshes the data
func updateCellValue(drv driver.Driver, tableSchema *driver.TableSchema, columns []string, data [][]string, rowIdx, colIdx int, newValue string, offset int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Get the column name
		if colIdx >= len(columns) {
			return cellUpdateFailedMsg{err: fmt.Errorf("column index out of bounds")}
		}
		columnName := columns[colIdx]

		// Build WHERE clause based on primary key or all columns
		// For simplicity, we'll use all columns to identify the row
		if rowIdx >= len(data) {
			return cellUpdateFailedMsg{err: fmt.Errorf("row index out of bounds")}
		}
		oldRow := data[rowIdx]

		// Execute UPDATE
		err := drv.UpdateCell(ctx, tableSchema.SchemaName, tableSchema.TableName, columns, oldRow, columnName, newValue)
		if err != nil {
			return cellUpdateFailedMsg{err: err}
		}

		// Return success with the new value (don't refresh data)
		return cellUpdateSuccessMsg{
			newValue: newValue,
		}
	}
}

// Init initializes the bubbletea model
func (m Model) Init() tea.Cmd {
	// Start connection attempt
	return connectToDatabase(m.driver)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case connectionSuccessMsg:
		m.connectionStatus = Connected
		m.connectionError = ""
		m.tablesLoading = true
		// After successful connection, fetch tables
		return m, fetchTables(m.driver)

	case connectionFailedMsg:
		m.connectionStatus = ConnectionFailed
		m.connectionError = msg.err.Error()
		return m, nil

	case tablesLoadedMsg:
		m.tablesLoading = false
		m.tables = msg.tables
		m.allEntities = msg.allEntities
		m.tablesError = ""
		m.selectedRow = 0
		m.scrollOffset = 0
		return m, nil

	case tablesLoadFailedMsg:
		m.tablesLoading = false
		m.tablesError = msg.err.Error()
		return m, nil

	case tableDataLoadedMsg:
		m.tableDataLoading = false
		m.tableColumns = msg.columns
		m.tableData = msg.rows
		m.allTableData = msg.rows // Store unfiltered data
		m.tableDataError = ""
		m.tableContentFilter = "" // Clear filter on new data load
		// Reset scroll and selection when new data is loaded
		m.tableDataScrollX = 0
		m.tableDataScrollY = 0
		m.selectedDataRow = 0
		m.selectedDataCol = 0
		// Note: tableDataOffset is managed by pagination handlers, not reset here
		return m, nil

	case tableDataLoadFailedMsg:
		m.tableDataLoading = false
		m.tableDataError = msg.err.Error()
		return m, nil

	case sqlQueryResultMsg:
		// SQL query executed successfully
		m.tableViewMode = true
		m.isCustomQuery = true
		m.executedSQLQuery = msg.query
		m.currentViewTable = nil
		m.tableColumns = msg.columns
		m.tableData = msg.rows
		m.allTableData = msg.rows // Store unfiltered data
		m.tableContentFilter = "" // Clear filter
		m.tableDataLoading = false
		m.tableDataError = ""
		m.tableDataScrollX = 0
		m.tableDataScrollY = 0
		m.selectedDataRow = 0
		m.selectedDataCol = 0
		m.tableDataOffset = 0
		return m, nil

	case sqlQueryFailedMsg:
		// SQL query failed - show query and complete error in table view
		m.tableDataLoading = false
		m.tableDataError = fmt.Sprintf("Query: %s\n\nError: %v", msg.query, msg.err)
		return m, nil

	case cellUpdateSuccessMsg:
		// Cell update successful - close edit popup and stay on current page
		m.cellEditMode = false
		m.cellEditValue = ""
		m.cellEditCursor = 0
		m.cellEditCommandMode = false
		m.cellEditCommand = ""
		// Update the cell value in the current data without refreshing
		if m.cellEditRowIdx >= 0 && m.cellEditRowIdx < len(m.tableData) {
			if m.cellEditColIdx >= 0 && m.cellEditColIdx < len(m.tableData[m.cellEditRowIdx]) {
				m.tableData[m.cellEditRowIdx][m.cellEditColIdx] = msg.newValue
			}
		}
		// Clear any previous errors
		m.tableDataError = ""
		return m, nil

	case cellUpdateFailedMsg:
		// Cell update failed - show error in edit popup or as notification
		// For now, we'll show it in the table error field and close the popup
		m.cellEditMode = false
		m.cellEditValue = ""
		m.cellEditCursor = 0
		m.cellEditCommandMode = false
		m.cellEditCommand = ""
		m.tableDataError = fmt.Sprintf("Update failed: %v", msg.err)
		return m, nil

	case clearClipboardMsgType:
		// Clear the clipboard notification message
		m.clipboardMessage = ""
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// handleKeyPress handles all keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
				// Push current state to history before navigating to new table
				m = m.pushHistory()

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
			// Push current state to history before navigating
			m = m.pushHistory()

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
	switch msg.String() {
	case "enter":
		// Open table view for selected table
		if len(m.tables) > 0 && m.selectedRow < len(m.tables) {
			// Push current state to history before navigating
			m = m.pushHistory()

			selectedTable := m.tables[m.selectedRow]
			m.tableViewMode = true
			m.currentViewTable = &selectedTable
			m.tableDataLoading = true
			m.tableDataError = ""
			m.tableDataOffset = 0 // Start from first page
			return m, fetchTableData(m.driver, selectedTable.SchemaName, selectedTable.TableName, 0)
		}

	case "o":
		// Navigate back in history
		return m.navigateBack()

	case "i":
		// Navigate forward in history
		return m.navigateForward()

	case "/":
		// Activate inline search mode
		m.inlineSearchMode = true
		m.inlineSearchQuery = ""
		m.inlineSearchSuggestion = ""

	case "ctrl+p":
		// Open search popup
		m.searchMode = true
		m.searchQuery = ""
		m.searchSuggestion = ""
		m.filteredTables = m.allEntities
		m.searchSelected = 0

	case "s":
		// Open SQL query mode
		m.sqlQueryMode = true
		m.sqlQueryInput = ""
		m.sqlQueryCursor = 0

	case "q":
		// Quit application
		return m, tea.Quit

	case ":":
		m.commandMode = true
		m.commandBuffer = ":"

	case "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		// Move down in table list
		if m.selectedRow < len(m.tables)-1 {
			m.selectedRow++
		}

	case "k", "up":
		// Move up in table list
		if m.selectedRow > 0 {
			m.selectedRow--
		}

	case "g":
		// Go to top
		m.selectedRow = 0
		m.scrollOffset = 0

	case "G":
		// Go to bottom
		if len(m.tables) > 0 {
			m.selectedRow = len(m.tables) - 1
		}
	}

	return m, nil
}

// handleTableViewModeKeys handles keyboard input in table view mode
func (m Model) handleTableViewModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Clear content filter if active, otherwise exit table view mode
		if m.tableContentFilter != "" {
			m.tableContentFilter = ""
			m.tableData = m.allTableData // Restore unfiltered data
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

	case "q":
		// Quit application
		return m, tea.Quit

	case "s":
		// Open SQL query mode - show current query or allow new query
		if m.isCustomQuery && m.executedSQLQuery != "" {
			// For custom queries, show the executed query
			m.sqlQueryMode = true
			m.sqlQueryInput = m.executedSQLQuery
			m.sqlQueryCursor = len([]rune(m.sqlQueryInput))
		} else if m.currentViewTable != nil {
			// For regular tables, generate the SELECT query
			query := fmt.Sprintf("SELECT * FROM \"%s\".\"%s\" LIMIT 500 OFFSET %d",
				m.currentViewTable.SchemaName,
				m.currentViewTable.TableName,
				m.tableDataOffset)
			m.sqlQueryMode = true
			m.sqlQueryInput = query
			m.sqlQueryCursor = len([]rune(query))
		} else {
			// No current query context, start with empty query
			m.sqlQueryMode = true
			m.sqlQueryInput = ""
			m.sqlQueryCursor = 0
		}

	case "/":
		// Open content filter input
		// Switch to a special mode for typing the filter
		m.inlineSearchMode = true
		m.inlineSearchQuery = m.tableContentFilter

	case "ctrl+c":
		return m, tea.Quit

	case "o":
		// Navigate back in history
		return m.navigateBack()

	case "ctrl+p":
		// Open search popup to switch to another table
		m.searchMode = true
		m.searchQuery = ""
		m.searchSuggestion = ""
		m.filteredTables = m.allEntities
		m.searchSelected = 0

	case "n":
		// Next page - load next 500 rows
		if m.currentViewTable != nil && len(m.tableData) == 500 {
			// Only paginate if we have a full page (meaning there might be more data)
			m.tableDataOffset += 500
			m.tableDataLoading = true
			return m, fetchTableData(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, m.tableDataOffset)
		}

	case "p":
		// Previous page - load previous 500 rows
		if m.currentViewTable != nil && m.tableDataOffset > 0 {
			// Only go back if we're not at the first page
			m.tableDataOffset -= 500
			if m.tableDataOffset < 0 {
				m.tableDataOffset = 0
			}
			m.tableDataLoading = true
			return m, fetchTableData(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, m.tableDataOffset)
		}

	case "h", "left":
		// Scroll left
		if m.tableDataScrollX > 0 {
			m.tableDataScrollX -= 10
			if m.tableDataScrollX < 0 {
				m.tableDataScrollX = 0
			}
		}

	case "l", "right":
		// Scroll right
		m.tableDataScrollX += 10

	case "w":
		// Move to next column (cell) in current row
		if len(m.tableColumns) > 0 && m.selectedDataCol < len(m.tableColumns)-1 {
			m.selectedDataCol++
			// Auto-scroll horizontally to keep selected cell visible
			m = m.adjustTableDataHorizontalScroll()
		}

	case "b":
		// Move to previous column (cell) in current row
		if m.selectedDataCol > 0 {
			m.selectedDataCol--
			// Auto-scroll horizontally to keep selected cell visible
			m = m.adjustTableDataHorizontalScroll()
		}

	case "j", "down":
		// Move cursor down
		if m.selectedDataRow < len(m.tableData)-1 {
			m.selectedDataRow++
			// Adjust scroll to keep cursor visible
			m = m.adjustTableDataScroll()
		}

	case "k", "up":
		// Move cursor up
		if m.selectedDataRow > 0 {
			m.selectedDataRow--
			// Adjust scroll to keep cursor visible
			m = m.adjustTableDataScroll()
		}

	case "g":
		// Go to top
		m.selectedDataRow = 0
		m.tableDataScrollY = 0

	case "G":
		// Go to bottom
		if len(m.tableData) > 0 {
			m.selectedDataRow = len(m.tableData) - 1
			// Adjust scroll to keep cursor visible
			m = m.adjustTableDataScroll()
		}

	case "v":
		// Show cell value in popup
		if len(m.tableData) > 0 && m.selectedDataRow < len(m.tableData) {
			row := m.tableData[m.selectedDataRow]
			if m.selectedDataCol < len(row) {
				m = m.openCellValuePopup(row[m.selectedDataCol])
			}
		}

	case "i":
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

	case "V":
		// Show entire record in popup (all key-value pairs)
		if len(m.tableData) > 0 && m.selectedDataRow < len(m.tableData) {
			m = m.openRecordViewPopup()
		}

	case "y":
		// Copy current cell value to clipboard
		if len(m.tableData) > 0 && m.selectedDataRow < len(m.tableData) {
			row := m.tableData[m.selectedDataRow]
			if m.selectedDataCol < len(row) {
				m = m.copyCellToClipboard(row[m.selectedDataCol])
				return m, clearClipboardMessage()
			}
		}

	case "Y":
		// Copy entire row to clipboard in CSV format
		if len(m.tableData) > 0 && m.selectedDataRow < len(m.tableData) {
			m = m.copyRowToClipboard()
			return m, clearClipboardMessage()
		}
	}

	return m, nil
}

// adjustTableDataScroll adjusts the scroll offset to keep the selected row visible
func (m Model) adjustTableDataScroll() Model {
	// Calculate visible rows based on available height
	availableHeight := m.height - 14 // Same calculation as in renderTableDataView
	if availableHeight < 5 {
		availableHeight = 5
	}
	visibleRows := availableHeight - 2
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Adjust scroll offset to keep selected row visible
	if m.selectedDataRow < m.tableDataScrollY {
		m.tableDataScrollY = m.selectedDataRow
	} else if m.selectedDataRow >= m.tableDataScrollY+visibleRows {
		m.tableDataScrollY = m.selectedDataRow - visibleRows + 1
	}

	return m
}

// adjustTableDataHorizontalScroll adjusts horizontal scroll to keep selected cell visible
func (m Model) adjustTableDataHorizontalScroll() Model {
	if len(m.tableColumns) == 0 {
		return m
	}

	// Calculate column widths (same logic as in renderTableDataView)
	columnWidths := make([]int, len(m.tableColumns))
	for i, col := range m.tableColumns {
		columnWidths[i] = len(col)
		// Check data rows for max width
		for _, row := range m.tableData {
			if i < len(row) && len(row[i]) > columnWidths[i] {
				columnWidths[i] = len(row[i])
			}
		}
		// Cap column width at 30 characters for display
		if columnWidths[i] > 30 {
			columnWidths[i] = 30
		}
		// Minimum width of 10
		if columnWidths[i] < 10 {
			columnWidths[i] = 10
		}
	}

	// Calculate the position of the selected column
	selectedColStart := 0
	for i := 0; i < m.selectedDataCol; i++ {
		selectedColStart += columnWidths[i] + 3 // +3 for separator
	}
	selectedColWidth := columnWidths[m.selectedDataCol] + 3 // Include separator
	selectedColEnd := selectedColStart + selectedColWidth

	availableWidth := m.width - 12
	if availableWidth < 40 {
		availableWidth = 40
	}

	// Adjust scroll to show selected column
	if selectedColStart < m.tableDataScrollX {
		// Selected column is before visible area, scroll left
		// Position so the selected column starts at the left edge
		m.tableDataScrollX = selectedColStart
	} else if selectedColEnd > m.tableDataScrollX+availableWidth {
		// Selected column is after visible area, scroll right
		// Position so the selected column starts at the left edge of visible area
		m.tableDataScrollX = selectedColStart
	}

	return m
}

// captureCurrentState captures the current view state
func (m Model) captureCurrentState() HistoryState {
	return HistoryState{
		tableViewMode:      m.tableViewMode,
		selectedRow:        m.selectedRow,
		scrollOffset:       m.scrollOffset,
		currentViewTable:   m.currentViewTable,
		tableDataOffset:    m.tableDataOffset,
		selectedDataRow:    m.selectedDataRow,
		selectedDataCol:    m.selectedDataCol,
		tableDataScrollX:   m.tableDataScrollX,
		tableDataScrollY:   m.tableDataScrollY,
	}
}

// pushHistory saves current state and pushes to history stack
func (m Model) pushHistory() Model {
	currentState := m.captureCurrentState()

	// If we're in the middle of history (after going back), truncate forward history
	if m.historyIndex < len(m.historyStack)-1 {
		m.historyStack = m.historyStack[:m.historyIndex+1]
	}

	// Add current state to history
	m.historyStack = append(m.historyStack, currentState)
	m.historyIndex = len(m.historyStack) - 1

	return m
}

// restoreState restores a previous state (without loading data)
func (m Model) restoreState(state HistoryState) Model {
	m.tableViewMode = state.tableViewMode
	m.selectedRow = state.selectedRow
	m.scrollOffset = state.scrollOffset
	m.currentViewTable = state.currentViewTable
	m.tableDataOffset = state.tableDataOffset
	m.selectedDataRow = state.selectedDataRow
	m.selectedDataCol = state.selectedDataCol
	m.tableDataScrollX = state.tableDataScrollX
	m.tableDataScrollY = state.tableDataScrollY

	return m
}

// navigateBack goes back in history (Ctrl+O)
func (m Model) navigateBack() (Model, tea.Cmd) {
	if m.historyIndex <= 0 {
		// No history to go back to
		return m, nil
	}

	// Move back in history
	m.historyIndex--
	state := m.historyStack[m.historyIndex]

	// Restore state
	m = m.restoreState(state)

	// If navigating to table view, reload data
	if m.tableViewMode && m.currentViewTable != nil {
		m.tableDataLoading = true
		return m, fetchTableData(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, m.tableDataOffset)
	}

	return m, nil
}

// navigateForward goes forward in history (Ctrl+I)
func (m Model) navigateForward() (Model, tea.Cmd) {
	if m.historyIndex >= len(m.historyStack)-1 {
		// No forward history
		return m, nil
	}

	// Move forward in history
	m.historyIndex++
	state := m.historyStack[m.historyIndex]

	// Restore state
	m = m.restoreState(state)

	// If navigating to table view, reload data
	if m.tableViewMode && m.currentViewTable != nil {
		m.tableDataLoading = true
		return m, fetchTableData(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, m.tableDataOffset)
	}

	return m, nil
}

// openCellValuePopup opens the popup to display a cell value
func (m Model) openCellValuePopup(value string) Model {
	m.cellValuePopupMode = true
	m.cellValuePopupContent = value
	m.cellValuePopupScroll = 0
	m.cellValuePopupSelected = 0

	// Try to parse as JSON
	m.cellValuePopupIsJSON = false
	m.cellValuePopupTree = nil

	if value != "" && (value[0] == '{' || value[0] == '[') {
		// Attempt to parse as JSON
		tree, isJSON := buildJSONTree(value)
		if isJSON {
			m.cellValuePopupIsJSON = true
			m.cellValuePopupTree = tree
		}
	}

	return m
}

// handleCellValuePopupKeys handles keyboard input in cell value popup mode
func (m Model) handleCellValuePopupKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "v":
		// Close popup
		m.cellValuePopupMode = false
		m.cellValuePopupContent = ""
		m.cellValuePopupIsJSON = false
		m.cellValuePopupTree = nil
		m.cellValuePopupScroll = 0
		m.cellValuePopupSelected = 0

	case "j", "down":
		// Scroll down or navigate to next node
		if m.cellValuePopupIsJSON && len(m.cellValuePopupTree) > 0 {
			if m.cellValuePopupSelected < len(m.cellValuePopupTree)-1 {
				m.cellValuePopupSelected++
			}
		} else {
			m.cellValuePopupScroll++
		}

	case "k", "up":
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

	case "h", "left":
		// Collapse node in JSON tree
		if m.cellValuePopupIsJSON && len(m.cellValuePopupTree) > 0 {
			m = m.collapseJSONNode()
		}

	case "l", "right", "enter":
		// Expand node in JSON tree
		if m.cellValuePopupIsJSON && len(m.cellValuePopupTree) > 0 {
			m = m.expandJSONNode()
		}

	case "g":
		// Go to top
		m.cellValuePopupSelected = 0
		m.cellValuePopupScroll = 0

	case "G":
		// Go to bottom
		if m.cellValuePopupIsJSON && len(m.cellValuePopupTree) > 0 {
			m.cellValuePopupSelected = len(m.cellValuePopupTree) - 1
		}

	case "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

// expandJSONNode expands the selected JSON node
func (m Model) expandJSONNode() Model {
	if m.cellValuePopupSelected >= len(m.cellValuePopupTree) {
		return m
	}

	node := &m.cellValuePopupTree[m.cellValuePopupSelected]
	if !node.HasChildren {
		return m
	}

	if node.Expanded {
		// Already expanded, do nothing
		return m
	}

	// Mark as expanded
	node.Expanded = true

	// Rebuild the visible tree
	m.cellValuePopupTree = rebuildVisibleJSONTree(m.cellValuePopupTree)

	return m
}

// collapseJSONNode collapses the selected JSON node
func (m Model) collapseJSONNode() Model {
	if m.cellValuePopupSelected >= len(m.cellValuePopupTree) {
		return m
	}

	node := &m.cellValuePopupTree[m.cellValuePopupSelected]
	if !node.HasChildren {
		return m
	}

	if !node.Expanded {
		// Already collapsed, do nothing
		return m
	}

	// Mark as collapsed
	node.Expanded = false

	// Rebuild the visible tree
	m.cellValuePopupTree = rebuildVisibleJSONTree(m.cellValuePopupTree)

	return m
}

// openRecordViewPopup opens the popup to display an entire record as key-value pairs
func (m Model) openRecordViewPopup() Model {
	m.recordViewMode = true
	m.recordViewData = m.tableData[m.selectedDataRow]
	m.recordViewColumns = m.tableColumns
	m.recordViewSelected = 0
	m.recordViewScroll = 0

	return m
}

// handleRecordViewKeys handles keyboard input in record view popup mode
func (m Model) handleRecordViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "V":
		// Close record view popup
		m.recordViewMode = false
		m.recordViewData = nil
		m.recordViewColumns = nil
		m.recordViewSelected = 0
		m.recordViewScroll = 0

	case "v":
		// Open cell value popup for the selected field
		if m.recordViewSelected < len(m.recordViewData) {
			m = m.openCellValuePopup(m.recordViewData[m.recordViewSelected])
		}

	case "j", "down":
		// Navigate down in field list
		if m.recordViewSelected < len(m.recordViewData)-1 {
			m.recordViewSelected++
		}

	case "k", "up":
		// Navigate up in field list
		if m.recordViewSelected > 0 {
			m.recordViewSelected--
		}

	case "g":
		// Go to top
		m.recordViewSelected = 0
		m.recordViewScroll = 0

	case "G":
		// Go to bottom
		if len(m.recordViewData) > 0 {
			m.recordViewSelected = len(m.recordViewData) - 1
		}

	case "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

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

// handleSQLQueryModeKeys handles keyboard input in SQL query mode
func (m Model) handleSQLQueryModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Convert string to runes for proper UTF-8 handling
	runes := []rune(m.sqlQueryInput)

	switch msg.String() {
	case "enter":
		// Execute the SQL query
		if m.sqlQueryInput != "" {
			m.sqlQueryMode = false
			m.tableDataLoading = true
			return m, executeSQLQuery(m.driver, m.sqlQueryInput)
		}

	case "esc":
		// Cancel SQL query mode
		m.sqlQueryMode = false
		m.sqlQueryInput = ""
		m.sqlQueryCursor = 0

	case "left":
		// Move cursor left
		if m.sqlQueryCursor > 0 {
			m.sqlQueryCursor--
		}

	case "right":
		// Move cursor right
		if m.sqlQueryCursor < len(runes) {
			m.sqlQueryCursor++
		}

	case "home", "ctrl+a":
		// Move cursor to start
		m.sqlQueryCursor = 0

	case "end", "ctrl+e":
		// Move cursor to end
		m.sqlQueryCursor = len(runes)

	case "backspace":
		// Delete character before cursor
		if m.sqlQueryCursor > 0 && m.sqlQueryCursor <= len(runes) {
			runes = append(runes[:m.sqlQueryCursor-1], runes[m.sqlQueryCursor:]...)
			m.sqlQueryInput = string(runes)
			m.sqlQueryCursor--
		}

	case "delete":
		// Delete character at cursor
		if m.sqlQueryCursor >= 0 && m.sqlQueryCursor < len(runes) {
			runes = append(runes[:m.sqlQueryCursor], runes[m.sqlQueryCursor+1:]...)
			m.sqlQueryInput = string(runes)
		}

	case "ctrl+c":
		return m, tea.Quit

	default:
		// Insert at cursor position
		// Allow multi-character keys like space, tab, etc.
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
			// Insert at cursor position
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
	// If in command mode within edit mode
	if m.cellEditCommandMode {
		switch msg.String() {
		case "enter":
			// Execute command
			if m.cellEditCommand == ":w" || m.cellEditCommand == ":w " {
				// Save the cell value and exit edit mode
				return m, updateCellValue(m.driver, m.currentViewTable, m.tableColumns, m.tableData, m.cellEditRowIdx, m.cellEditColIdx, m.cellEditValue, m.tableDataOffset)
			} else if m.cellEditCommand == ":q" || m.cellEditCommand == ":q " {
				// Quit without saving
				m.cellEditMode = false
				m.cellEditCommandMode = false
				m.cellEditCommand = ""
			}
			m.cellEditCommandMode = false
			m.cellEditCommand = ""

		case "esc":
			// Cancel command mode
			m.cellEditCommandMode = false
			m.cellEditCommand = ""

		case "backspace":
			// Delete last character in command
			if len(m.cellEditCommand) > 0 {
				m.cellEditCommand = m.cellEditCommand[:len(m.cellEditCommand)-1]
			}

		case "ctrl+c":
			return m, tea.Quit

		default:
			// Append to command
			keyStr := msg.String()
			if len(keyStr) == 1 || keyStr == "space" {
				if keyStr == "space" {
					m.cellEditCommand += " "
				} else {
					m.cellEditCommand += keyStr
				}
			}
		}
		return m, nil
	}

	// Normal edit mode (not command mode)
	// Convert string to runes for proper UTF-8 handling
	runes := []rune(m.cellEditValue)

	switch msg.String() {
	case "esc":
		// Cancel editing
		m.cellEditMode = false
		m.cellEditValue = ""
		m.cellEditCursor = 0
		m.cellEditCommandMode = false
		m.cellEditCommand = ""

	case ":":
		// Enter command mode
		m.cellEditCommandMode = true
		m.cellEditCommand = ":"

	case "left":
		// Move cursor left
		if m.cellEditCursor > 0 {
			m.cellEditCursor--
		}

	case "right":
		// Move cursor right
		if m.cellEditCursor < len(runes) {
			m.cellEditCursor++
		}

	case "home", "ctrl+a":
		// Move cursor to start
		m.cellEditCursor = 0

	case "end", "ctrl+e":
		// Move cursor to end
		m.cellEditCursor = len(runes)

	case "backspace":
		// Delete character before cursor
		if m.cellEditCursor > 0 && m.cellEditCursor <= len(runes) {
			runes = append(runes[:m.cellEditCursor-1], runes[m.cellEditCursor:]...)
			m.cellEditValue = string(runes)
			m.cellEditCursor--
		}

	case "delete":
		// Delete character at cursor
		if m.cellEditCursor >= 0 && m.cellEditCursor < len(runes) {
			runes = append(runes[:m.cellEditCursor], runes[m.cellEditCursor+1:]...)
			m.cellEditValue = string(runes)
		}

	case "enter":
		// Insert newline
		runes = append(runes[:m.cellEditCursor], append([]rune{'\n'}, runes[m.cellEditCursor:]...)...)
		m.cellEditValue = string(runes)
		m.cellEditCursor++

	case "ctrl+c":
		return m, tea.Quit

	default:
		// Insert at cursor position
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
			// Insert at cursor position
			if m.cellEditCursor >= 0 && m.cellEditCursor <= len(runes) {
				runes = append(runes[:m.cellEditCursor], append([]rune{charToInsert}, runes[m.cellEditCursor:]...)...)
				m.cellEditValue = string(runes)
				m.cellEditCursor++
			}
		}
	}

	return m, nil
}

// clearClipboardMessage returns a command that clears the clipboard message after a delay
func clearClipboardMessage() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return clearClipboardMsgType{}
	})
}
