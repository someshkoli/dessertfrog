package driver

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// ClickHouseDriver implements the Driver interface for ClickHouse databases
type ClickHouseDriver struct {
	config *Config
	conn   driver.Conn
}

// NewClickHouseDriver creates a new ClickHouse driver instance
func NewClickHouseDriver(config *Config) *ClickHouseDriver {
	return &ClickHouseDriver{
		config: config,
	}
}

// Connect establishes a connection to the ClickHouse database
func (c *ClickHouseDriver) Connect(ctx context.Context) error {
	// Build connection options
	opts := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)},
		Auth: clickhouse.Auth{
			Database: c.config.Database,
			Username: c.config.User,
			Password: c.config.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:      5 * time.Second,
		MaxOpenConns:     10,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Hour,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	}

	// Configure TLS based on SSLMode
	switch c.config.SSLMode {
	case "require":
		opts.TLS = &tls.Config{InsecureSkipVerify: true}
	case "verify-full":
		opts.TLS = &tls.Config{}
	}

	// Create connection
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to create clickhouse connection: %w", err)
	}

	c.conn = conn

	// Verify connection is working
	if err := c.Ping(ctx); err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to ping clickhouse after connection: %w", err)
	}

	return nil
}

// Close closes the ClickHouse connection
func (c *ClickHouseDriver) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Ping verifies the connection to the database is still alive
func (c *ClickHouseDriver) Ping(ctx context.Context) error {
	if c.conn == nil {
		return fmt.Errorf("connection is not established")
	}
	return c.conn.Ping(ctx)
}

// GetConnectionInfo returns information about the current connection
func (c *ClickHouseDriver) GetConnectionInfo() ConnectionInfo {
	return ConnectionInfo{
		Host:     c.config.Host,
		Port:     c.config.Port,
		Database: c.config.Database,
		User:     c.config.User,
		Driver:   "clickhouse",
	}
}

// GetTables returns a list of tables with basic information
func (c *ClickHouseDriver) GetTables(ctx context.Context) ([]TableSchema, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("connection is not established")
	}

	var query string
	var args []interface{}

	if c.config.Database != "" {
		// Query tables in the specific database
		query = `
			SELECT
				database,
				name as table_name,
				engine as table_type,
				total_rows,
				''
			FROM system.tables
			WHERE database = ?
				AND engine NOT IN ('SystemDatabase', 'SystemTables')
			ORDER BY name
		`
		args = append(args, c.config.Database)
	} else {
		// Show all non-system databases
		query = `
			SELECT
				database,
				name as table_name,
				engine as table_type,
				total_rows,
				''
			FROM system.tables
			WHERE database NOT IN ('system', 'INFORMATION_SCHEMA', 'information_schema')
			ORDER BY database, name
		`
	}

	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableSchema

	// Create scan destinations for table rows
	tableColumnTypes := rows.ColumnTypes()
	tableScanDest := createScanDestinations(tableColumnTypes)

	for rows.Next() {
		var schema TableSchema

		if err := rows.Scan(tableScanDest...); err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		// Extract values from scan destinations
		database := fmt.Sprintf("%v", getValueFromPointer(tableScanDest[0]))
		tableName := fmt.Sprintf("%v", getValueFromPointer(tableScanDest[1]))
		tableType := fmt.Sprintf("%v", getValueFromPointer(tableScanDest[2]))
		rowCountRaw := getValueFromPointer(tableScanDest[3])
		comment := fmt.Sprintf("%v", getValueFromPointer(tableScanDest[4]))

		schema.SchemaName = database
		schema.TableName = tableName
		schema.TableType = tableType

		// Convert rowCount using reflection to handle any numeric type
		schema.RowCount = convertRowCountToInt64(rowCountRaw)

		schema.Comment = comment
		schema.EntityType = EntityTable

		// Get column count for this table
		colCountQuery := `
			SELECT COUNT(*)
			FROM system.columns
			WHERE database = ? AND table = ?
		`
		var columnCount uint64
		err := c.conn.QueryRow(ctx, colCountQuery, database, tableName).Scan(&columnCount)
		if err != nil {
			// If column count query fails, set to 0 but don't fail the whole operation
			schema.Columns = make([]ColumnSchema, 0)
		} else {
			schema.Columns = make([]ColumnSchema, columnCount)
		}

		tables = append(tables, schema)

		// Reset scan destinations for next row
		tableScanDest = createScanDestinations(tableColumnTypes)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	return tables, nil
}

// GetTableSchema returns detailed schema information for a specific table
func (c *ClickHouseDriver) GetTableSchema(ctx context.Context, schemaName, tableName string) (*TableSchema, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("connection is not established")
	}

	// Get table information
	schema := &TableSchema{
		SchemaName: schemaName,
		TableName:  tableName,
		EntityType: EntityTable,
	}

	// Get table metadata
	tableQuery := `
		SELECT
			engine,
			total_rows
		FROM system.tables
		WHERE database = ? AND name = ?
	`
	var rowCount uint64
	err := c.conn.QueryRow(ctx, tableQuery, schemaName, tableName).Scan(&schema.TableType, &rowCount)
	if err != nil {
		return nil, fmt.Errorf("failed to query table info: %w", err)
	}
	schema.RowCount = int64(rowCount)

	// Get column information
	columnQuery := `
		SELECT
			name,
			type,
			default_kind,
			default_expression,
			comment,
			is_in_primary_key
		FROM system.columns
		WHERE database = ? AND table = ?
		ORDER BY position
	`

	rows, err := c.conn.Query(ctx, columnQuery, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnSchema
		var defaultKind, defaultExpr, comment string
		var isPrimaryKey uint8

		err := rows.Scan(
			&col.Name,
			&col.DataType,
			&defaultKind,
			&defaultExpr,
			&comment,
			&isPrimaryKey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column row: %w", err)
		}

		// Parse ClickHouse type to determine nullability
		col.IsNullable = strings.HasPrefix(col.DataType, "Nullable(")

		// Set default value if present
		if defaultExpr != "" {
			col.DefaultValue = &defaultExpr
		}

		// Set primary key flag
		col.IsPrimaryKey = isPrimaryKey == 1

		// ClickHouse doesn't have traditional auto increment, but has special defaults
		col.IsAutoIncr = defaultKind == "DEFAULT" && strings.Contains(strings.ToLower(defaultExpr), "uuid")

		col.Comment = comment

		schema.Columns = append(schema.Columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating column rows: %w", err)
	}

	// Get primary key information
	pkQuery := `
		SELECT primary_key
		FROM system.tables
		WHERE database = ? AND name = ?
	`
	var primaryKey string
	if err := c.conn.QueryRow(ctx, pkQuery, schemaName, tableName).Scan(&primaryKey); err == nil && primaryKey != "" {
		// Parse primary key columns
		pkCols := parsePrimaryKeyColumns(primaryKey)

		idx := IndexInfo{
			Name:      "PRIMARY",
			Columns:   pkCols,
			IsUnique:  true,
			IsPrimary: true,
		}
		schema.Indexes = append(schema.Indexes, idx)
	}

	return schema, nil
}

// parsePrimaryKeyColumns parses the primary key definition from ClickHouse
func parsePrimaryKeyColumns(pkDef string) []string {
	// Remove parentheses and split by comma
	pkDef = strings.TrimSpace(pkDef)
	pkDef = strings.TrimPrefix(pkDef, "(")
	pkDef = strings.TrimSuffix(pkDef, ")")

	if pkDef == "" {
		return nil
	}

	cols := strings.Split(pkDef, ",")
	result := make([]string, 0, len(cols))
	for _, col := range cols {
		col = strings.TrimSpace(col)
		// Remove any function calls or expressions
		if idx := strings.Index(col, "("); idx > 0 {
			col = col[:idx]
		}
		if col != "" {
			result = append(result, col)
		}
	}
	return result
}

// GetAllEntities returns all database entities (tables and views)
func (c *ClickHouseDriver) GetAllEntities(ctx context.Context) ([]TableSchema, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("connection is not established")
	}

	var allEntities []TableSchema

	// Fetch tables
	tables, err := c.fetchTablesAndViews(ctx)
	if err != nil {
		return nil, err
	}
	allEntities = append(allEntities, tables...)

	return allEntities, nil
}

// fetchTablesAndViews fetches tables and views
func (c *ClickHouseDriver) fetchTablesAndViews(ctx context.Context) ([]TableSchema, error) {
	var query string
	var args []interface{}

	if c.config.Database != "" {
		query = `
			SELECT
				database,
				name as table_name,
				engine as table_type,
				total_rows
			FROM system.tables
			WHERE database = ?
				AND engine NOT IN ('SystemDatabase', 'SystemTables')
			ORDER BY name
		`
		args = append(args, c.config.Database)
	} else {
		query = `
			SELECT
				database,
				name as table_name,
				engine as table_type,
				total_rows
			FROM system.tables
			WHERE database NOT IN ('system', 'INFORMATION_SCHEMA', 'information_schema')
			ORDER BY database, name
		`
	}

	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables and views: %w", err)
	}
	defer rows.Close()

	var entities []TableSchema

	// Create scan destinations
	columnTypes := rows.ColumnTypes()
	scanDest := createScanDestinations(columnTypes)

	for rows.Next() {
		var entity TableSchema

		if err := rows.Scan(scanDest...); err != nil {
			return nil, fmt.Errorf("failed to scan table/view row: %w", err)
		}

		// Extract values from scan destinations
		database := fmt.Sprintf("%v", getValueFromPointer(scanDest[0]))
		tableName := fmt.Sprintf("%v", getValueFromPointer(scanDest[1]))
		tableType := fmt.Sprintf("%v", getValueFromPointer(scanDest[2]))
		rowCountRaw := getValueFromPointer(scanDest[3])

		entity.SchemaName = database
		entity.TableName = tableName
		entity.TableType = tableType

		// Convert rowCount using reflection to handle any numeric type
		entity.RowCount = convertRowCountToInt64(rowCountRaw)

		// Determine entity type based on engine
		if strings.Contains(tableType, "View") {
			entity.EntityType = EntityView
		} else if strings.Contains(tableType, "MaterializedView") {
			entity.EntityType = EntityMaterializedView
		} else {
			entity.EntityType = EntityTable
		}

		// Get column count
		colCountQuery := `
			SELECT COUNT(*)
			FROM system.columns
			WHERE database = ? AND table = ?
		`
		var columnCount uint64
		err := c.conn.QueryRow(ctx, colCountQuery, database, tableName).Scan(&columnCount)
		if err != nil {
			// If column count query fails, set to 0 but don't fail the whole operation
			entity.Columns = make([]ColumnSchema, 0)
		} else {
			entity.Columns = make([]ColumnSchema, columnCount)
		}

		entities = append(entities, entity)

		// Reset scan destinations for next row
		scanDest = createScanDestinations(columnTypes)
	}

	return entities, rows.Err()
}

// GetTableData fetches actual data from a table or view
func (c *ClickHouseDriver) GetTableData(ctx context.Context, schemaName, tableName string, limit, offset int) ([]string, [][]string, time.Duration, time.Duration, error) {
	if c.conn == nil {
		return nil, nil, 0, 0, fmt.Errorf("connection is not established")
	}

	// Build the query with LIMIT and OFFSET
	query := fmt.Sprintf("SELECT * FROM `%s`.`%s` LIMIT %d OFFSET %d", schemaName, tableName, limit, offset)

	// Execute query and measure query time
	queryStart := time.Now()
	rows, err := c.conn.Query(ctx, query)
	queryTime := time.Since(queryStart)

	if err != nil {
		return nil, nil, queryTime, 0, fmt.Errorf("failed to query table data: %w", err)
	}
	defer rows.Close()

	// Get column names
	columnTypes := rows.ColumnTypes()
	columns := make([]string, len(columnTypes))
	for i, ct := range columnTypes {
		columns[i] = ct.Name()
	}

	// Fetch all rows and measure fetch time
	fetchStart := time.Now()
	var data [][]string

	// Create typed scan targets based on column types
	scanDest := createScanDestinations(columnTypes)

	for rows.Next() {
		if err := rows.Scan(scanDest...); err != nil {
			fetchTime := time.Since(fetchStart)
			return nil, nil, queryTime, fetchTime, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert scanned values to strings
		row := make([]string, len(scanDest))
		for i, dest := range scanDest {
			row[i] = formatClickHouseValue(getValueFromPointer(dest))
		}

		data = append(data, row)

		// Reset scan destinations for next row
		scanDest = createScanDestinations(columnTypes)
	}
	fetchTime := time.Since(fetchStart)

	if err := rows.Err(); err != nil {
		return nil, nil, queryTime, fetchTime, fmt.Errorf("error iterating rows: %w", err)
	}

	return columns, data, queryTime, fetchTime, nil
}

// ExecuteQuery executes a custom SQL query and returns the results
func (c *ClickHouseDriver) ExecuteQuery(ctx context.Context, query string) ([]string, [][]string, time.Duration, time.Duration, error) {
	if c.conn == nil {
		return nil, nil, 0, 0, fmt.Errorf("connection is not established")
	}

	// Execute the query and measure query time
	queryStart := time.Now()
	rows, err := c.conn.Query(ctx, query)
	queryTime := time.Since(queryStart)

	if err != nil {
		return nil, nil, queryTime, 0, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columnTypes := rows.ColumnTypes()
	columns := make([]string, len(columnTypes))
	for i, ct := range columnTypes {
		columns[i] = ct.Name()
	}

	// Fetch all rows and measure fetch time
	fetchStart := time.Now()
	var data [][]string

	// Create typed scan targets based on column types
	scanDest := createScanDestinations(columnTypes)

	for rows.Next() {
		if err := rows.Scan(scanDest...); err != nil {
			fetchTime := time.Since(fetchStart)
			return nil, nil, queryTime, fetchTime, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert scanned values to strings
		row := make([]string, len(scanDest))
		for i, dest := range scanDest {
			row[i] = formatClickHouseValue(getValueFromPointer(dest))
		}

		data = append(data, row)

		// Reset scan destinations for next row
		scanDest = createScanDestinations(columnTypes)
	}
	fetchTime := time.Since(fetchStart)

	if err := rows.Err(); err != nil {
		return nil, nil, queryTime, fetchTime, fmt.Errorf("error iterating rows: %w", err)
	}

	return columns, data, queryTime, fetchTime, nil
}

// formatClickHouseValue converts a ClickHouse value to a string for display
func formatClickHouseValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}

	switch v := val.(type) {
	case []interface{}, map[string]interface{}:
		// Handle arrays and maps (JSON-like structures)
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(jsonBytes)
	case []uint8:
		// Byte arrays - convert to string
		return string(v)
	case time.Time:
		// Format time consistently
		return v.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%v", val)
	}
}

// UpdateCell updates a single cell in a table
func (c *ClickHouseDriver) UpdateCell(ctx context.Context, tableSchema *TableSchema, columns []string, oldRow []string, columnName, newValue string) error {
	if c.conn == nil {
		return fmt.Errorf("connection is not established")
	}

	// ClickHouse uses ALTER TABLE for mutations
	// This is different from traditional UPDATE statements

	// Get schema and table names
	var schemaName, tableName string
	if tableSchema != nil {
		schemaName = tableSchema.SchemaName
		tableName = tableSchema.TableName
	} else {
		return fmt.Errorf("table schema is required for update")
	}

	// Build WHERE clause to identify the row
	var whereParts []string

	// Determine which columns to use in WHERE clause
	// Priority: 1) Primary keys (if available), 2) All columns
	var whereColumns []string

	// Check if we have table schema with primary key information
	if tableSchema != nil && len(tableSchema.Columns) > 0 {
		// Find primary key columns
		for _, col := range tableSchema.Columns {
			if col.IsPrimaryKey {
				whereColumns = append(whereColumns, col.Name)
			}
		}
	}

	// If no primary keys found, use all columns to identify the row
	if len(whereColumns) == 0 {
		whereColumns = columns
	}

	// Build WHERE clause using selected columns
	for _, col := range whereColumns {
		// Find the index of this column in the columns array
		colIdx := -1
		for i, c := range columns {
			if c == col {
				colIdx = i
				break
			}
		}

		if colIdx >= 0 && colIdx < len(oldRow) {
			if oldRow[colIdx] == "NULL" {
				whereParts = append(whereParts, fmt.Sprintf("`%s` IS NULL", col))
			} else {
				// Escape single quotes in values
				escapedValue := strings.ReplaceAll(oldRow[colIdx], "'", "''")
				whereParts = append(whereParts, fmt.Sprintf("`%s` = '%s'", col, escapedValue))
			}
		}
	}

	whereClause := strings.Join(whereParts, " AND ")

	// Escape single quotes in the new value
	escapedNewValue := strings.ReplaceAll(newValue, "'", "''")

	// Build ALTER TABLE UPDATE query
	query := fmt.Sprintf(
		"ALTER TABLE `%s`.`%s` UPDATE `%s` = '%s' WHERE %s",
		schemaName,
		tableName,
		columnName,
		escapedNewValue,
		whereClause,
	)

	// Execute the UPDATE
	err := c.conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to update cell: %w\nQuery:\n%s", err, query)
	}

	return nil
}

// DeleteRows deletes multiple rows from a table in a batch operation
func (c *ClickHouseDriver) DeleteRows(ctx context.Context, tableSchema *TableSchema, columns []string, rows [][]string) (int64, error) {
	if c.conn == nil {
		return 0, fmt.Errorf("connection is not established")
	}

	if tableSchema == nil {
		return 0, fmt.Errorf("table schema is required for delete")
	}

	if len(rows) == 0 {
		return 0, nil // Nothing to delete
	}

	// Determine which columns to use in WHERE clause
	// Priority: 1) Primary keys (if available), 2) All columns
	var whereColumns []string

	// Check if we have table schema with primary key information
	if len(tableSchema.Columns) > 0 {
		// Find primary key columns
		for _, col := range tableSchema.Columns {
			if col.IsPrimaryKey {
				whereColumns = append(whereColumns, col.Name)
			}
		}
	}

	// If no primary keys found, use all columns to identify the row
	if len(whereColumns) == 0 {
		whereColumns = columns
	}

	// Build DELETE query with multiple row conditions combined with OR
	// Format: WHERE (col1 = val1 AND col2 = val2) OR (col1 = val3 AND col2 = val4) OR ...
	var whereClauses []string

	for _, row := range rows {
		var rowWhereParts []string

		// Build WHERE clause for this row (conditions combined with AND)
		for _, col := range whereColumns {
			// Find the index of this column in the columns array
			colIdx := -1
			for i, c := range columns {
				if c == col {
					colIdx = i
					break
				}
			}

			if colIdx >= 0 && colIdx < len(row) {
				if row[colIdx] == "NULL" {
					rowWhereParts = append(rowWhereParts, fmt.Sprintf("`%s` IS NULL", col))
				} else {
					// Escape single quotes in values
					escapedValue := strings.ReplaceAll(row[colIdx], "'", "''")
					rowWhereParts = append(rowWhereParts, fmt.Sprintf("`%s` = '%s'", col, escapedValue))
				}
			}
		}

		if len(rowWhereParts) > 0 {
			// Wrap each row's conditions in parentheses and combine with AND
			whereClauses = append(whereClauses, "("+strings.Join(rowWhereParts, " AND ")+")")
		}
	}

	if len(whereClauses) == 0 {
		return 0, fmt.Errorf("no valid rows to delete")
	}

	// Combine all row WHERE clauses with OR
	whereClause := strings.Join(whereClauses, " OR ")

	// Build ALTER TABLE DELETE query (ClickHouse uses ALTER TABLE for deletes)
	query := fmt.Sprintf(
		"ALTER TABLE `%s`.`%s` DELETE WHERE %s",
		tableSchema.SchemaName,
		tableSchema.TableName,
		whereClause,
	)

	// Execute the DELETE
	err := c.conn.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete rows: %w\nQuery:\n%s", err, query)
	}

	// ClickHouse ALTER TABLE DELETE is asynchronous and doesn't return affected rows count
	// Return the number of rows we attempted to delete
	return int64(len(rows)), nil
}

// Helper functions for ClickHouse data scanning

// convertRowCountToInt64 converts any numeric type to int64 using reflection
func convertRowCountToInt64(val interface{}) int64 {
	if val == nil {
		return 0
	}

	v := reflect.ValueOf(val)

	// If it's a pointer, dereference it first
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(v.Uint())
	case reflect.Float32, reflect.Float64:
		return int64(v.Float())
	case reflect.String:
		var parsed int64
		fmt.Sscanf(v.String(), "%d", &parsed)
		return parsed
	default:
		// Try to parse the string representation as last resort
		var parsed int64
		fmt.Sscanf(fmt.Sprintf("%v", val), "%d", &parsed)
		return parsed
	}
}

// createScanDestinations creates properly typed scan destinations for ClickHouse
func createScanDestinations(columnTypes []driver.ColumnType) []interface{} {
	scanDest := make([]interface{}, len(columnTypes))
	for i, ct := range columnTypes {
		// Use the ScanType provided by ClickHouse driver to create appropriate destination
		scanType := ct.ScanType()
		if scanType != nil {
			// Create a pointer to a new instance of the scan type
			scanDest[i] = reflect.New(scanType).Interface()
		} else {
			// Fallback to interface{} if scan type is unknown
			var v interface{}
			scanDest[i] = &v
		}
	}
	return scanDest
}

// getValueFromPointer extracts the actual value from a pointer using reflection
func getValueFromPointer(ptr interface{}) interface{} {
	if ptr == nil {
		return nil
	}

	v := reflect.ValueOf(ptr)

	// If it's a pointer, dereference it
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		return v.Elem().Interface()
	}

	// If it's not a pointer, return as is
	return ptr
}
