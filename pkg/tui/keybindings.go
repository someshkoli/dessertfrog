package tui

// Command represents an action that can be performed in the application
type Command string

const (
	// Navigation commands
	CommandNavigateUp           Command = "navigate_up"
	CommandNavigateDown         Command = "navigate_down"
	CommandNavigateLeft         Command = "navigate_left"
	CommandNavigateRight        Command = "navigate_right"
	CommandNavigateLastColumn   Command = "navigate_last_column"
	CommandNavigateFirstColumn  Command = "navigate_first_column"
	CommandNavigateWordForward  Command = "navigate_word_forward"
	CommandNavigateWordBackward Command = "navigate_word_backward"
	CommandPageUp               Command = "page_up"
	CommandPageDown             Command = "page_down"
	CommandGoToTop              Command = "go_to_top"
	CommandGoToBottom           Command = "go_to_bottom"
	CommandHistoryBack          Command = "history_back"
	CommandHistoryForward       Command = "history_forward"

	// Action commands
	CommandConfirm Command = "confirm"
	CommandCancel  Command = "cancel"
	CommandQuit    Command = "quit"
	CommandBack    Command = "back"

	// Export
	CommandExport Command = "export_data"

	// Text input commands
	CommandCursorLeft    Command = "cursor_left"
	CommandCursorRight   Command = "cursor_right"
	CommandCursorHome    Command = "cursor_home"
	CommandCursorEnd     Command = "cursor_end"
	CommandDeleteChar    Command = "delete_char"
	CommandBackspace     Command = "backspace"
	CommandInsertNewline Command = "insert_newline"

	// View commands
	CommandOpenTable             Command = "open_table"
	CommandOpenSearch            Command = "open_search"
	CommandOpenSQLQuery          Command = "open_sql_query"
	CommandOpenCommandMode       Command = "open_command_mode"
	CommandInlineSearch          Command = "inline_search"
	CommandOpenConnectionManager Command = "open_connection_manager"

	// Table view commands
	CommandNextPage       Command = "next_page"
	CommandPreviousPage   Command = "previous_page"
	CommandEditCell       Command = "edit_cell"
	CommandDeleteRow      Command = "CommandDeleteRow"
	CommandCopyCellValue  Command = "copy_cell_value"
	CommandCopyRow        Command = "copy_row"
	CommandOpenCellPopup  Command = "open_cell_popup"
	CommandOpenRecordView Command = "open_record_view"
	CommandFilterContent  Command = "filter_content"

	// Debug commands
	CommandToggleDebug      Command = "toggle_debug"
	CommandClearDebugLogs   Command = "clear_debug_logs"
	CommandToggleDebugFocus Command = "toggle_debug_focus"

	// Popup commands
	CommandToggleJSONNode     Command = "toggle_json_node"
	CommandSwitchDebugSection Command = "switch_debug_section"
)

// KeyBinding maps a key combination to a command
type KeyBinding struct {
	Key     string
	Command Command
}

// KeyBindings holds all key bindings for different modes
type KeyBindings struct {
	Global       []KeyBinding // Keys that work in all modes
	Normal       []KeyBinding // Keys in normal mode (table list)
	TableView    []KeyBinding // Keys in table view mode
	Search       []KeyBinding // Keys in search popup
	SearchNormal []KeyBinding // Keys in search popup
	InlineSearch []KeyBinding // Keys in inline search mode
	CommandMode  []KeyBinding // Keys in command mode
	CellEdit     []KeyBinding // Keys in cell edit mode
	SQLQuery     []KeyBinding // Keys in SQL query mode
	CellPopup    []KeyBinding // Keys in cell value popup
	RecordView   []KeyBinding // Keys in record view popup
	DebugPanel   []KeyBinding // Keys in debug panel
	DebugDetail  []KeyBinding // Keys in debug detail popup
}

// DefaultKeyBindings returns the default key binding configuration
func DefaultKeyBindings() KeyBindings {
	return KeyBindings{
		Global: []KeyBinding{
			{"f12", CommandToggleDebug},
			{"f11", CommandClearDebugLogs},
			{"f10", CommandToggleDebugFocus},
			{"ctrl+c", CommandQuit},
		},
		Normal: []KeyBinding{
			// Navigation
			{"j", CommandNavigateDown},
			{"down", CommandNavigateDown},
			{"k", CommandNavigateUp},
			{"up", CommandNavigateUp},
			{"g", CommandGoToTop},
			{"G", CommandGoToBottom},
			{"ctrl+d", CommandPageDown},
			{"ctrl+u", CommandPageUp},
			{"ctrl+h", CommandHistoryBack},
			{"ctrl+l", CommandHistoryForward},

			// Actions
			{"enter", CommandOpenTable},
			{"ctrl+p", CommandOpenSearch},
			{"s", CommandOpenSQLQuery},
			{":", CommandOpenCommandMode},
			{"/", CommandInlineSearch},
			{"c", CommandOpenConnectionManager},
			{"d", CommandOpenConnectionManager},
			{"q", CommandQuit},
		},
		TableView: []KeyBinding{
			// Navigation
			{"j", CommandNavigateDown},
			{"down", CommandNavigateDown},
			{"k", CommandNavigateUp},
			{"up", CommandNavigateUp},
			{"h", CommandNavigateLeft},
			{"left", CommandNavigateLeft},
			{"l", CommandNavigateRight},
			{"right", CommandNavigateRight},
			{"$", CommandNavigateLastColumn},
			{"_", CommandNavigateFirstColumn},
			{"0", CommandNavigateFirstColumn},
			{"w", CommandNavigateWordForward},
			{"b", CommandNavigateWordBackward},
			{"g", CommandGoToTop},
			{"G", CommandGoToBottom},
			{"ctrl+d", CommandPageDown},
			{"ctrl+u", CommandPageUp},
			{"ctrl+h", CommandHistoryBack},
			{"ctrl+l", CommandHistoryForward},

			// Pagination
			{"n", CommandNextPage},
			{"p", CommandPreviousPage},

			// Actions
			{"i", CommandEditCell},
			{"y", CommandCopyCellValue},
			{"Y", CommandCopyRow},
			{"enter", CommandOpenCellPopup},
			{"c", CommandOpenConnectionManager},
			{"v", CommandOpenCellPopup},
			{"r", CommandOpenRecordView},
			{"V", CommandOpenRecordView},
			{"/", CommandFilterContent},
			{"ctrl+p", CommandOpenSearch},
			{"s", CommandOpenSQLQuery},
			{"esc", CommandBack},
			{"q", CommandBack},

			// Export
			{"q", CommandBack},
		},
		Search: []KeyBinding{
			{"enter", CommandConfirm},
			{"esc", CommandCancel},
			{"up", CommandNavigateUp},
			{"down", CommandNavigateDown},
			{"ctrl+n", CommandNavigateDown},
			{"ctrl+p", CommandNavigateUp},
		},
		InlineSearch: []KeyBinding{
			{"enter", CommandConfirm},
			{"esc", CommandCancel},
		},
		CommandMode: []KeyBinding{
			{"enter", CommandConfirm},
			{"esc", CommandCancel},
		},
		CellEdit: []KeyBinding{
			{"esc", CommandCancel},
			{"enter", CommandInsertNewline},
		},
		SQLQuery: []KeyBinding{
			{"esc", CommandCancel},
			{"home", CommandCursorHome},
			{"end", CommandCursorEnd},
		},
		CellPopup: []KeyBinding{
			{"esc", CommandCancel},
			{"q", CommandCancel},
			{"v", CommandOpenCellPopup},
			{"y", CommandCopyCellValue},
			{"j", CommandNavigateDown},
			{"k", CommandNavigateUp},
			{"down", CommandNavigateDown},
			{"up", CommandNavigateUp},
			{"h", CommandNavigateLeft},
			{"left", CommandNavigateLeft},
			{"l", CommandNavigateRight},
			{"right", CommandNavigateRight},
			{"enter", CommandToggleJSONNode},
			{"g", CommandGoToTop},
			{"G", CommandGoToBottom},
		},
		RecordView: []KeyBinding{
			{"esc", CommandCancel},
			{"q", CommandCancel},
			{"V", CommandCancel},
			{"j", CommandNavigateDown},
			{"k", CommandNavigateUp},
			{"down", CommandNavigateDown},
			{"up", CommandNavigateUp},
			{"v", CommandOpenCellPopup},
			{"y", CommandCopyCellValue},
			{"g", CommandGoToTop},
			{"G", CommandGoToBottom},
		},
		DebugPanel: []KeyBinding{
			{"j", CommandNavigateDown},
			{"k", CommandNavigateUp},
			{"down", CommandNavigateDown},
			{"up", CommandNavigateUp},
			{"tab", CommandSwitchDebugSection},
			{"enter", CommandConfirm},
			{"esc", CommandCancel},
			{"g", CommandGoToTop},
			{"G", CommandGoToBottom},
		},
		DebugDetail: []KeyBinding{
			{"esc", CommandCancel},
			{"q", CommandCancel},
			{"enter", CommandCancel},
		},
	}
}

// getCommand returns the command for a given key in a specific binding set
func getCommand(key string, bindings []KeyBinding) (Command, bool) {
	for _, binding := range bindings {
		if binding.Key == key {
			return binding.Command, true
		}
	}
	return "", false
}

// MergeKeyBindings merges config key bindings with default key bindings
// Config bindings override defaults if the key already exists
func MergeKeyBindings(defaults KeyBindings, configBindings map[string][]KeyBinding) KeyBindings {
	merged := defaults

	// Merge each mode if provided in config
	if bindings, ok := configBindings["global"]; ok && len(bindings) > 0 {
		merged.Global = mergeBindingList(defaults.Global, bindings)
	}
	if bindings, ok := configBindings["normal"]; ok && len(bindings) > 0 {
		merged.Normal = mergeBindingList(defaults.Normal, bindings)
	}
	if bindings, ok := configBindings["table_view"]; ok && len(bindings) > 0 {
		merged.TableView = mergeBindingList(defaults.TableView, bindings)
	}
	if bindings, ok := configBindings["search"]; ok && len(bindings) > 0 {
		merged.Search = mergeBindingList(defaults.Search, bindings)
	}
	if bindings, ok := configBindings["inline_search"]; ok && len(bindings) > 0 {
		merged.InlineSearch = mergeBindingList(defaults.InlineSearch, bindings)
	}
	if bindings, ok := configBindings["command_mode"]; ok && len(bindings) > 0 {
		merged.CommandMode = mergeBindingList(defaults.CommandMode, bindings)
	}
	if bindings, ok := configBindings["cell_edit"]; ok && len(bindings) > 0 {
		merged.CellEdit = mergeBindingList(defaults.CellEdit, bindings)
	}
	if bindings, ok := configBindings["sql_query"]; ok && len(bindings) > 0 {
		merged.SQLQuery = mergeBindingList(defaults.SQLQuery, bindings)
	}
	if bindings, ok := configBindings["cell_popup"]; ok && len(bindings) > 0 {
		merged.CellPopup = mergeBindingList(defaults.CellPopup, bindings)
	}
	if bindings, ok := configBindings["record_view"]; ok && len(bindings) > 0 {
		merged.RecordView = mergeBindingList(defaults.RecordView, bindings)
	}
	if bindings, ok := configBindings["debug_panel"]; ok && len(bindings) > 0 {
		merged.DebugPanel = mergeBindingList(defaults.DebugPanel, bindings)
	}
	if bindings, ok := configBindings["debug_detail"]; ok && len(bindings) > 0 {
		merged.DebugDetail = mergeBindingList(defaults.DebugDetail, bindings)
	}

	return merged
}

// mergeBindingList merges two binding lists, with override taking precedence for duplicate keys
func mergeBindingList(defaults []KeyBinding, overrides []KeyBinding) []KeyBinding {
	// Create a map of keys from overrides
	overrideMap := make(map[string]Command)
	for _, binding := range overrides {
		overrideMap[binding.Key] = binding.Command
	}

	// Start with defaults, replacing any that are overridden
	result := make([]KeyBinding, 0, len(defaults))
	for _, binding := range defaults {
		if cmd, exists := overrideMap[binding.Key]; exists {
			// Use override command
			result = append(result, KeyBinding{Key: binding.Key, Command: cmd})
			delete(overrideMap, binding.Key) // Mark as processed
		} else {
			// Keep default
			result = append(result, binding)
		}
	}

	// Add any new bindings from overrides that weren't in defaults
	for key, cmd := range overrideMap {
		result = append(result, KeyBinding{Key: key, Command: cmd})
	}

	return result
}
