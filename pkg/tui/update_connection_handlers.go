package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// handleConnectionSuccess handles successful database connection
func (m Model) handleConnectionSuccess(msg connectionSuccessMsg) (tea.Model, tea.Cmd) {
	m.connectionStatus = Connected
	m.connectionError = ""
	m.tablesLoading = true
	// After successful connection, save to history and fetch tables
	if m.connHistory != nil {
		err := m.connHistory.Add(
			m.dbConfig.Driver,
			m.dbConfig.Host,
			m.dbConfig.Port,
			m.dbConfig.Username,
			m.dbConfig.Password,
			m.dbConfig.Database,
			m.dbConfig.Schema,
			m.dbConfig.SSLMode,
		)
		if err != nil {
			// Log error but don't fail connection
			m = m.debugLog(fmt.Sprintf("Failed to save connection to history: %v", err))
		} else {
			m = m.debugLog("Connection saved to history successfully")
		}
	} else {
		m = m.debugLog("Connection history is nil, cannot save connection")
	}
	return m, fetchTables(m.driver)
}

// handleConnectionFailed handles failed database connection
func (m Model) handleConnectionFailed(msg connectionFailedMsg) (tea.Model, tea.Cmd) {
	m.connectionStatus = ConnectionFailed
	m.connectionError = msg.err.Error()
	return m, nil
}

// handleConnectionSwitchSuccess handles successful connection switch
func (m Model) handleConnectionSwitchSuccess(msg connectionSwitchSuccessMsg) (tea.Model, tea.Cmd) {
	// Close old driver
	if m.driver != nil {
		_ = m.driver.Close()
	}

	// Update to new connection
	m.driver = msg.driver
	m.dbConfig = msg.dbConfig
	m.sqlHistory = msg.sqlHistory
	m.connectionStatus = Connected
	m.connectionError = ""

	// Reset view state
	m.tableViewMode = false
	m.currentViewTable = nil
	m.tableData = nil
	m.tableColumns = nil
	m.selectedRow = 0
	m.scrollOffset = 0
	m.historyStack = make([]HistoryState, 0)
	m.historyIndex = -1

	// Save to connection history
	if m.connHistory != nil {
		err := m.connHistory.Add(
			msg.dbConfig.Driver,
			msg.dbConfig.Host,
			msg.dbConfig.Port,
			msg.dbConfig.Username,
			msg.dbConfig.Password,
			msg.dbConfig.Database,
			msg.dbConfig.Schema,
			msg.dbConfig.SSLMode,
		)
		if err != nil {
			m = m.debugLog(fmt.Sprintf("Failed to save switched connection to history: %v", err))
		}
	}

	// Fetch tables for new connection
	m.tablesLoading = true
	return m, fetchTables(m.driver)
}

// handleConnectionSwitchFailed handles failed connection switch
func (m Model) handleConnectionSwitchFailed(msg connectionSwitchFailedMsg) (tea.Model, tea.Cmd) {
	m.connectionError = fmt.Sprintf("Connection failed: %v", msg.err)
	m.connectionStatus = Disconnected
	// If we have connection input values, reopen the form to show error and allow retry
	hasInputValues := false
	for _, input := range m.connInputs {
		if input.Value() != "" {
			hasInputValues = true
			break
		}
	}
	if hasInputValues {
		m.connInputMode = true
		return m, nil
	}
	// Otherwise reopen connection manager to show error and allow retry
	m.connManagerMode = true
	if m.connHistory != nil {
		m.filteredConnections = m.connHistory.GetAll()
	}
	return m, nil
}
