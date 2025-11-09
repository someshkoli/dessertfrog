package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// PostgresDriver implements the Driver interface for PostgreSQL databases
type PostgresDriver struct {
	config *Config
	conn   *pgx.Conn
}

// NewPostgresDriver creates a new PostgreSQL driver instance
func NewPostgresDriver(config *Config) *PostgresDriver {
	return &PostgresDriver{
		config: config,
	}
}

// Connect establishes a connection to the PostgreSQL database
func (p *PostgresDriver) Connect(ctx context.Context) error {
	// Build connection string
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.config.Host,
		p.config.Port,
		p.config.User,
		p.config.Password,
		p.config.Database,
		p.config.SSLMode,
	)

	// Establish connection
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	p.conn = conn

	// Verify connection is working
	if err := p.Ping(ctx); err != nil {
		p.conn.Close(ctx)
		return fmt.Errorf("failed to ping postgres after connection: %w", err)
	}

	return nil
}

// Close closes the PostgreSQL connection
func (p *PostgresDriver) Close() error {
	if p.conn != nil {
		return p.conn.Close(context.Background())
	}
	return nil
}

// Ping verifies the connection to the database is still alive
func (p *PostgresDriver) Ping(ctx context.Context) error {
	if p.conn == nil {
		return fmt.Errorf("connection is not established")
	}
	return p.conn.Ping(ctx)
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
	if p.conn == nil {
		return nil, fmt.Errorf("connection is not established")
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
		rows, err = p.conn.Query(ctx, query, p.config.Schema)
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
		rows, err = p.conn.Query(ctx, query)
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
	if p.conn == nil {
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
	err := p.conn.QueryRow(ctx, tableQuery, schemaName, tableName).Scan(&schema.TableType, &comment)
	if err != nil {
		return nil, fmt.Errorf("failed to query table info: %w", err)
	}

	if comment != nil {
		schema.Comment = *comment
	}

	// Now get column information
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
			col_description((c.table_schema || '.' || c.table_name)::regclass, c.ordinal_position) as column_comment
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
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position
	`

	rows, err := p.conn.Query(ctx, columnQuery, schemaName, tableName)
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

	return &schema, nil
}

// GetAllEntities returns all database entities (tables, views, functions, triggers)
func (p *PostgresDriver) GetAllEntities(ctx context.Context) ([]TableSchema, error) {
	if p.conn == nil {
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
		rows, err = p.conn.Query(ctx, query, p.config.Schema)
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
		rows, err = p.conn.Query(ctx, query)
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
		rows, err = p.conn.Query(ctx, query, p.config.Schema)
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
		rows, err = p.conn.Query(ctx, query)
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
		rows, err = p.conn.Query(ctx, query, p.config.Schema)
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
		rows, err = p.conn.Query(ctx, query)
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
func (p *PostgresDriver) GetTableData(ctx context.Context, schemaName, tableName string, limit, offset int) ([]string, [][]string, error) {
	if p.conn == nil {
		return nil, nil, fmt.Errorf("connection is not established")
	}

	// Build the query with LIMIT and OFFSET
	query := fmt.Sprintf(`SELECT * FROM "%s"."%s" LIMIT $1 OFFSET $2`, schemaName, tableName)

	// Execute query
	rows, err := p.conn.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query table data: %w", err)
	}
	defer rows.Close()

	// Get column names
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	// Fetch all rows
	var data [][]string
	for rows.Next() {
		// Get values
		values, err := rows.Values()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
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

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return columns, data, nil
}

// ExecuteQuery executes a custom SQL query and returns the results
func (p *PostgresDriver) ExecuteQuery(ctx context.Context, query string) ([]string, [][]string, error) {
	if p.conn == nil {
		return nil, nil, fmt.Errorf("connection is not established")
	}

	// Execute the query
	rows, err := p.conn.Query(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	// Fetch all rows
	var data [][]string
	for rows.Next() {
		// Get values
		values, err := rows.Values()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
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

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return columns, data, nil
}
// UpdateCell updates a single cell in a table
func (p *PostgresDriver) UpdateCell(ctx context.Context, schemaName, tableName string, columns []string, oldRow []string, columnName, newValue string) error {
	if p.conn == nil {
		return fmt.Errorf("connection is not established")
	}

	// Build UPDATE query with WHERE clause that matches all columns of the old row
	// This ensures we update exactly the row we intend to update
	var whereParts []string
	var args []interface{}
	argIndex := 1

	// Add the new value as the first argument
	args = append(args, newValue)
	setClause := fmt.Sprintf("%s = $%d", pgx.Identifier{columnName}.Sanitize(), argIndex)
	argIndex++

	// Build WHERE clause using all column values to uniquely identify the row
	for i, col := range columns {
		if i < len(oldRow) {
			if oldRow[i] == "NULL" {
				// Handle NULL values
				whereParts = append(whereParts, fmt.Sprintf("%s IS NULL", pgx.Identifier{col}.Sanitize()))
			} else {
				whereParts = append(whereParts, fmt.Sprintf("%s = $%d", pgx.Identifier{col}.Sanitize(), argIndex))
				args = append(args, oldRow[i])
				argIndex++
			}
		}
	}

	whereClause := strings.Join(whereParts, " AND ")

	// Build full UPDATE query
	query := fmt.Sprintf(
		"UPDATE %s.%s SET %s WHERE %s",
		pgx.Identifier{schemaName}.Sanitize(),
		pgx.Identifier{tableName}.Sanitize(),
		setClause,
		whereClause,
	)

	// Execute the UPDATE
	_, err := p.conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update cell: %w", err)
	}

	return nil
}
