package tui

import "github.com/someshkoli/dessertfrog/pkg/driver"

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

// Cell update messages
type cellUpdateSuccessMsg struct {
	newValue string
}
type cellUpdateFailedMsg struct {
	err error
}
