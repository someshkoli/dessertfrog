package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// saveCellToBuffer saves the current cell edit to the buffer and closes the popup
func (m Model) saveCellToBuffer() (tea.Model, tea.Cmd) {
	// Create buffer key from row:col indices
	bufferKey := fmt.Sprintf("%d:%d", m.cellEditRowIdx, m.cellEditColIdx)

	// Save to buffer
	m.cellEditBuffer[bufferKey] = m.cellEditValue
	m.cellEditBufferCount = len(m.cellEditBuffer)

	// Update the visual display immediately (local update, not DB)
	if m.cellEditRowIdx >= 0 && m.cellEditRowIdx < len(m.tableData) {
		if m.cellEditColIdx >= 0 && m.cellEditColIdx < len(m.tableData[m.cellEditRowIdx]) {
			m.tableData[m.cellEditRowIdx][m.cellEditColIdx] = m.cellEditValue
		}
	}

	m = m.debugLog("Saved cell to buffer: row=%d, col=%d, pending=%d",
		m.cellEditRowIdx, m.cellEditColIdx, m.cellEditBufferCount)

	// Close the edit popup
	m.cellEditMode = false
	m.cellEditValue = ""
	m.cellEditCursor = 0
	m.cellEditCommandMode = false
	m.cellEditCommand = ""

	return m, nil
}

// batchUpdateCells updates all cells in the buffer to the database
func (m Model) batchUpdateCells() (tea.Model, tea.Cmd) {
	if len(m.cellEditBuffer) == 0 {
		// No pending changes
		m.cellEditMode = false
		m.cellEditCommandMode = false
		m.cellEditCommand = ""
		return m, nil
	}

	m = m.debugLog("Starting batch update for %d cells", len(m.cellEditBuffer))

	// Create commands for all cell updates
	var cmds []tea.Cmd
	for bufferKey, newValue := range m.cellEditBuffer {
		var rowIdx, colIdx int
		fmt.Sscanf(bufferKey, "%d:%d", &rowIdx, &colIdx)

		m = m.debugLog("Updating cell: row=%d, col=%d, value=%s", rowIdx, colIdx, newValue)

		// Create update command for this cell
		cmd := updateCellValue(
			m.driver,
			m.currentViewTable,
			m.tableColumns,
			m.tableData,
			rowIdx,
			colIdx,
			newValue,
			m.tableDataOffset,
		)
		cmds = append(cmds, cmd)
	}

	// Clear the buffer after scheduling updates
	m.cellEditBuffer = make(map[string]string)
	m.cellEditBufferCount = 0

	// Close edit mode
	m.cellEditMode = false
	m.cellEditCommandMode = false
	m.cellEditCommand = ""

	// Execute all update commands in batch
	return m, tea.Batch(cmds...)
}
