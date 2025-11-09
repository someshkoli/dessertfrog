package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// captureCurrentState captures the current view state
func (m Model) captureCurrentState() HistoryState {
	return HistoryState{
		tableViewMode:    m.tableViewMode,
		selectedRow:      m.selectedRow,
		scrollOffset:     m.scrollOffset,
		currentViewTable: m.currentViewTable,
		tableDataOffset:  m.tableDataOffset,
		selectedDataRow:  m.selectedDataRow,
		selectedDataCol:  m.selectedDataCol,
		tableDataScrollX: m.tableDataScrollX,
		tableDataScrollY: m.tableDataScrollY,
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

	m = m.debugLog("History pushed: index=%d, stack_size=%d", m.historyIndex, len(m.historyStack))
	for _, history := range m.historyStack {
		if history.tableViewMode {
			m = m.debugLog("History elements: %s", history.currentViewTable.TableName)
		} else {
			m = m.debugLog("History elements: home")
		}
	}

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
	m = m.debugLog("Navigate back requested: current_index=%d, stack_size=%d", m.historyIndex, len(m.historyStack))
	m = m.debugLog("  Before: isNavigatingHistory=%v", m.isNavigatingHistory)

	if m.historyIndex <= 0 {
		// No history to go back to
		m = m.debugLog("Navigate back: no history available (index=%d)", m.historyIndex)
		return m, nil
	}

	// Set flag to prevent pushing to history during navigation
	m.isNavigatingHistory = true

	// Move back in history
	m.historyIndex--
	state := m.historyStack[m.historyIndex]

	m = m.debugLog("Navigate back: index=%d->%d, stack_size=%d", m.historyIndex+1, m.historyIndex, len(m.historyStack))
	m = m.debugLog("  After index change: isNavigatingHistory=%v", m.isNavigatingHistory)

	// Log current history stack
	for i, history := range m.historyStack {
		marker := " "
		if i == m.historyIndex {
			marker = ">"
		}
		if history.tableViewMode {
			if history.currentViewTable != nil {
				m = m.debugLog("  %s[%d] %s", marker, i, history.currentViewTable.TableName)
			} else {
				m = m.debugLog("  %s[%d] <nil table>", marker, i)
			}
		} else {
			m = m.debugLog("  %s[%d] home", marker, i)
		}
	}

	// Restore state
	m = m.restoreState(state)

	// If navigating to table view, reload data
	if m.tableViewMode && m.currentViewTable != nil {
		m.tableDataLoading = true
		m = m.debugLog("  Triggering data load for %s", m.currentViewTable.TableName)
		return m, fetchTableData(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, m.tableDataOffset)
	}

	// Clear the flag if we didn't trigger data load
	m.isNavigatingHistory = false
	m = m.debugLog("  No data load needed, cleared flag")

	return m, nil
}

// navigateForward goes forward in history (Ctrl+I)
func (m Model) navigateForward() (Model, tea.Cmd) {
	m = m.debugLog("Navigate forward requested: current_index=%d, stack_size=%d", m.historyIndex, len(m.historyStack))
	m = m.debugLog("  Before: isNavigatingHistory=%v", m.isNavigatingHistory)

	if m.historyIndex >= len(m.historyStack)-1 {
		// No forward history
		m = m.debugLog("Navigate forward: no forward history available (index=%d, stack_size=%d)", m.historyIndex, len(m.historyStack))
		return m, nil
	}

	// Set flag to prevent pushing to history during navigation
	m.isNavigatingHistory = true

	// Move forward in history
	m.historyIndex++
	state := m.historyStack[m.historyIndex]

	m = m.debugLog("Navigate forward: index=%d->%d, stack_size=%d", m.historyIndex-1, m.historyIndex, len(m.historyStack))
	m = m.debugLog("  After index change: isNavigatingHistory=%v", m.isNavigatingHistory)

	// Log current history stack
	for i, history := range m.historyStack {
		marker := " "
		if i == m.historyIndex {
			marker = ">"
		}
		if history.tableViewMode {
			if history.currentViewTable != nil {
				m = m.debugLog("  %s[%d] %s", marker, i, history.currentViewTable.TableName)
			} else {
				m = m.debugLog("  %s[%d] <nil table>", marker, i)
			}
		} else {
			m = m.debugLog("  %s[%d] home", marker, i)
		}
	}

	// Restore state
	m = m.restoreState(state)

	// If navigating to table view, reload data
	if m.tableViewMode && m.currentViewTable != nil {
		m.tableDataLoading = true
		m = m.debugLog("  Triggering data load for %s", m.currentViewTable.TableName)
		return m, fetchTableData(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, m.tableDataOffset)
	}

	// Clear the flag if we didn't trigger data load
	m.isNavigatingHistory = false
	m = m.debugLog("  No data load needed, cleared flag")

	return m, nil
}
