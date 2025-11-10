package driver

import (
	"context"
	"time"
)

// EntityType represents the type of database entity
type EntityType string

const (
	EntityTable            EntityType = "table"
	EntityView             EntityType = "view"
	EntityMaterializedView EntityType = "materialized_view"
	EntityFunction         EntityType = "function"
	EntityTrigger          EntityType = "trigger"
)

// ColumnSchema holds detailed information about a table column
type ColumnSchema struct {
	Name         string
	DataType     string // e.g., "varchar", "integer", "timestamp"
	IsNullable   bool
	DefaultValue *string // nil if no default
	MaxLength    *int    // for varchar, char, etc.
	Precision    *int    // for numeric types
	Scale        *int    // for numeric types
	IsPrimaryKey bool
	IsUnique     bool
	IsAutoIncr   bool   // auto increment / serial
	IsForeignKey bool
	ForeignTable string // Referenced table (if foreign key)
	ForeignColumn string // Referenced column (if foreign key)
	Comment      string // column comment if available
}

// IndexInfo holds information about a table index
type IndexInfo struct {
	Name      string
	Columns   []string
	IsUnique  bool
	IsPrimary bool
}

// TableSchema holds detailed schema information about a table or database entity
type TableSchema struct {
	EntityType  EntityType
	SchemaName  string
	TableName   string // Also used for function name, trigger name, etc.
	Columns     []ColumnSchema
	Indexes     []IndexInfo
	RowCount    int64  // -1 if not available/calculated
	TableType   string // "BASE TABLE", "VIEW", "MATERIALIZED VIEW", etc.
	Comment     string // entity comment if available
	CreateTime  *string
	UpdateTime  *string
}

// Driver defines the interface that all database drivers must implement
type Driver interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context) error

	// Close closes the database connection
	Close() error

	// Ping verifies the connection to the database is still alive
	Ping(ctx context.Context) error

	// GetConnectionInfo returns information about the current connection
	GetConnectionInfo() ConnectionInfo

	// GetTables returns a list of tables with basic info (schema name, table name, column count)
	GetTables(ctx context.Context) ([]TableSchema, error)

	// GetAllEntities returns all database entities (tables, views, functions, triggers, etc.)
	GetAllEntities(ctx context.Context) ([]TableSchema, error)

	// GetTableSchema returns detailed schema information for a specific table
	GetTableSchema(ctx context.Context, schemaName, tableName string) (*TableSchema, error)

	// GetTableData fetches the actual data from a table/view
	// Returns column names, rows of data (as strings for display), and timing info
	// limit: maximum number of rows to fetch (e.g., 100)
	// offset: number of rows to skip (for pagination)
	// queryTime: time taken to execute the query
	// fetchTime: time taken to fetch/scan the results
	GetTableData(ctx context.Context, schemaName, tableName string, limit, offset int) (columns []string, rows [][]string, queryTime, fetchTime time.Duration, error error)

	// ExecuteQuery executes a custom SQL query and returns the results
	// Returns column names, rows of data (as strings for display), and timing info
	// queryTime: time taken to execute the query
	// fetchTime: time taken to fetch/scan the results
	ExecuteQuery(ctx context.Context, query string) (columns []string, rows [][]string, queryTime, fetchTime time.Duration, error error)

	// UpdateCell updates a single cell in a table
	// tableSchema: the table schema (includes primary key info, can be nil)
	// columns: all column names for the table
	// oldRow: the original row data (used to identify the row in WHERE clause)
	// columnName: the name of the column to update
	// newValue: the new value to set (as string)
	UpdateCell(ctx context.Context, tableSchema *TableSchema, columns []string, oldRow []string, columnName, newValue string) error
}

// ConnectionInfo holds metadata about a database connection
type ConnectionInfo struct {
	Host     string
	Port     int
	Database string
	User     string
	Driver   string
}

// Config represents common configuration for database drivers
type Config struct {
	Host     string
	Port     int
	Database string
	Schema   string
	User     string
	Password string
	SSLMode  string
}
