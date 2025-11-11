package tui

import (
	"fmt"
	"strings"

	"github.com/someshkoli/dessertfrog/pkg/driver"
	"github.com/someshkoli/dessertfrog/pkg/helpers"
)

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// fuzzyMatch performs fuzzy matching on a string
// Uses the common helpers package
func fuzzyMatch(query, target string) bool {
	return helpers.FuzzyMatch(query, target)
}

// filterTables filters tables based on fuzzy search query with optional prefix
// Supports prefix format: "table/name", "view/name", "function/name", etc.
func filterTables(tables []driver.TableSchema, query string) []driver.TableSchema {
	if query == "" {
		return tables
	}

	// Check if query has a prefix (e.g., "table/", "view/", etc.)
	var entityTypeFilter driver.EntityType
	var searchTerm string

	if strings.Contains(query, "/") {
		parts := strings.SplitN(query, "/", 2)
		prefix := strings.ToLower(strings.TrimSpace(parts[0]))
		searchTerm = strings.TrimSpace(parts[1])

		// Map prefix to entity type
		switch prefix {
		case "table", "t":
			entityTypeFilter = driver.EntityTable
		case "view", "v":
			entityTypeFilter = driver.EntityView
		case "materialized_view", "mview", "mv":
			entityTypeFilter = driver.EntityMaterializedView
		case "function", "func", "f":
			entityTypeFilter = driver.EntityFunction
		case "trigger", "trig":
			entityTypeFilter = driver.EntityTrigger
		}
	} else {
		searchTerm = query
	}

	var filtered []driver.TableSchema
	for _, table := range tables {
		// If entity type filter is set, only match that type
		if entityTypeFilter != "" && table.EntityType != entityTypeFilter {
			continue
		}

		// Match against table name or schema name
		if fuzzyMatch(searchTerm, table.TableName) || fuzzyMatch(searchTerm, table.SchemaName) {
			filtered = append(filtered, table)
		}
	}
	return filtered
}

// getEntityTypePrefix returns the display prefix for an entity type
func getEntityTypePrefix(entityType driver.EntityType) string {
	switch entityType {
	case driver.EntityTable:
		return "table"
	case driver.EntityView:
		return "view"
	case driver.EntityMaterializedView:
		return "mview"
	case driver.EntityFunction:
		return "function"
	case driver.EntityTrigger:
		return "trigger"
	default:
		return "unknown"
	}
}

// formatEntityDisplay formats an entity for display with type prefix
func formatEntityDisplay(entity driver.TableSchema) string {
	prefix := getEntityTypePrefix(entity.EntityType)
	return fmt.Sprintf("%s/%s", prefix, entity.TableName)
}

// getAutocompleteSuggestion returns an autocomplete suggestion for the query
// Returns empty string if no suggestion
func getAutocompleteSuggestion(query string) string {
	if query == "" {
		return ""
	}

	// List of available prefixes
	prefixes := []string{"table/", "view/", "materialized_view/", "mview/", "function/", "trigger/"}

	// Check if query already contains a slash
	if strings.Contains(query, "/") {
		return ""
	}

	lowerQuery := strings.ToLower(query)

	// Find matching prefix
	for _, prefix := range prefixes {
		if strings.HasPrefix(prefix, lowerQuery) {
			// Return the remaining part to complete
			return prefix[len(query):]
		}
	}

	return ""
}

// filterTableData filters table rows based on fuzzy search query
// Searches across all columns in each row
func filterTableData(rows [][]string, query string) [][]string {
	if query == "" {
		return rows
	}

	var filtered [][]string
	queryLower := strings.ToLower(query)

	for _, row := range rows {
		// Check if any cell in the row matches the query
		for _, cell := range row {
			if fuzzyMatch(queryLower, cell) {
				filtered = append(filtered, row)
				break // Move to next row once we find a match
			}
		}
	}

	return filtered
}
