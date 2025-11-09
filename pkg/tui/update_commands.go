package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/someshkoli/dessertfrog/pkg/driver"
)

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

// fetchTableSchema fetches the full schema for a table (including primary keys)
func fetchTableSchema(drv driver.Driver, schemaName, tableName string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Fetch full table schema with column details
		schema, err := drv.GetTableSchema(ctx, schemaName, tableName)
		if err != nil {
			// If schema fetch fails, it's not critical - we can still work without PK info
			return tableSchemaLoadFailedMsg{err: err}
		}

		return tableSchemaLoadedMsg{
			schema: schema,
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
		err := drv.UpdateCell(ctx, tableSchema, columns, oldRow, columnName, newValue)
		if err != nil {
			return cellUpdateFailedMsg{err: err}
		}

		// Return success with the new value (don't refresh data)
		return cellUpdateSuccessMsg{
			newValue: newValue,
		}
	}
}

// clearClipboardMessage returns a command that clears the clipboard message after a delay
func clearClipboardMessage() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return clearClipboardMsgType{}
	})
}
