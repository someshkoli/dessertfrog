package config

// ColorSchemeConfig represents the color scheme configuration
type ColorSchemeConfig struct {
	// Primary colors
	Primary    string `yaml:"primary,omitempty"`    // Main accent color (default: #7D56F4)
	Secondary  string `yaml:"secondary,omitempty"`  // Secondary accent (default: #808080)
	Background string `yaml:"background,omitempty"` // Background color (default: #1a1a1a)
	Foreground string `yaml:"foreground,omitempty"` // Foreground text (default: #FFFFFF)

	// Text colors
	Text            string `yaml:"text,omitempty"`             // Regular text color (default: 255)
	HelpText        string `yaml:"help_text,omitempty"`        // Help text color (default: #626262)
	GhostText       string `yaml:"ghost_text,omitempty"`       // Autocomplete ghost text (default: #666666)
	StatusText      string `yaml:"status_text,omitempty"`      // Status line text (default: #AAAAAA)
	InputBackground string `yaml:"input_background,omitempty"` // Text input background (default: 235)

	// Border colors
	BorderActive   string `yaml:"border_active,omitempty"`   // Active border (default: #7D56F4)
	BorderInactive string `yaml:"border_inactive,omitempty"` // Inactive border (default: #4A4A4A)
	BorderNormal   string `yaml:"border_normal,omitempty"`   // Normal border (default: #808080)

	// Selection and highlight
	SelectionBg string `yaml:"selection_bg,omitempty"` // Selection background (default: #7D56F4)
	SelectionFg string `yaml:"selection_fg,omitempty"` // Selection foreground (default: #FFFFFF)

	// Status colors
	StatusConnected    string `yaml:"status_connected,omitempty"`    // Connected (default: #00FF00)
	StatusDisconnected string `yaml:"status_disconnected,omitempty"` // Disconnected (default: #FF0000)
	StatusConnecting   string `yaml:"status_connecting,omitempty"`   // Connecting (default: #FFFF00)
	StatusError        string `yaml:"status_error,omitempty"`        // Error (default: #FF6666)

	// Command and SQL
	CommandBg string `yaml:"command_bg,omitempty"` // Command line background (default: #7D56F4)
	CommandFg string `yaml:"command_fg,omitempty"` // Command line foreground (default: #FAFAFA)

	// Popup and overlay
	PopupBg     string `yaml:"popup_bg,omitempty"`     // Popup background (default: #1a1a1a)
	PopupBorder string `yaml:"popup_border,omitempty"` // Popup border (default: 62)
	OverlayBg   string `yaml:"overlay_bg,omitempty"`   // Overlay background (default: #000000)

	// Status line
	StatusLineBg string `yaml:"status_line_bg,omitempty"` // Status line background (default: #2C2C2C)
	StatusLineFg string `yaml:"status_line_fg,omitempty"` // Status line foreground (default: #FFFFFF)

	// Schema panel colors
	SchemaTitle      string `yaml:"schema_title,omitempty"`       // Table title (default: 205)
	SchemaSection    string `yaml:"schema_section,omitempty"`     // Section headers (default: 39)
	SchemaField      string `yaml:"schema_field,omitempty"`       // Field values (default: 252)
	SchemaColumnName string `yaml:"schema_column_name,omitempty"` // Column names (default: 228)
	SchemaType       string `yaml:"schema_type,omitempty"`        // Data types (default: 141)
	SchemaPrimaryKey string `yaml:"schema_primary_key,omitempty"` // Primary key indicator (default: 205)
	SchemaForeignKey string `yaml:"schema_foreign_key,omitempty"` // Foreign key indicator (default: 214)
	SchemaEmpty      string `yaml:"schema_empty,omitempty"`       // Empty message (default: 241)
	SchemaLoading    string `yaml:"schema_loading,omitempty"`     // Loading message (default: 241)

	// Cell editing colors
	CellPendingEdit string `yaml:"cell_pending_edit,omitempty"` // Pending edit indicator (default: 220)

	// Table data colors
	TableFilter    string `yaml:"table_filter,omitempty"`    // Filter indicator (default: 226)
	TableClipboard string `yaml:"table_clipboard,omitempty"` // Clipboard success (default: 120)

	// SQL history colors
	SQLHistoryTitle    string `yaml:"sql_history_title,omitempty"`    // History title (default: 205)
	SQLHistoryCount    string `yaml:"sql_history_count,omitempty"`    // Count text (default: 240)
	SQLHistorySelected string `yaml:"sql_history_selected,omitempty"` // Selected entry (default: 255)
	SQLHistoryNormal   string `yaml:"sql_history_normal,omitempty"`   // Normal entry (default: 252)
	SQLHistoryBorder   string `yaml:"sql_history_border,omitempty"`   // Border (default: 63)

	// JSON viewer colors
	JSONKey   string `yaml:"json_key,omitempty"`   // JSON keys (default: 39)
	JSONValue string `yaml:"json_value,omitempty"` // JSON values (default: 228)
	JSONType  string `yaml:"json_type,omitempty"`  // JSON type info (default: 241)

	// Debug panel colors
	DebugSection  string `yaml:"debug_section,omitempty"`  // Section headers (default: #7AA2F7)
	DebugLog      string `yaml:"debug_log,omitempty"`      // Log entries (default: #C0C0C0)
	DebugSelected string `yaml:"debug_selected,omitempty"` // Selected item (default: #3E3EF2)
	DebugFocus    string `yaml:"debug_focus,omitempty"`    // Focus indicator (default: #00FF00)
}

// DefaultColorScheme returns the default color scheme
func DefaultColorScheme() ColorSchemeConfig {
	return ColorSchemeConfig{
		Primary:            "#7D56F4",
		Secondary:          "#808080",
		Background:         "#1a1a1a",
		Foreground:         "#FFFFFF",
		Text:               "255",
		HelpText:           "#626262",
		GhostText:          "#666666",
		StatusText:         "#AAAAAA",
		InputBackground:    "235",
		BorderActive:       "#7D56F4",
		BorderInactive:     "#4A4A4A",
		BorderNormal:       "#808080",
		SelectionBg:        "#7D56F4",
		SelectionFg:        "#FFFFFF",
		StatusConnected:    "#00FF00",
		StatusDisconnected: "#FF0000",
		StatusConnecting:   "#FFFF00",
		StatusError:        "#FF6666",
		CommandBg:          "#7D56F4",
		CommandFg:          "#FAFAFA",
		PopupBg:            "#1a1a1a",
		PopupBorder:        "62",
		OverlayBg:          "#000000",
		StatusLineBg:       "#2C2C2C",
		StatusLineFg:       "#FFFFFF",
		SchemaTitle:        "205",
		SchemaSection:      "39",
		SchemaField:        "252",
		SchemaColumnName:   "228",
		SchemaType:         "141",
		SchemaPrimaryKey:   "205",
		SchemaForeignKey:   "214",
		SchemaEmpty:        "241",
		SchemaLoading:      "241",
		CellPendingEdit:    "220",
		TableFilter:        "226",
		TableClipboard:     "120",
		SQLHistoryTitle:    "205",
		SQLHistoryCount:    "240",
		SQLHistorySelected: "255",
		SQLHistoryNormal:   "252",
		SQLHistoryBorder:   "63",
		JSONKey:            "39",
		JSONValue:          "228",
		JSONType:           "241",
		DebugSection:       "#7AA2F7",
		DebugLog:           "#C0C0C0",
		DebugSelected:      "#3E3EF2",
		DebugFocus:         "#00FF00",
	}
}

// MergeWithDefaults merges the color scheme config with defaults
func (c *ColorSchemeConfig) MergeWithDefaults() ColorSchemeConfig {
	defaults := DefaultColorScheme()
	result := *c

	// Use reflection or manual checks - manual is clearer here
	if result.Primary == "" {
		result.Primary = defaults.Primary
	}
	if result.Secondary == "" {
		result.Secondary = defaults.Secondary
	}
	if result.Background == "" {
		result.Background = defaults.Background
	}
	if result.Foreground == "" {
		result.Foreground = defaults.Foreground
	}
	if result.Text == "" {
		result.Text = defaults.Text
	}
	if result.HelpText == "" {
		result.HelpText = defaults.HelpText
	}
	if result.GhostText == "" {
		result.GhostText = defaults.GhostText
	}
	if result.StatusText == "" {
		result.StatusText = defaults.StatusText
	}
	if result.InputBackground == "" {
		result.InputBackground = defaults.InputBackground
	}
	if result.BorderActive == "" {
		result.BorderActive = defaults.BorderActive
	}
	if result.BorderInactive == "" {
		result.BorderInactive = defaults.BorderInactive
	}
	if result.BorderNormal == "" {
		result.BorderNormal = defaults.BorderNormal
	}
	if result.SelectionBg == "" {
		result.SelectionBg = defaults.SelectionBg
	}
	if result.SelectionFg == "" {
		result.SelectionFg = defaults.SelectionFg
	}
	if result.StatusConnected == "" {
		result.StatusConnected = defaults.StatusConnected
	}
	if result.StatusDisconnected == "" {
		result.StatusDisconnected = defaults.StatusDisconnected
	}
	if result.StatusConnecting == "" {
		result.StatusConnecting = defaults.StatusConnecting
	}
	if result.StatusError == "" {
		result.StatusError = defaults.StatusError
	}
	if result.CommandBg == "" {
		result.CommandBg = defaults.CommandBg
	}
	if result.CommandFg == "" {
		result.CommandFg = defaults.CommandFg
	}
	if result.PopupBg == "" {
		result.PopupBg = defaults.PopupBg
	}
	if result.PopupBorder == "" {
		result.PopupBorder = defaults.PopupBorder
	}
	if result.OverlayBg == "" {
		result.OverlayBg = defaults.OverlayBg
	}
	if result.StatusLineBg == "" {
		result.StatusLineBg = defaults.StatusLineBg
	}
	if result.StatusLineFg == "" {
		result.StatusLineFg = defaults.StatusLineFg
	}
	if result.SchemaTitle == "" {
		result.SchemaTitle = defaults.SchemaTitle
	}
	if result.SchemaSection == "" {
		result.SchemaSection = defaults.SchemaSection
	}
	if result.SchemaField == "" {
		result.SchemaField = defaults.SchemaField
	}
	if result.SchemaColumnName == "" {
		result.SchemaColumnName = defaults.SchemaColumnName
	}
	if result.SchemaType == "" {
		result.SchemaType = defaults.SchemaType
	}
	if result.SchemaPrimaryKey == "" {
		result.SchemaPrimaryKey = defaults.SchemaPrimaryKey
	}
	if result.SchemaForeignKey == "" {
		result.SchemaForeignKey = defaults.SchemaForeignKey
	}
	if result.SchemaEmpty == "" {
		result.SchemaEmpty = defaults.SchemaEmpty
	}
	if result.SchemaLoading == "" {
		result.SchemaLoading = defaults.SchemaLoading
	}
	if result.CellPendingEdit == "" {
		result.CellPendingEdit = defaults.CellPendingEdit
	}
	if result.TableFilter == "" {
		result.TableFilter = defaults.TableFilter
	}
	if result.TableClipboard == "" {
		result.TableClipboard = defaults.TableClipboard
	}
	if result.SQLHistoryTitle == "" {
		result.SQLHistoryTitle = defaults.SQLHistoryTitle
	}
	if result.SQLHistoryCount == "" {
		result.SQLHistoryCount = defaults.SQLHistoryCount
	}
	if result.SQLHistorySelected == "" {
		result.SQLHistorySelected = defaults.SQLHistorySelected
	}
	if result.SQLHistoryNormal == "" {
		result.SQLHistoryNormal = defaults.SQLHistoryNormal
	}
	if result.SQLHistoryBorder == "" {
		result.SQLHistoryBorder = defaults.SQLHistoryBorder
	}
	if result.JSONKey == "" {
		result.JSONKey = defaults.JSONKey
	}
	if result.JSONValue == "" {
		result.JSONValue = defaults.JSONValue
	}
	if result.JSONType == "" {
		result.JSONType = defaults.JSONType
	}
	if result.DebugSection == "" {
		result.DebugSection = defaults.DebugSection
	}
	if result.DebugLog == "" {
		result.DebugLog = defaults.DebugLog
	}
	if result.DebugSelected == "" {
		result.DebugSelected = defaults.DebugSelected
	}
	if result.DebugFocus == "" {
		result.DebugFocus = defaults.DebugFocus
	}

	return result
}
