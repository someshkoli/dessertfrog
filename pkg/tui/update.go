package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

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

		// Update schema panel line count for first table
		m = m.updateSchemaPanelLineCount()

		// Start loading detailed schema for first table
		if len(m.tables) > 0 {
			m.schemaInfoLoading = true
			m.schemaInfo = nil
			return m, fetchSchemaInfo(m.driver, m.tables[0].SchemaName, m.tables[0].TableName)
		}

		// Push initial home state to history
		m = m.pushHistory()

		return m, nil

	case tablesLoadFailedMsg:
		m.tablesLoading = false
		m.tablesError = msg.err.Error()
		return m, nil

	case tableDataLoadedMsg:
		m = m.debugLog("tableDataLoadedMsg received: rows=%d, cols=%d", len(msg.rows), len(msg.columns))
		m = m.debugLog("  isNavigatingHistory=%v", m.isNavigatingHistory)
		m = m.debugLog("  Before: historyIndex=%d, stack_size=%d", m.historyIndex, len(m.historyStack))

		m.tableDataLoading = false
		m.tableColumns = msg.columns
		m.tableData = msg.rows
		m.allTableData = msg.rows // Store unfiltered data
		m.tableDataError = ""
		m.tableContentFilter = "" // Clear filter on new data load
		m.queryTime = msg.queryTime
		m.fetchTime = msg.fetchTime
		// Reset scroll and selection when new data is loaded
		m.tableDataScrollX = 0
		m.tableDataScrollY = 0
		m.selectedDataRow = 0
		m.selectedDataCol = 0
		// Note: tableDataOffset is managed by pagination handlers, not reset here

		// Push to history only if this is not a history navigation
		if !m.isNavigatingHistory {
			m = m.debugLog("  Not navigating history, pushing new state")
			m = m.pushHistory()
		} else {
			// Clear the flag after handling history navigation
			m = m.debugLog("  Navigating history, NOT pushing, clearing flag")
			m.isNavigatingHistory = false
		}

		m = m.debugLog("  After: historyIndex=%d, stack_size=%d", m.historyIndex, len(m.historyStack))

		return m, nil

	case tableDataLoadFailedMsg:
		m.tableDataLoading = false
		m.tableDataError = msg.err.Error()
		return m, nil

	case tableSchemaLoadedMsg:
		// Update currentViewTable with full schema (includes primary key info)
		if m.currentViewTable != nil && msg.schema != nil {
			// Only update if it's the same table
			if m.currentViewTable.SchemaName == msg.schema.SchemaName &&
				m.currentViewTable.TableName == msg.schema.TableName {
				m.currentViewTable = msg.schema
				m = m.debugLog("Table schema loaded: %d columns, PK columns: %v",
					len(msg.schema.Columns), getPrimaryKeyNames(msg.schema))
			}
		}
		return m, nil

	case tableSchemaLoadFailedMsg:
		// Schema load failed - not critical, log and continue
		m = m.debugLog("Failed to load table schema: %v", msg.err)
		return m, nil

	case schemaInfoLoadedMsg:
		// Schema info loaded successfully for schema panel
		m.schemaInfoLoading = false
		m.schemaInfo = msg.schema
		// Update line count with new detailed schema
		m = m.updateSchemaPanelLineCount()
		return m, nil

	case schemaInfoLoadFailedMsg:
		// Schema info load failed - not critical
		m.schemaInfoLoading = false
		m = m.debugLog("Failed to load schema info: %v", msg.err)
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
		m.queryTime = msg.queryTime
		m.fetchTime = msg.fetchTime

		// Push to history
		if !m.isNavigatingHistory {
			m = m.debugLog("  SQL query result, pushing to history")
			m = m.pushHistory()
		} else {
			m = m.debugLog("  Navigating history, NOT pushing, clearing flag")
			m.isNavigatingHistory = false
		}

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
