package tui

import (
	"github.com/someshkoli/dessertfrog/pkg/connhistory"
	"github.com/someshkoli/dessertfrog/pkg/driver"
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

// JSONNode represents a node in the JSON tree for interactive viewing
type JSONNode struct {
	Key         string      // Key name (for objects) or index (for arrays)
	Value       interface{} // The actual value
	Type        string      // "object", "array", "string", "number", "boolean", "null"
	Expanded    bool        // Whether the node is expanded (for objects/arrays)
	Depth       int         // Nesting depth for indentation
	HasChildren bool        // Whether this node has children
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

// Model represents the bubbletea application state
type Model struct {
	dbConfig               DBConfig
	driver                 driver.Driver
	connectionStatus       ConnectionStatus
	connectionError        string
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
	commandMode            bool
	commandBuffer          string
	width                  int
	height                 int

	// Table data view
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

	// Cell value popup
	cellValuePopupMode     bool       // Whether cell value popup is active
	cellValuePopupContent  string     // The cell value to display
	cellValuePopupIsJSON   bool       // Whether the value is JSON
	cellValuePopupTree     []JSONNode // JSON tree structure
	cellValuePopupScroll   int        // Scroll position in popup
	cellValuePopupSelected int        // Selected node in JSON tree

	// Record view popup (entire row as key-value pairs)
	recordViewMode     bool     // Whether record view popup is active
	recordViewData     []string // The row data (all column values)
	recordViewColumns  []string // The column names
	recordViewSelected int      // Currently selected field index
	recordViewScroll   int      // Scroll position in record view

	// Clipboard notification
	clipboardMessage string // Temporary message shown after copying to clipboard

	// Cell edit mode (multi-cell editing with buffer)
	cellEditMode        bool              // Whether cell edit popup is active
	cellEditValue       string            // The cell value being edited
	cellEditCursor      int               // Cursor position in edit input
	cellEditRowIdx      int               // Row index of cell being edited
	cellEditColIdx      int               // Column index of cell being edited
	cellEditCommandMode bool              // Whether in command mode (:w to save)
	cellEditCommand     string            // Command buffer for :w
	cellEditBuffer      map[string]string // Buffer of pending cell edits: "rowIdx:colIdx" -> newValue
	cellEditBufferCount int               // Number of pending edits

	// SQL query mode
	sqlQueryMode             bool                      // Whether SQL query input is active
	sqlQueryInput            string                    // Current SQL query being typed
	sqlQueryCursor           int                       // Cursor position in SQL query input
	executedSQLQuery         string                    // The SQL query that was executed (shown in title)
	isCustomQuery            bool                      // Whether current table view is from a custom SQL query
	sqlHistory               *sqlhistory.History       // SQL query history for current connection
	sqlHistorySuggestions    []sqlhistory.HistoryEntry // Filtered suggestions with timestamps
	sqlHistorySelected       int                       // Selected index in suggestions list (-1 means no selection)
	sqlHistorySuggestionsVisible bool                  // Whether suggestions popup is visible

	// Navigation history
	historyStack        []HistoryState // Stack of previous states
	historyIndex        int            // Current position in history (-1 means no history)
	isNavigatingHistory bool           // Flag to prevent pushing during history navigation

	// Debug mode
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

	// Key bindings
	keyBindings KeyBindings // Configurable key bindings

	// Styles
	styles Styles // Color scheme and styling

	// Schema panel (right side of home screen)
	schemaPanelFocused   bool                  // Whether schema panel has focus (vs tables list)
	schemaPanelSelected  int                   // Selected index in schema panel
	schemaPanelScroll    int                   // Scroll offset in schema panel
	schemaPanelLineCount int                   // Total number of lines in schema panel
	schemaInfo           *driver.TableSchema   // Cached detailed schema for selected table
	schemaInfoLoading    bool                  // Whether schema info is loading

	// Connection manager popup
	connManagerMode        bool                         // Whether connection manager popup is active
	connHistory            *connhistory.History         // Connection history
	connManagerFilter      string                       // Filter query for connections
	connManagerSelected    int                          // Selected connection index
	connManagerScroll      int                          // Scroll offset in connection list
	filteredConnections    []connhistory.ConnectionEntry // Filtered connections
	connManagerInsertMode  bool                         // Whether in insert mode (true) or normal/navigate mode (false)
}

// NewModel creates a new TUI model with database configuration
func NewModel(config DBConfig, keyBindings KeyBindings, styles Styles) Model {
	// Create driver based on config
	var drv driver.Driver
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
	sqlHist, err := sqlhistory.NewHistory(
		config.Driver,
		config.Host,
		config.Port,
		config.Database,
		config.Schema,
		config.Username,
		1000, // Max 1000 queries
	)
	if err != nil {
		// If history initialization fails, continue without it
		sqlHist = nil
	}

	// Initialize connection history
	connHist, err := connhistory.NewHistory()
	if err != nil {
		// If connection history initialization fails, continue without it
		connHist = nil
	}

	return Model{
		dbConfig:             config,
		driver:               drv,
		connectionStatus:     Connecting,
		scrollOffset:         0,
		selectedRow:          0,
		searchMode:           false,
		searchQuery:          "",
		searchSelected:       0,
		commandMode:          false,
		commandBuffer:        "",
		historyStack:         make([]HistoryState, 0),
		historyIndex:         -1, // No history initially
		isNavigatingHistory:  false,
		debugMode:            false,
		debugLogs:            make([]string, 0),
		debugMaxLogs:         100, // Keep last 100 debug messages
		debugPanelFocused:    false,
		debugSelectedSection: 1, // Start with logs section
		debugSelectedLog:     0,
		debugLogScrollOffset: 0,
		debugDetailMode:      false,
		debugStateLines:      make([]string, 0),
		cellEditBuffer:       make(map[string]string),
		cellEditBufferCount:  0,
		keyBindings:          keyBindings,
		styles:               styles,
		sqlHistory:           sqlHist,
		sqlHistorySelected:   -1,
		sqlHistorySuggestionsVisible: false,
		sqlHistorySuggestions: make([]sqlhistory.HistoryEntry, 0),
		connHistory:           connHist,
		connManagerMode:       false,
		connManagerFilter:     "",
		connManagerSelected:   0,
		connManagerScroll:     0,
		filteredConnections:   make([]connhistory.ConnectionEntry, 0),
		connManagerInsertMode: true, // Start in insert mode by default
	}
}
