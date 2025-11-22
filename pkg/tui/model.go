package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/someshkoli/dessertfrog/pkg/connhistory"
	"github.com/someshkoli/dessertfrog/pkg/driver"
	"github.com/someshkoli/dessertfrog/pkg/encryption"
	"github.com/someshkoli/dessertfrog/pkg/sqlhistory"
)

// DBConfig holds database connection information
type DBConfig struct {
	Driver   string
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Schema   string
}

// ConnectionStatus represents the state of database connection
type ConnectionStatus int

const (
	Disconnected ConnectionStatus = iota
	Connecting
	Connected
	ConnectionFailed
)

// ConnectionState manages database connection state
type ConnectionState struct {
	dbConfig         DBConfig
	driver           driver.Driver
	connectionStatus ConnectionStatus
	connectionError  string
}

// JSONNode represents a node in the JSON tree for interactive viewing
type JSONNode struct {
	Key         string      // Key name (for objects) or index (for arrays)
	Value       interface{} // The actual value
	Type        string      // "object", "array", "string", "number", "boolean", "null"
	Expanded    bool        // Whether the node is expanded (for objects/arrays)
	Depth       int         // Nesting depth for indentation
	HasChildren bool        // Whether this node has children
}

// TableListViewState manages the main table list view
type TableListViewState struct {
	tables                 []driver.TableSchema // Only tables (shown in main view)
	allEntities            []driver.TableSchema // All entities (used for search)
	tablesLoading          bool
	tablesError            string
	scrollOffset           int    // Current scroll position in table list
	selectedRow            int    // Currently selected row index
	inlineSearchMode       bool   // Whether inline search is active on main view
	inlineSearchQuery      string // Inline search query
	inlineSearchSuggestion string // Inline autocomplete suggestion
	searchMode             bool   // Whether search popup is open
	searchQuery            string // Popup search query
	searchSuggestion       string // Popup autocomplete suggestion
	filteredTables         []driver.TableSchema
	searchSelected         int // Selected index in filtered results
}

// SchemaPanelState manages the schema panel (right side of home screen)
type SchemaPanelState struct {
	schemaPanelFocused   bool                // Whether schema panel has focus (vs tables list)
	schemaPanelSelected  int                 // Selected index in schema panel
	schemaPanelScroll    int                 // Scroll offset in schema panel
	schemaPanelLineCount int                 // Total number of lines in schema panel
	schemaInfo           *driver.TableSchema // Cached detailed schema for selected table
	schemaInfoLoading    bool                // Whether schema info is loading
}

// TableListState manages the table list view and schema panel
type TableListState struct {
	TableListViewState
	SchemaPanelState
}

// TableDataState manages table data viewing and pagination
type TableDataState struct {
	tableViewMode      bool                // Whether table data view is active
	currentViewTable   *driver.TableSchema // Table being viewed
	tableColumns       []string            // Column names
	tableData          [][]string          // Table data rows
	tableDataScrollX   int                 // Horizontal scroll position
	tableDataScrollY   int                 // Vertical scroll position (for viewport)
	selectedDataRow    int                 // Currently selected row in table data view
	selectedDataCol    int                 // Currently selected column in table data view
	tableDataOffset    int                 // Current offset for pagination (0, 500, 1000, etc.)
	tableDataLoading   bool                // Whether table data is loading
	tableDataError     string              // Error loading table data
	tableContentFilter string              // Content search filter
	allTableData       [][]string          // Unfiltered table data (backup for filtering)
	queryTime          string              // Time taken to execute query (e.g., "15ms")
	fetchTime          string              // Time taken to fetch data (e.g., "23ms")
	deletedRows        map[int]bool        // Track which rows are marked for deletion (row index -> true)
	deletedRowsCount   int                 // Number of rows marked for deletion
	lastKeyPress       rune                // Last key pressed (for detecting dd sequence)
	visualMode         bool                // Whether in visual mode (for selecting multiple rows)
	visualStartRow     int                 // Starting row of visual selection
}

// CellValuePopupState manages cell value popup for viewing single cell content
type CellValuePopupState struct {
	cellValuePopupMode     bool       // Whether cell value popup is active
	cellValuePopupContent  string     // The cell value to display
	cellValuePopupIsJSON   bool       // Whether the value is JSON
	cellValuePopupTree     []JSONNode // JSON tree structure
	cellValuePopupScroll   int        // Scroll position in popup
	cellValuePopupSelected int        // Selected node in JSON tree
}

// RecordViewState manages record view popup (entire row as key-value pairs)
type RecordViewState struct {
	recordViewMode     bool     // Whether record view popup is active
	recordViewData     []string // The row data (all column values)
	recordViewColumns  []string // The column names
	recordViewSelected int      // Currently selected field index
	recordViewScroll   int      // Scroll position in record view
}

// ClipboardState manages clipboard operations and notifications
type ClipboardState struct {
	clipboardMessage string // Temporary message shown after copying to clipboard
}

// CellEditState manages cell editing mode with buffer for multi-cell edits
type CellEditState struct {
	cellEditMode        bool              // Whether cell edit popup is active
	cellEditValue       string            // The cell value being edited
	cellEditCursor      int               // Cursor position in edit input
	cellEditRowIdx      int               // Row index of cell being edited
	cellEditColIdx      int               // Column index of cell being edited
	cellEditCommandMode bool              // Whether in command mode (:w to save)
	cellEditCommand     string            // Command buffer for :w
	cellEditBuffer      map[string]string // Buffer of pending cell edits: "rowIdx:colIdx" -> newValue
	cellEditBufferCount int               // Number of pending edits
}

// CellOperationsState manages cell value viewing, editing, and clipboard operations
type CellOperationsState struct {
	CellValuePopupState
	RecordViewState
	ClipboardState
	CellEditState
}

// SQLQueryState manages SQL query input and history
type SQLQueryState struct {
	sqlQueryMode                 bool                      // Whether SQL query input is active
	sqlQueryInput                string                    // Current SQL query being typed
	sqlQueryCursor               int                       // Cursor position in SQL query input
	executedSQLQuery             string                    // The SQL query that was executed (shown in title)
	isCustomQuery                bool                      // Whether current table view is from a custom SQL query
	sqlHistory                   *sqlhistory.History       // SQL query history for current connection
	sqlHistorySuggestions        []sqlhistory.HistoryEntry // Filtered suggestions with timestamps
	sqlHistorySelected           int                       // Selected index in suggestions list (-1 means no selection)
	sqlHistorySuggestionsVisible bool                      // Whether suggestions popup is visible
}

// HistoryState captures the state of a view for navigation history
type HistoryState struct {
	// View mode
	tableViewMode bool

	// Table list state
	selectedRow  int
	scrollOffset int

	// Table view state
	currentViewTable *driver.TableSchema
	tableDataOffset  int
	selectedDataRow  int
	selectedDataCol  int
	tableDataScrollX int
	tableDataScrollY int

	// Custom query state
	isCustomQuery    bool
	executedSQLQuery string
	tableColumns     []string
	tableData        [][]string
}

// NavigationState manages navigation history (back/forward)
type NavigationState struct {
	historyStack        []HistoryState // Stack of previous states
	historyIndex        int            // Current position in history (-1 means no history)
	isNavigatingHistory bool           // Flag to prevent pushing during history navigation
}

// DebugState manages debug panel and logging
type DebugState struct {
	debugMode            bool     // Whether debug overlay is visible
	debugLogs            []string // Debug log messages (ring buffer)
	debugMaxLogs         int      // Maximum number of logs to keep
	debugPanelFocused    bool     // Whether debug panel has keyboard focus
	debugSelectedSection int      // 0=state, 1=logs
	debugSelectedLog     int      // Selected log line index
	debugLogScrollOffset int      // Scroll offset for log display
	debugDetailMode      bool     // Whether detail popup is shown
	debugDetailContent   string   // Content to show in detail popup
	debugDetailTitle     string   // Title of detail popup
	debugStateLines      []string // Cached state lines for navigation
}

// ConnectionManagerState manages saved connections and connection input form
type ConnectionManagerState struct {
	// Connection manager popup
	connManagerMode       bool                          // Whether connection manager popup is active
	connHistory           *connhistory.History          // Connection history
	connManagerFilter     textinput.Model               // Filter input for connections
	connManagerSelected   int                           // Selected connection index
	connManagerScroll     int                           // Scroll offset in connection list
	filteredConnections   []connhistory.ConnectionEntry // Filtered connections
	connManagerInsertMode bool                          // Whether in insert mode (true) or normal/navigate mode (false)

	// New connection input form
	connInputMode  bool              // Whether new connection input form is active
	connInputField int               // Currently focused input field (0=driver, 1=host, 2=port, etc.)
	connInputs     []textinput.Model // Text input fields for connection form
}

// KeySelectorState manages encryption key selection popup
type KeySelectorState struct {
	keySelectorMode       bool               // Whether key selector popup is active
	keySelectorFilter     string             // Filter query for keys
	keySelectorSelected   int                // Selected key index
	keySelectorScroll     int                // Scroll offset in key list
	keySelectorInsertMode bool               // Whether in insert mode
	availableKeys         []encryption.Key   // Discovered encryption keys
	filteredKeys          []encryption.Key   // Filtered keys
	encryptionConfig      *encryption.Config // Current encryption configuration
	encryptionKey         *encryption.Key    // Current encryption key
}

// PassphrasePromptState manages passphrase prompt popup
type PassphrasePromptState struct {
	passphrasePromptMode bool            // Whether passphrase prompt is active
	passphraseInput      textinput.Model // Passphrase input field
	passphraseKeyName    string          // Name of key requiring passphrase
	passphraseKeyPath    string          // Path of key requiring passphrase
	passphraseForNewKey  bool            // Whether this passphrase is for creating a new key
}

// EncryptionState manages encryption keys and passphrase prompts
type EncryptionState struct {
	KeySelectorState
	PassphrasePromptState
}

// HelpPopupState manages help popup state
type HelpPopupState struct {
	helpPopupMode   bool // Whether help popup is visible
	helpPopupScroll int  // Scroll position in help popup
}

// UIState manages UI dimensions, styling, and command mode
type UIState struct {
	width         int
	height        int
	commandMode   bool
	commandBuffer string
	keyBindings   KeyBindings // Configurable key bindings
	styles        Styles      // Color scheme and styling
}

// Model represents the bubbletea application state
type Model struct {
	ConnectionState
	TableListState
	TableDataState
	CellOperationsState
	SQLQueryState
	NavigationState
	DebugState
	ConnectionManagerState
	EncryptionState
	HelpPopupState
	UIState
}

// NewModel creates a new TUI model with database configuration
func NewModel(config DBConfig, keyBindings KeyBindings, styles Styles) Model {
	var drv driver.Driver
	var sqlHist *sqlhistory.History
	var connectionStatus ConnectionStatus

	// Only create driver and SQL history if we have a database configuration
	if config.Driver != "" {
		// Create driver based on config
		driverConfig := &driver.Config{
			Host:     config.Host,
			Port:     config.Port,
			Database: config.Database,
			Schema:   config.Schema,
			User:     config.Username,
			Password: config.Password,
			SSLMode:  "disable", // Default to disable for now
		}

		switch config.Driver {
		case "postgres", "postgresql":
			drv = driver.NewPostgresDriver(driverConfig)
		case "clickhouse", "ch":
			drv = driver.NewClickHouseDriver(driverConfig)
		default:
			// For now, default to postgres
			drv = driver.NewPostgresDriver(driverConfig)
		}

		// Initialize SQL history for this connection
		sqlHist, _ = sqlhistory.NewHistory(
			config.Driver,
			config.Host,
			config.Port,
			config.Database,
			config.Schema,
			config.Username,
			1000, // Max 1000 queries
		)

		connectionStatus = Connecting
	} else {
		// No database configured - will start with connection manager
		connectionStatus = Disconnected
	}

	m := Model{
		ConnectionState: ConnectionState{
			dbConfig:         config,
			driver:           drv,
			connectionStatus: connectionStatus,
		},
		TableListState: TableListState{
			TableListViewState: TableListViewState{
				scrollOffset:   0,
				selectedRow:    0,
				searchMode:     false,
				searchQuery:    "",
				searchSelected: 0,
			},
			SchemaPanelState: SchemaPanelState{},
		},
		TableDataState: TableDataState{
			deletedRows:      make(map[int]bool),
			deletedRowsCount: 0,
			visualMode:       true,
		},
		CellOperationsState: CellOperationsState{
			CellValuePopupState: CellValuePopupState{},
			RecordViewState:     RecordViewState{},
			ClipboardState:      ClipboardState{},
			CellEditState: CellEditState{
				cellEditBuffer:      make(map[string]string),
				cellEditBufferCount: 0,
			},
		},
		SQLQueryState: SQLQueryState{
			sqlHistory:                   sqlHist,
			sqlHistorySelected:           -1,
			sqlHistorySuggestionsVisible: false,
			sqlHistorySuggestions:        make([]sqlhistory.HistoryEntry, 0),
		},
		NavigationState: NavigationState{
			historyStack:        make([]HistoryState, 0),
			historyIndex:        -1, // No history initially
			isNavigatingHistory: false,
		},
		DebugState: DebugState{
			debugMode:            false,
			debugLogs:            make([]string, 0),
			debugMaxLogs:         100, // Keep last 100 debug messages
			debugPanelFocused:    false,
			debugSelectedSection: 1, // Start with logs section
			debugSelectedLog:     0,
			debugLogScrollOffset: 0,
			debugDetailMode:      false,
			debugStateLines:      make([]string, 0),
		},
		ConnectionManagerState: ConnectionManagerState{
			connHistory:           nil,
			connManagerMode:       false,
			connManagerSelected:   0,
			connManagerScroll:     0,
			filteredConnections:   make([]connhistory.ConnectionEntry, 0),
			connManagerInsertMode: true, // Start in insert mode by default
			connInputMode:         false,
			connInputField:        0,
		},
		EncryptionState: EncryptionState{
			KeySelectorState: KeySelectorState{
				keySelectorMode:       false,
				keySelectorFilter:     "",
				keySelectorSelected:   0,
				keySelectorScroll:     0,
				keySelectorInsertMode: true, // Start in insert mode by default
				availableKeys:         make([]encryption.Key, 0),
				filteredKeys:          make([]encryption.Key, 0),
				encryptionConfig:      nil,
				encryptionKey:         nil,
			},
			PassphrasePromptState: PassphrasePromptState{},
		},
		UIState: UIState{
			commandMode:   false,
			commandBuffer: "",
			keyBindings:   keyBindings,
			styles:        styles,
		},
	}

	// Initialize textinput fields after styles are set
	m.connInputs = m.makeConnectionInputs()
	m.passphraseInput = m.makePassphraseInput()
	m.connManagerFilter = m.makeConnectionManagerFilter()

	return m
}

// makeConnectionInputs creates the text input fields for the connection form
func (m Model) makeConnectionInputs() []textinput.Model {
	inputs := []textinput.Model{
		func() textinput.Model {
			input := textinput.New()
			input.Placeholder = "postgres, mariadb, clickhouse"
			input.Focus()
			input.CharLimit = 50
			input.Width = 50
			input.Prompt = ""
			input.TextStyle = m.styles.TextInputStyle
			input.PlaceholderStyle = m.styles.TextInputPlaceholderStyle
			return input
		}(),
		func() textinput.Model {
			input := textinput.New()
			input.Placeholder = "localhost"
			input.CharLimit = 100
			input.Width = 50
			input.Prompt = ""
			input.TextStyle = m.styles.TextInputStyle
			input.PlaceholderStyle = m.styles.TextInputPlaceholderStyle
			return input
		}(),
		func() textinput.Model {
			input := textinput.New()
			input.Placeholder = "5432 (default based on driver)"
			input.CharLimit = 5
			input.Width = 50
			input.Prompt = ""
			input.TextStyle = m.styles.TextInputStyle
			input.PlaceholderStyle = m.styles.TextInputPlaceholderStyle
			return input
		}(),
		func() textinput.Model {
			input := textinput.New()
			input.Placeholder = "postgres (default based on driver)"
			input.CharLimit = 50
			input.Width = 50
			input.Prompt = ""
			input.TextStyle = m.styles.TextInputStyle
			input.PlaceholderStyle = m.styles.TextInputPlaceholderStyle
			return input
		}(),
		func() textinput.Model {
			input := textinput.New()
			input.Placeholder = "password"
			input.EchoMode = textinput.EchoPassword
			input.EchoCharacter = '•'
			input.CharLimit = 100
			input.Width = 50
			input.Prompt = ""
			input.TextStyle = m.styles.TextInputStyle
			input.PlaceholderStyle = m.styles.TextInputPlaceholderStyle
			return input
		}(),
		func() textinput.Model {
			input := textinput.New()
			input.Placeholder = "postgres (default based on driver)"
			input.CharLimit = 50
			input.Width = 50
			input.Prompt = ""
			input.TextStyle = m.styles.TextInputStyle
			input.PlaceholderStyle = m.styles.TextInputPlaceholderStyle
			return input
		}(),
		func() textinput.Model {
			input := textinput.New()
			input.Placeholder = "public (postgres only)"
			input.CharLimit = 50
			input.Width = 50
			input.Prompt = ""
			input.TextStyle = m.styles.TextInputStyle
			input.PlaceholderStyle = m.styles.TextInputPlaceholderStyle
			return input
		}(),
	}

	return inputs
}

// makePassphraseInput creates the text input field for passphrase prompt
func (m Model) makePassphraseInput() textinput.Model {
	input := textinput.New()
	input.Placeholder = "Enter passphrase"
	input.EchoMode = textinput.EchoPassword
	input.EchoCharacter = '•'
	input.CharLimit = 200
	input.Width = 60
	input.Prompt = ""
	input.Focus()
	input.TextStyle = m.styles.TextInputStyle
	input.PlaceholderStyle = m.styles.TextInputPlaceholderStyle
	return input
}

// makeConnectionManagerFilter creates the text input field for connection manager filter
func (m Model) makeConnectionManagerFilter() textinput.Model {
	input := textinput.New()
	input.Placeholder = "Type to filter connections..."
	input.CharLimit = 100
	input.Width = 80
	input.Prompt = ""
	input.Focus()
	input.TextStyle = m.styles.TextInputStyle
	input.PlaceholderStyle = m.styles.TextInputPlaceholderStyle
	return input
}
