package tui

import (
	"fmt"
	"strings"
	"time"
)

// debugLog adds a debug message to the log buffer
func (m Model) debugLog(format string, args ...interface{}) Model {
	timestamp := time.Now().Format("15:04:05.000")
	message := fmt.Sprintf("[%s] %s", timestamp, fmt.Sprintf(format, args...))

	// Add to logs
	m.debugLogs = append(m.debugLogs, message)

	// Keep only last N logs (ring buffer)
	if len(m.debugLogs) > m.debugMaxLogs {
		m.debugLogs = m.debugLogs[len(m.debugLogs)-m.debugMaxLogs:]
		// Adjust scroll and selection if they're out of bounds
		if m.debugLogScrollOffset > 0 {
			m.debugLogScrollOffset--
		}
		if m.debugSelectedLog > 0 {
			m.debugSelectedLog--
		}
	}

	// If not focused, auto-scroll to show new log
	if !m.debugPanelFocused {
		// Auto-scroll to bottom to show latest log
		m.debugLogScrollOffset = len(m.debugLogs) - 1
		if m.debugLogScrollOffset < 0 {
			m.debugLogScrollOffset = 0
		}
	}

	return m
}

// getDebugState returns a formatted string with current app state
func (m Model) getDebugState() string {
	var state string

	// Connection info
	state += fmt.Sprintf("Connection: %v\n", m.connectionStatus)
	state += fmt.Sprintf("Driver: %s\n", m.dbConfig.Driver)
	state += fmt.Sprintf("Host: %s:%d\n", m.dbConfig.Host, m.dbConfig.Port)
	state += fmt.Sprintf("Database: %s\n", m.dbConfig.Database)
	state += fmt.Sprintf("Schema: %s\n\n", m.dbConfig.Schema)

	// View state
	state += fmt.Sprintf("Table View Mode: %v\n", m.tableViewMode)
	state += fmt.Sprintf("Search Mode: %v\n", m.searchMode)
	state += fmt.Sprintf("Inline Search Mode: %v\n", m.inlineSearchMode)
	state += fmt.Sprintf("Command Mode: %v\n", m.commandMode)
	state += fmt.Sprintf("Cell Edit Mode: %v\n", m.cellEditMode)
	state += fmt.Sprintf("SQL Query Mode: %v\n", m.sqlQueryMode)
	state += fmt.Sprintf("Cell Popup Mode: %v\n", m.cellValuePopupMode)
	state += fmt.Sprintf("Record View Mode: %v\n\n", m.recordViewMode)

	// Table list state
	state += fmt.Sprintf("Tables Count: %d\n", len(m.tables))
	state += fmt.Sprintf("All Entities Count: %d\n", len(m.allEntities))
	state += fmt.Sprintf("Selected Row: %d\n", m.selectedRow)
	state += fmt.Sprintf("Scroll Offset: %d\n\n", m.scrollOffset)

	// Table view state
	if m.tableViewMode {
		if m.currentViewTable != nil {
			state += fmt.Sprintf("Current Table: %s.%s\n", m.currentViewTable.SchemaName, m.currentViewTable.TableName)
			state += fmt.Sprintf("Entity Type: %s\n", m.currentViewTable.EntityType)
		}
		state += fmt.Sprintf("Columns: %d\n", len(m.tableColumns))
		state += fmt.Sprintf("Rows: %d\n", len(m.tableData))
		state += fmt.Sprintf("Selected Cell: [%d, %d]\n", m.selectedDataRow, m.selectedDataCol)
		state += fmt.Sprintf("Scroll: X=%d, Y=%d\n", m.tableDataScrollX, m.tableDataScrollY)
		state += fmt.Sprintf("Data Offset: %d\n", m.tableDataOffset)
		state += fmt.Sprintf("Content Filter: %q\n", m.tableContentFilter)
		state += fmt.Sprintf("Is Custom Query: %v\n\n", m.isCustomQuery)
	}

	// History state
	state += fmt.Sprintf("History Index: %d\n", m.historyIndex)
	state += fmt.Sprintf("History Stack Size: %d\n\n", len(m.historyStack))

	// Window size
	state += fmt.Sprintf("Window: %dx%d\n", m.width, m.height)

	return state
}

// toggleDebugMode toggles the debug overlay
func (m Model) toggleDebugMode() Model {
	m.debugMode = !m.debugMode
	if m.debugMode {
		m = m.debugLog("Debug mode enabled")
		m.debugPanelFocused = false
		m.debugSelectedSection = 1
		// Start with last log selected
		if len(m.debugLogs) > 0 {
			m.debugSelectedLog = len(m.debugLogs) - 1
		} else {
			m.debugSelectedLog = 0
		}
		// Auto-scroll to show latest logs
		m.debugLogScrollOffset = len(m.debugLogs) - 1
		if m.debugLogScrollOffset < 0 {
			m.debugLogScrollOffset = 0
		}
		// Cache state lines for navigation
		m.debugStateLines = strings.Split(m.getDebugState(), "\n")
	} else {
		// Clear focus when closing
		m.debugPanelFocused = false
		m.debugDetailMode = false
	}
	return m
}

// clearDebugLogs clears all debug logs
func (m Model) clearDebugLogs() Model {
	m.debugLogs = make([]string, 0)
	m.debugSelectedLog = 0
	m.debugLogScrollOffset = 0
	m = m.debugLog("Debug logs cleared")
	return m
}

// toggleDebugFocus toggles keyboard focus on debug panel
func (m Model) toggleDebugFocus() Model {
	m.debugPanelFocused = !m.debugPanelFocused
	if m.debugPanelFocused {
		m = m.debugLog("Debug panel focused")
		// Refresh state lines when focusing
		m.debugStateLines = strings.Split(m.getDebugState(), "\n")
	}
	return m
}

// debugNavigateUp moves selection up in debug panel
func (m Model) debugNavigateUp() Model {
	if m.debugSelectedSection == 0 {
		// In state section - no navigation yet (could add line-by-line later)
		return m
	} else {
		// In logs section
		if m.debugSelectedLog > 0 {
			m.debugSelectedLog--
			m = m.adjustDebugLogScroll()
		}
	}
	return m
}

// debugNavigateDown moves selection down in debug panel
func (m Model) debugNavigateDown() Model {
	if m.debugSelectedSection == 0 {
		// In state section - no navigation yet
		return m
	} else {
		// In logs section
		if m.debugSelectedLog < len(m.debugLogs)-1 {
			m.debugSelectedLog++
			m = m.adjustDebugLogScroll()
		}
	}
	return m
}

// adjustDebugLogScroll adjusts scroll offset to keep selected log visible
func (m Model) adjustDebugLogScroll() Model {
	// Calculate visible lines based on debug panel height
	// Debug panel is 1/3 of screen height
	debugHeight := m.height / 3
	if debugHeight < 10 {
		debugHeight = 10
	}

	// Available lines for logs (accounting for title, section header, help text)
	visibleLogLines := debugHeight - 6
	if visibleLogLines < 5 {
		visibleLogLines = 5
	}

	// Adjust scroll to keep selection visible
	if m.debugSelectedLog < m.debugLogScrollOffset {
		// Selected log is above visible area, scroll up
		m.debugLogScrollOffset = m.debugSelectedLog
	} else if m.debugSelectedLog >= m.debugLogScrollOffset+visibleLogLines {
		// Selected log is below visible area, scroll down
		m.debugLogScrollOffset = m.debugSelectedLog - visibleLogLines + 1
	}

	return m
}

// debugSwitchSection switches between state and logs sections
func (m Model) debugSwitchSection() Model {
	if m.debugSelectedSection == 0 {
		m.debugSelectedSection = 1
		m.debugSelectedLog = 0
	} else {
		m.debugSelectedSection = 0
	}
	return m
}

// debugOpenDetail opens detail popup for selected item
func (m Model) debugOpenDetail() Model {
	if m.debugSelectedSection == 0 {
		// Show full state
		m.debugDetailMode = true
		m.debugDetailTitle = "App State (Full)"
		m.debugDetailContent = m.getDebugState()
	} else {
		// Show selected log
		if m.debugSelectedLog >= 0 && m.debugSelectedLog < len(m.debugLogs) {
			m.debugDetailMode = true
			m.debugDetailTitle = fmt.Sprintf("Log Entry #%d", m.debugSelectedLog+1)
			m.debugDetailContent = m.debugLogs[m.debugSelectedLog]
		}
	}
	return m
}

// debugCloseDetail closes the detail popup
func (m Model) debugCloseDetail() Model {
	m.debugDetailMode = false
	m.debugDetailContent = ""
	m.debugDetailTitle = ""
	return m
}
