package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/someshkoli/dessertfrog/pkg/config"
)

// Styles holds all styled components for the TUI
type Styles struct {
	// Title and text styles
	TitleStyle                lipgloss.Style
	HelpStyle                 lipgloss.Style

	// Border styles
	BorderStyle               lipgloss.Style
	ActiveBorderStyle         lipgloss.Style
	InactiveBorderStyle       lipgloss.Style
	InfoBorderStyle           lipgloss.Style
	ScreenBorderStyle         lipgloss.Style
	StatusBoxStyle            lipgloss.Style

	// Connection status styles
	ConnectedStyle            lipgloss.Style
	DisconnectedStyle         lipgloss.Style
	ConnectingStyle           lipgloss.Style

	// Command mode styles
	CommandLineStyle          lipgloss.Style

	// Table selection styles
	SelectedRowStyle          lipgloss.Style

	// Search popup styles
	PopupStyle                lipgloss.Style
	SearchInputStyle          lipgloss.Style
	ActiveSearchInputStyle    lipgloss.Style
	InactiveSearchInputStyle  lipgloss.Style
	OverlayStyle              lipgloss.Style

	// Ghost text style
	GhostTextStyle            lipgloss.Style

	// Status line styles
	StatusLineStyle           lipgloss.Style
	StatusLineLeftStyle       lipgloss.Style

	// Error box style
	ErrorBoxStyle             lipgloss.Style

	// Schema panel styles
	SchemaTitleStyle          lipgloss.Style
	SchemaSectionStyle        lipgloss.Style
	SchemaFieldStyle          lipgloss.Style
	SchemaColumnNameStyle     lipgloss.Style
	SchemaTypeStyle           lipgloss.Style
	SchemaPrimaryKeyStyle     lipgloss.Style
	SchemaForeignKeyStyle     lipgloss.Style
	SchemaEmptyStyle          lipgloss.Style
	SchemaLoadingStyle        lipgloss.Style

	// Cell edit styles
	CellEditInputBoxStyle     lipgloss.Style
	CellEditPopupStyle        lipgloss.Style
	CellPendingEditStyle      lipgloss.Style

	// Table data styles
	TableFilterStyle          lipgloss.Style
	TableClipboardStyle       lipgloss.Style

	// SQL history styles
	SQLHistoryTitleStyle      lipgloss.Style
	SQLHistoryCountStyle      lipgloss.Style
	SQLHistorySelectedStyle   lipgloss.Style
	SQLHistoryNormalStyle     lipgloss.Style
	SQLHistoryBorderStyle     lipgloss.Style

	// Record/Cell popup styles
	RecordPopupStyle          lipgloss.Style
	RecordKeyStyle            lipgloss.Style
	RecordValueStyle          lipgloss.Style
	RecordJSONIndicatorStyle  lipgloss.Style

	// JSON tree styles (for cell popup)
	JSONKeyStyle              lipgloss.Style
	JSONValueStyle            lipgloss.Style
	JSONTypeStyle             lipgloss.Style

	// Debug panel styles
	DebugBorderStyle          lipgloss.Style
	DebugTitleStyle           lipgloss.Style
	DebugSectionStyle         lipgloss.Style
	DebugLogStyle             lipgloss.Style
	DebugSelectedStyle        lipgloss.Style
	DebugFocusIndicatorStyle  lipgloss.Style
	DebugLeftColumnStyle      lipgloss.Style
	DebugRightColumnStyle     lipgloss.Style
	DebugContentStyle         lipgloss.Style

	// Connection manager styles
	ConnManagerPopupStyle       lipgloss.Style
	ConnManagerTitleStyle       lipgloss.Style
	ConnManagerFilterStyle      lipgloss.Style
	ConnManagerRowStyle         lipgloss.Style
	ConnManagerInsertModeStyle  lipgloss.Style
	ConnManagerNormalModeStyle  lipgloss.Style
	ScrollIndicatorStyle        lipgloss.Style
	ErrorStyle                  lipgloss.Style
	TableRowStyle               lipgloss.Style

	// Passphrase prompt styles
	PassphrasePromptStyle       lipgloss.Style
	PassphraseTitleStyle        lipgloss.Style
	PassphraseInputStyle        lipgloss.Style
	PassphraseKeyInfoStyle      lipgloss.Style
}

// NewStyles creates a new Styles instance from a color scheme
func NewStyles(scheme config.ColorSchemeConfig) Styles {
	return Styles{
		// Title and text styles
		TitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(scheme.Primary)),

		HelpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.HelpText)),

		// Border styles
		BorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.BorderNormal)).
			Padding(0, 1),

		ActiveBorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.BorderActive)).
			Padding(0, 1),

		InactiveBorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.BorderInactive)).
			Padding(0, 1),

		InfoBorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.BorderNormal)).
			Padding(0, 1),

		ScreenBorderStyle: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color(scheme.BorderNormal)).
			Padding(0, 1),

		StatusBoxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.BorderNormal)).
			Padding(0, 1),

		// Connection status styles
		ConnectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.StatusConnected)).
			Bold(true),

		DisconnectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.StatusDisconnected)).
			Bold(true),

		ConnectingStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.StatusConnecting)).
			Bold(true),

		// Command mode styles
		CommandLineStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.CommandFg)).
			Background(lipgloss.Color(scheme.CommandBg)).
			Padding(0, 1),

		// Table selection styles
		SelectedRowStyle: lipgloss.NewStyle().
			Background(lipgloss.Color(scheme.SelectionBg)).
			Foreground(lipgloss.Color(scheme.SelectionFg)).
			Bold(true),

		// Search popup styles
		PopupStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.PopupBorder)).
			Padding(1, 2).
			Background(lipgloss.Color(scheme.PopupBg)),

		SearchInputStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.BorderNormal)).
			Padding(0, 1),

		ActiveSearchInputStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.BorderActive)).
			Padding(0, 1),

		InactiveSearchInputStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.BorderInactive)).
			Padding(0, 1),

		OverlayStyle: lipgloss.NewStyle().
			Background(lipgloss.Color(scheme.OverlayBg)).
			Foreground(lipgloss.Color(scheme.Foreground)),

		// Ghost text style
		GhostTextStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.GhostText)).
			Faint(true),

		// Status line styles
		StatusLineStyle: lipgloss.NewStyle().
			Background(lipgloss.Color(scheme.StatusLineBg)).
			Foreground(lipgloss.Color(scheme.StatusLineFg)).
			Padding(0, 0).
			Inline(true),

		StatusLineLeftStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.StatusText)),

		// Error box style
		ErrorBoxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.StatusError)).
			Padding(2, 4).
			Foreground(lipgloss.Color(scheme.StatusError)).
			Bold(true),

		// Schema panel styles
		SchemaTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(scheme.SchemaTitle)).
			Underline(true),

		SchemaSectionStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(scheme.SchemaSection)),

		SchemaFieldStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SchemaField)),

		SchemaColumnNameStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SchemaColumnName)),

		SchemaTypeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SchemaType)),

		SchemaPrimaryKeyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SchemaPrimaryKey)).
			Bold(true),

		SchemaForeignKeyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SchemaForeignKey)),

		SchemaEmptyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SchemaEmpty)).
			Italic(true),

		SchemaLoadingStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SchemaLoading)).
			Italic(true),

		// Cell edit styles
		CellEditInputBoxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.PopupBorder)).
			Padding(0, 1).
			Height(3),

		CellEditPopupStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.PopupBorder)).
			Padding(1, 2),

		CellPendingEditStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.CellPendingEdit)),

		// Table data styles
		TableFilterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.TableFilter)).
			Bold(true),

		TableClipboardStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.TableClipboard)).
			Bold(true),

		// SQL history styles
		SQLHistoryTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(scheme.SQLHistoryTitle)),

		SQLHistoryCountStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SQLHistoryCount)),

		SQLHistorySelectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SQLHistorySelected)).
			Background(lipgloss.Color(scheme.SelectionBg)).
			Bold(true),

		SQLHistoryNormalStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SQLHistoryNormal)),

		SQLHistoryBorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.SQLHistoryBorder)).
			Padding(0, 1),

		// Record/Cell popup styles
		RecordPopupStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.PopupBorder)).
			Padding(1, 2),

		RecordKeyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.JSONKey)).
			Bold(true),

		RecordValueStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SchemaField)),

		RecordJSONIndicatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.JSONType)).
			Italic(true),

		// JSON tree styles
		JSONKeyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.JSONKey)).
			Bold(true),

		JSONValueStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.JSONValue)),

		JSONTypeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.JSONType)).
			Italic(true),

		// Debug panel styles
		DebugBorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.BorderNormal)).
			Padding(0, 1),

		DebugTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(scheme.Primary)),

		DebugSectionStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(scheme.DebugSection)),

		DebugLogStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.DebugLog)),

		DebugSelectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.Foreground)).
			Background(lipgloss.Color(scheme.DebugSelected)).
			Bold(true),

		DebugFocusIndicatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.DebugFocus)).
			Bold(true),

		DebugLeftColumnStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.DebugSection)).
			Bold(true),

		DebugRightColumnStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SchemaField)),

		DebugContentStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.SchemaField)),

		// Connection manager styles
		ConnManagerPopupStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.PopupBorder)).
			Padding(1, 2).
			Background(lipgloss.Color(scheme.PopupBg)),

		ConnManagerTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(scheme.Primary)),

		ConnManagerFilterStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.BorderActive)).
			Padding(0, 1),

		ConnManagerRowStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.Foreground)),

		ConnManagerInsertModeStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("10")). // Green background for insert mode
			Foreground(lipgloss.Color("0")).  // Black text
			Bold(true),

		ConnManagerNormalModeStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("12")). // Blue background for normal mode
			Foreground(lipgloss.Color("0")).  // Black text
			Bold(true),

		ScrollIndicatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.GhostText)).
			Italic(true),

		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.StatusError)).
			Italic(true),

		TableRowStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(scheme.Foreground)),

		// Passphrase prompt styles
		PassphrasePromptStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.PopupBorder)).
			Padding(1, 2).
			Background(lipgloss.Color(scheme.PopupBg)),

		PassphraseTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(scheme.Primary)),

		PassphraseInputStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(scheme.BorderActive)).
			Padding(0, 1),

		PassphraseKeyInfoStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true),
	}
}
