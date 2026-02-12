/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/someshkoli/dessertfrog/pkg/config"
	"github.com/someshkoli/dessertfrog/pkg/tui"
	"github.com/spf13/cobra"
)

// Database connection flags
var (
	dbDriver   string
	dbHost     string
	dbPort     int
	dbUsername string
	dbPassword string
	dbName     string
	dbSchema   string
	configFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dessertfrog",
	Short: "A TUI database browser for MariaDB, PostgreSQL, and ClickHouse",
	Long: `dessertfrog is a terminal UI database browser that supports multiple SQL databases.
Currently supported databases:
  - MariaDB
  - PostgreSQL
  - ClickHouse

Connect to your database and browse tables, schemas, and data with an intuitive terminal interface.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// If no driver is specified, allow starting without database connection
		if dbDriver == "" {
			return nil
		}

		// Set driver-specific defaults if not explicitly provided
		if !cmd.Flags().Changed("host") {
			dbHost = "localhost"
		}

		if !cmd.Flags().Changed("port") {
			switch dbDriver {
			case "mariadb", "mysql":
				dbPort = 3306
			case "postgres", "postgresql":
				dbPort = 5432
			case "clickhouse", "ch":
				dbPort = 9000
			default:
				return fmt.Errorf("unsupported database driver: %s (supported: mariadb, postgres, clickhouse)", dbDriver)
			}
		}

		if !cmd.Flags().Changed("username") {
			switch dbDriver {
			case "mariadb", "mysql":
				dbUsername = "root"
			case "postgres", "postgresql":
				dbUsername = "postgres"
			case "clickhouse", "ch":
				dbUsername = "default"
			}
		}

		if !cmd.Flags().Changed("database") {
			switch dbDriver {
			case "mariadb", "mysql":
				dbName = "mysql"
			case "postgres", "postgresql":
				dbName = "postgres"
			case "clickhouse", "ch":
				dbName = "default"
			}
		}

		if !cmd.Flags().Changed("schema") {
			switch dbDriver {
			case "postgres", "postgresql":
				dbSchema = "public"
			case "mariadb", "mysql":
				dbSchema = "" // MySQL doesn't use schemas the same way
			case "clickhouse", "ch":
				dbSchema = "" // ClickHouse uses databases, not schemas
			}
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration file
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			fmt.Printf("Error loading config file: %v\n", err)
			os.Exit(1)
		}

		// Merge key bindings from config with defaults
		defaultKeyBindings := tui.DefaultKeyBindings()
		configKeyBindings := tui.ConvertConfigKeyBindings(cfg)
		keyBindings := tui.MergeKeyBindings(defaultKeyBindings, configKeyBindings)

		// Create styles from color scheme
		mergedColorScheme := cfg.ColorScheme.MergeWithDefaults()
		styles := tui.NewStyles(mergedColorScheme)

		// Create database configuration
		dbConfig := tui.DBConfig{
			Driver:   dbDriver,
			Host:     dbHost,
			Port:     dbPort,
			Username: dbUsername,
			Password: dbPassword,
			Database: dbName,
			Schema:   dbSchema,
		}

		// Create and start the bubbletea program
		p := tea.NewProgram(tui.NewModel(dbConfig, keyBindings, styles), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Configuration file flag
	rootCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "Path to config file (default: ~/.config/dessertfrog/config.yaml)")

	// Database connection flags (all optional now)
	rootCmd.Flags().StringVarP(&dbDriver, "driver", "d", "", "Database driver (mariadb, postgres, clickhouse)")
	rootCmd.Flags().StringVarP(&dbHost, "host", "", "", "Database host (default: localhost)")
	rootCmd.Flags().IntVarP(&dbPort, "port", "p", 0, "Database port (default: 5432 for postgres, 3306 for mariadb, 9000 for clickhouse)")
	rootCmd.Flags().StringVarP(&dbUsername, "username", "u", "", "Database username (default: postgres for postgres, root for mariadb, default for clickhouse)")
	rootCmd.Flags().StringVarP(&dbPassword, "password", "P", "", "Database password")
	rootCmd.Flags().StringVarP(&dbName, "database", "n", "", "Database name (default: postgres for postgres, mysql for mariadb, default for clickhouse)")
	rootCmd.Flags().StringVarP(&dbSchema, "schema", "s", "", "Database schema (default: public for postgres)")
}
