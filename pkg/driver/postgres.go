package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDriver implements the Driver interface for PostgreSQL databases
type PostgresDriver struct {
	config *Config
	pool   *pgxpool.Pool
}

// NewPostgresDriver creates a new PostgreSQL driver instance
func NewPostgresDriver(config *Config) *PostgresDriver {
	return &PostgresDriver{
		config: config,
	}
}

// Connect establishes a connection pool to the PostgreSQL database
func (p *PostgresDriver) Connect(ctx context.Context) error {
	// Build connection string
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=10",
		p.config.Host,
		p.config.Port,
		p.config.User,
		p.config.Password,
		p.config.Database,
		p.config.SSLMode,
	)

	// Create connection pool
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return fmt.Errorf("failed to create postgres connection pool: %w", err)
	}

	p.pool = pool

	// Verify connection is working
	if err := p.Ping(ctx); err != nil {
		p.pool.Close()
		return fmt.Errorf("failed to ping postgres after connection: %w", err)
	}

	return nil
}

// Close closes the PostgreSQL connection pool
func (p *PostgresDriver) Close() error {
	if p.pool != nil {
		p.pool.Close()
	}
	return nil
}

// Ping verifies the connection to the database is still alive
func (p *PostgresDriver) Ping(ctx context.Context) error {
	if p.pool == nil {
		return fmt.Errorf("connection pool is not established")
	}
	return p.pool.Ping(ctx)
}

// GetConnectionInfo returns information about the current connection
func (p *PostgresDriver) GetConnectionInfo() ConnectionInfo {
	return ConnectionInfo{
		Host:     p.config.Host,
		Port:     p.config.Port,
		Database: p.config.Database,
		User:     p.config.User,
		Driver:   "postgres",
	}
}

// GetTables returns a list of tables with basic information
func (p *PostgresDriver) GetTables(ctx context.Context) ([]TableSchema, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("connection pool is not established")
	}

	var query string
	var rows pgx.Rows
	var err error

	if p.config.Schema != "" {
		// Filter by specific schema
		query = `
			SELECT
				t.table_schema,
				t.table_name,
				t.table_type,
				COUNT(c.column_name) as column_count,
				obj_description((t.table_schema || '.' || t.table_name)::regclass, 'pg_class') as table_comment,
				COALESCE(pgc.reltuples, 0)::bigint as row_count
			FROM information_schema.tables t
			LEFT JOIN information_schema.columns c
				ON t.table_schema = c.table_schema
				AND t.table_name = c.table_name
			LEFT JOIN pg_class pgc
				ON pgc.relname = t.table_name
				AND pgc.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = t.table_schema)
			WHERE t.table_schema = $1
			GROUP BY t.table_schema, t.table_name, t.table_type, pgc.reltuples
			ORDER BY t.table_name
		`
		rows, err = p.pool.Query(ctx, query, p.config.Schema)
	} else {
		// Show all non-system schemas
		query = `
			SELECT
				t.table_schema,
				t.table_name,
				t.table_type,
				COUNT(c.column_name) as column_count,
				obj_description((t.table_schema || '.' || t.table_name)::regclass, 'pg_class') as table_comment,
				COALESCE(pgc.reltuples, 0)::bigint as row_count
			FROM information_schema.tables t
			LEFT JOIN information_schema.columns c
				ON t.table_schema = c.table_schema
				AND t.table_name = c.table_name
			LEFT JOIN pg_class pgc
				ON pgc.relname = t.table_name
				AND pgc.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = t.table_schema)
			WHERE t.table_schema NOT IN ('pg_catalog', 'information_schema')
			GROUP BY t.table_schema, t.table_name, t.table_type, pgc.reltuples
			ORDER BY t.table_schema, t.table_name
		`
		rows, err = p.pool.Query(ctx, query)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableSchema
	for rows.Next() {
		var schema TableSchema
		var comment *string
		var columnCount int64

		err := rows.Scan(
			&schema.SchemaName,
			&schema.TableName,
			&schema.TableType,
			&columnCount,
			&comment,
			&schema.RowCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		if comment != nil {
			schema.Comment = *comment
		}

		// Set entity type to table
		schema.EntityType = EntityTable

		// Store column count in Columns slice with empty entries to represent count
		schema.Columns = make([]ColumnSchema, columnCount)

		tables = append(tables, schema)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	return tables, nil
}

// GetTableSchema returns detailed schema information for a specific table
func (p *PostgresDriver) GetTableSchema(ctx context.Context, schemaName, tableName string) (*TableSchema, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("connection is not established")
	}

	// First, get table info
	var schema TableSchema
	schema.SchemaName = schemaName
	schema.TableName = tableName
	schema.RowCount = -1

	tableQuery := `
		SELECT
			table_type,
			obj_description((table_schema || '.' || table_name)::regclass, 'pg_class') as table_comment
		FROM information_schema.tables
		WHERE table_schema = $1 AND table_name = $2
	`

	var comment *string
	err := p.pool.QueryRow(ctx, tableQuery, schemaName, tableName).Scan(&schema.TableType, &comment)
	if err != nil {
		return nil, fmt.Errorf("failed to query table info: %w", err)
	}

	if comment != nil {
		schema.Comment = *comment
	}

	// Now get column information with foreign keys
	columnQuery := `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			c.numeric_precision,
			c.numeric_scale,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary_key,
			CASE WHEN c.column_name IN (
				SELECT kcu.column_name
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu
					ON tc.constraint_name = kcu.constraint_name
					AND tc.table_schema = kcu.table_schema
				WHERE tc.table_schema = c.table_schema
					AND tc.table_name = c.table_name
					AND tc.constraint_type = 'UNIQUE'
			) THEN true ELSE false END as is_unique,
			CASE WHEN c.column_default LIKE 'nextval%' THEN true ELSE false END as is_auto_increment,
			col_description((c.table_schema || '.' || c.table_name)::regclass, c.ordinal_position) as column_comment,
			CASE WHEN fk.column_name IS NOT NULL THEN true ELSE false END as is_foreign_key,
			COALESCE(fk.foreign_table, '') as foreign_table,
			COALESCE(fk.foreign_column, '') as foreign_column
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT kcu.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			WHERE tc.constraint_type = 'PRIMARY KEY'
				AND tc.table_schema = $1
				AND tc.table_name = $2
		) pk ON c.column_name = pk.column_name
		LEFT JOIN (
			SELECT
				kcu.column_name,
				ccu.table_name AS foreign_table,
				ccu.column_name AS foreign_column
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage ccu
				ON tc.constraint_name = ccu.constraint_name
				AND tc.table_schema = ccu.table_schema
			WHERE tc.constraint_type = 'FOREIGN KEY'
				AND tc.table_schema = $1
				AND tc.table_name = $2
		) fk ON c.column_name = fk.column_name
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position
	`

	rows, err := p.pool.Query(ctx, columnQuery, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnSchema
		var isNullable string
		var defaultVal, columnComment *string
		var maxLength, precision, scale *int

		err := rows.Scan(
			&col.Name,
			&col.DataType,
			&isNullable,
			&defaultVal,
			&maxLength,
			&precision,
			&scale,
			&col.IsPrimaryKey,
			&col.IsUnique,
			&col.IsAutoIncr,
			&columnComment,
			&col.IsForeignKey,
			&col.ForeignTable,
			&col.ForeignColumn,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column row: %w", err)
		}

		col.IsNullable = (isNullable == "YES")
		col.DefaultValue = defaultVal
		col.MaxLength = maxLength
		col.Precision = precision
		col.Scale = scale

		if columnComment != nil {
			col.Comment = *columnComment
		}

		schema.Columns = append(schema.Columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating column rows: %w", err)
	}

	// Fetch indexes for this table
	indexQuery := `
		SELECT
			i.indexname,
			i.indexdef,
			CASE WHEN i.indexname LIKE '%_pkey' THEN true ELSE false END as is_primary
		FROM pg_indexes i
		WHERE i.schemaname = $1 AND i.tablename = $2
		ORDER BY i.indexname
	`

	idxRows, err := p.pool.Query(ctx, indexQuery, schemaName, tableName)
	if err == nil {
		defer idxRows.Close()
		for idxRows.Next() {
			var idx IndexInfo
			var indexDef string
			var isPrimary bool

			err := idxRows.Scan(&idx.Name, &indexDef, &isPrimary)
			if err != nil {
				continue
			}

			idx.IsPrimary = isPrimary
			idx.IsUnique = strings.Contains(strings.ToUpper(indexDef), "UNIQUE")

			// Parse column names from index definition (simplified)
			// Example: CREATE INDEX idx_name ON table USING btree (col1, col2)
			if strings.Contains(indexDef, "(") && strings.Contains(indexDef, ")") {
				start := strings.Index(indexDef, "(")
				end := strings.LastIndex(indexDef, ")")
				colStr := indexDef[start+1 : end]
				cols := strings.Split(colStr, ",")
				for _, col := range cols {
					col = strings.TrimSpace(col)
					// Remove any function calls or expressions, keep just column name
					if spaceIdx := strings.Index(col, " "); spaceIdx > 0 {
						col = col[:spaceIdx]
					}
					idx.Columns = append(idx.Columns, col)
				}
			}

			schema.Indexes = append(schema.Indexes, idx)
		}
	}

	return &schema, nil
}

// GetAllEntities returns all database entities (tables, views, functions, triggers)
func (p *PostgresDriver) GetAllEntities(ctx context.Context) ([]TableSchema, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("connection is not established")
	}

	var allEntities []TableSchema

	// Fetch tables and views (including materialized views)
	tablesAndViews, err := p.fetchTablesAndViews(ctx)
	if err != nil {
		return nil, err
	}
	allEntities = append(allEntities, tablesAndViews...)

	// Fetch functions
	functions, err := p.fetchFunctions(ctx)
	if err != nil {
		return nil, err
	}
	allEntities = append(allEntities, functions...)

	// Fetch triggers
	triggers, err := p.fetchTriggers(ctx)
	if err != nil {
		return nil, err
	}
	allEntities = append(allEntities, triggers...)

	return allEntities, nil
}

// fetchTablesAndViews fetches tables and views
func (p *PostgresDriver) fetchTablesAndViews(ctx context.Context) ([]TableSchema, error) {
	var query string
	var rows pgx.Rows
	var err error

	if p.config.Schema != "" {
		query = `
			SELECT
				t.table_schema,
				t.table_name,
				t.table_type,
				COUNT(c.column_name) as column_count
			FROM information_schema.tables t
			LEFT JOIN information_schema.columns c
				ON t.table_schema = c.table_schema
				AND t.table_name = c.table_name
			WHERE t.table_schema = $1
				AND t.table_type IN ('BASE TABLE', 'VIEW', 'MATERIALIZED VIEW')
			GROUP BY t.table_schema, t.table_name, t.table_type
			ORDER BY t.table_name
		`
		rows, err = p.pool.Query(ctx, query, p.config.Schema)
	} else {
		query = `
			SELECT
				t.table_schema,
				t.table_name,
				t.table_type,
				COUNT(c.column_name) as column_count
			FROM information_schema.tables t
			LEFT JOIN information_schema.columns c
				ON t.table_schema = c.table_schema
				AND t.table_name = c.table_name
			WHERE t.table_schema NOT IN ('pg_catalog', 'information_schema')
				AND t.table_type IN ('BASE TABLE', 'VIEW', 'MATERIALIZED VIEW')
			GROUP BY t.table_schema, t.table_name, t.table_type
			ORDER BY t.table_schema, t.table_name
		`
		rows, err = p.pool.Query(ctx, query)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query tables and views: %w", err)
	}
	defer rows.Close()

	var entities []TableSchema
	for rows.Next() {
		var entity TableSchema
		var columnCount int64

		err := rows.Scan(
			&entity.SchemaName,
			&entity.TableName,
			&entity.TableType,
			&columnCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table/view row: %w", err)
		}

		// Set entity type based on table_type
		switch entity.TableType {
		case "BASE TABLE":
			entity.EntityType = EntityTable
		case "VIEW":
			entity.EntityType = EntityView
		case "MATERIALIZED VIEW":
			entity.EntityType = EntityMaterializedView
		}

		entity.RowCount = -1
		entity.Columns = make([]ColumnSchema, columnCount)

		entities = append(entities, entity)
	}

	return entities, rows.Err()
}

// fetchFunctions fetches user-defined functions
func (p *PostgresDriver) fetchFunctions(ctx context.Context) ([]TableSchema, error) {
	var query string
	var rows pgx.Rows
	var err error

	if p.config.Schema != "" {
		query = `
			SELECT
				n.nspname as schema_name,
				p.proname as function_name,
				pg_get_function_arguments(p.oid) as arguments
			FROM pg_proc p
			JOIN pg_namespace n ON p.pronamespace = n.oid
			WHERE n.nspname = $1
				AND p.prokind = 'f'
			ORDER BY p.proname
		`
		rows, err = p.pool.Query(ctx, query, p.config.Schema)
	} else {
		query = `
			SELECT
				n.nspname as schema_name,
				p.proname as function_name,
				pg_get_function_arguments(p.oid) as arguments
			FROM pg_proc p
			JOIN pg_namespace n ON p.pronamespace = n.oid
			WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
				AND p.prokind = 'f'
			ORDER BY n.nspname, p.proname
		`
		rows, err = p.pool.Query(ctx, query)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query functions: %w", err)
	}
	defer rows.Close()

	var entities []TableSchema
	for rows.Next() {
		var entity TableSchema
		var arguments string

		err := rows.Scan(
			&entity.SchemaName,
			&entity.TableName,
			&arguments,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan function row: %w", err)
		}

		entity.EntityType = EntityFunction
		entity.TableType = "FUNCTION"
		entity.Comment = arguments
		entity.RowCount = -1

		entities = append(entities, entity)
	}

	return entities, rows.Err()
}

// fetchTriggers fetches triggers
func (p *PostgresDriver) fetchTriggers(ctx context.Context) ([]TableSchema, error) {
	var query string
	var rows pgx.Rows
	var err error

	if p.config.Schema != "" {
		query = `
			SELECT
				n.nspname as schemaname,
				t.tgname as trigname,
				c.relname as tablename
			FROM pg_catalog.pg_trigger t
			JOIN pg_catalog.pg_class c ON t.tgrelid = c.oid
			JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
			WHERE n.nspname = $1
				AND NOT t.tgisinternal
			ORDER BY t.tgname
		`
		rows, err = p.pool.Query(ctx, query, p.config.Schema)
	} else {
		query = `
			SELECT
				n.nspname as schemaname,
				t.tgname as trigname,
				c.relname as tablename
			FROM pg_catalog.pg_trigger t
			JOIN pg_catalog.pg_class c ON t.tgrelid = c.oid
			JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
			WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
				AND NOT t.tgisinternal
			ORDER BY n.nspname, t.tgname
		`
		rows, err = p.pool.Query(ctx, query)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query triggers: %w", err)
	}
	defer rows.Close()

	var entities []TableSchema
	for rows.Next() {
		var entity TableSchema
		var tableName string

		err := rows.Scan(
			&entity.SchemaName,
			&entity.TableName,
			&tableName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trigger row: %w", err)
		}

		entity.EntityType = EntityTrigger
		entity.TableType = "TRIGGER"
		entity.Comment = "On table: " + tableName
		entity.RowCount = -1

		entities = append(entities, entity)
	}

	return entities, rows.Err()
}

// GetTableData fetches actual data from a table or view
func (p *PostgresDriver) GetTableData(ctx context.Context, schemaName, tableName string, limit, offset int) ([]string, [][]string, time.Duration, time.Duration, error) {
	if p.pool == nil {
		return nil, nil, 0, 0, fmt.Errorf("connection is not established")
	}

	// Build the query with LIMIT and OFFSET
	query := fmt.Sprintf(`SELECT * FROM "%s"."%s" LIMIT $1 OFFSET $2`, schemaName, tableName)

	// Execute query and measure query time
	queryStart := time.Now()
	rows, err := p.pool.Query(ctx, query, limit, offset)
	queryTime := time.Since(queryStart)

	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("failed to query table data: %w", err)
	}
	defer rows.Close()

	// Get column names
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	// Fetch all rows and measure fetch time
	fetchStart := time.Now()
	var data [][]string
	for rows.Next() {
		// Get values
		values, err := rows.Values()
		if err != nil {
			fetchTime := time.Since(fetchStart)
			return nil, nil, queryTime, fetchTime, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert values to strings
		row := make([]string, len(values))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
				// Check if value is a map or slice (likely JSON/JSONB)
				switch v := val.(type) {
				case map[string]interface{}, []interface{}:
					// Marshal back to JSON string
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						row[i] = fmt.Sprintf("%v", val)
					} else {
						row[i] = string(jsonBytes)
					}
				default:
					row[i] = fmt.Sprintf("%v", val)
				}
			}
		}

		data = append(data, row)
	}
	fetchTime := time.Since(fetchStart)

	if err := rows.Err(); err != nil {
		return nil, nil, queryTime, fetchTime, fmt.Errorf("error iterating rows: %w", err)
	}

	return columns, data, queryTime, fetchTime, nil
}

// ExecuteQuery executes a custom SQL query and returns the results
func (p *PostgresDriver) ExecuteQuery(ctx context.Context, query string) ([]string, [][]string, time.Duration, time.Duration, error) {
	if p.pool == nil {
		return nil, nil, 0, 0, fmt.Errorf("connection is not established")
	}

	// Execute the query and measure query time
	queryStart := time.Now()
	rows, err := p.pool.Query(ctx, query)
	queryTime := time.Since(queryStart)

	if err != nil {
		return nil, nil, queryTime, 0, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	// Fetch all rows and measure fetch time
	fetchStart := time.Now()
	var data [][]string
	for rows.Next() {
		// Get values
		values, err := rows.Values()
		if err != nil {
			fetchTime := time.Since(fetchStart)
			return nil, nil, queryTime, fetchTime, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert values to strings
		row := make([]string, len(values))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
				// Check if value is a map or slice (likely JSON/JSONB)
				switch v := val.(type) {
				case map[string]interface{}, []interface{}:
					// Marshal back to JSON string
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						row[i] = fmt.Sprintf("%v", val)
					} else {
						row[i] = string(jsonBytes)
					}
				default:
					row[i] = fmt.Sprintf("%v", val)
				}
			}
		}

		data = append(data, row)
	}
	fetchTime := time.Since(fetchStart)

	if err := rows.Err(); err != nil {
		return nil, nil, queryTime, fetchTime, fmt.Errorf("error iterating rows: %w", err)
	}

	return columns, data, queryTime, fetchTime, nil
}
// UpdateCell updates a single cell in a table
func (p *PostgresDriver) UpdateCell(ctx context.Context, tableSchema *TableSchema, columns []string, oldRow []string, columnName, newValue string) error {
	if p.pool == nil {
		return fmt.Errorf("connection is not established")
	}

	// Build UPDATE query
	var whereParts []string
	var args []interface{}
	argIndex := 1

	// Add the new value as the first argument
	args = append(args, newValue)
	setClause := fmt.Sprintf("%s = $%d", pgx.Identifier{columnName}.Sanitize(), argIndex)
	argIndex++

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
				// Handle NULL values
				whereParts = append(whereParts, fmt.Sprintf("%s IS NULL", pgx.Identifier{col}.Sanitize()))
			} else {
				whereParts = append(whereParts, fmt.Sprintf("%s = $%d", pgx.Identifier{col}.Sanitize(), argIndex))
				args = append(args, oldRow[colIdx])
				argIndex++
			}
		}
	}

	whereClause := strings.Join(whereParts, " AND ")

	// Get schema and table names
	var schemaName, tableName string
	if tableSchema != nil {
		schemaName = tableSchema.SchemaName
		tableName = tableSchema.TableName
	} else {
		return fmt.Errorf("table schema is required for update")
	}

	// Build full UPDATE query
	query := fmt.Sprintf(
		"UPDATE %s.%s SET %s WHERE %s",
		pgx.Identifier{schemaName}.Sanitize(),
		pgx.Identifier{tableName}.Sanitize(),
		setClause,
		whereClause,
	)

	// Execute the UPDATE
	_, err := p.pool.Exec(ctx, query, args...)
	if err != nil {
		// Format query for better readability in error
		formattedQuery := fmt.Sprintf("UPDATE %s.%s\nSET %s\nWHERE %s",
			pgx.Identifier{schemaName}.Sanitize(),
			pgx.Identifier{tableName}.Sanitize(),
			setClause,
			whereClause,
		)
		return fmt.Errorf("failed to update cell: %w\nQuery:\n%s\nArgs: %v", err, formattedQuery, args)
	}

	return nil
}
