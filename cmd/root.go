/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dessertfrog",
	Short: "A TUI database browser for MariaDB and PostgreSQL",
	Long: `dessertfrog is a terminal UI database browser that supports multiple SQL databases.
Currently supported databases:
  - MariaDB
  - PostgreSQL

Connect to your database and browse tables, schemas, and data with an intuitive terminal interface.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
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
			default:
				return fmt.Errorf("unsupported database driver: %s (supported: mariadb, postgres)", dbDriver)
			}
		}

		if !cmd.Flags().Changed("username") {
			switch dbDriver {
			case "mariadb", "mysql":
				dbUsername = "root"
			case "postgres", "postgresql":
				dbUsername = "postgres"
			}
		}

		if !cmd.Flags().Changed("database") {
			switch dbDriver {
			case "mariadb", "mysql":
				dbName = "mysql"
			case "postgres", "postgresql":
				dbName = "postgres"
			}
		}

		if !cmd.Flags().Changed("schema") {
			switch dbDriver {
			case "postgres", "postgresql":
				dbSchema = "public"
			case "mariadb", "mysql":
				dbSchema = "" // MySQL doesn't use schemas the same way
			}
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Create database configuration
		config := tui.DBConfig{
			Driver:   dbDriver,
			Host:     dbHost,
			Port:     dbPort,
			Username: dbUsername,
			Password: dbPassword,
			Database: dbName,
			Schema:   dbSchema,
		}

		// Create and start the bubbletea program
		p := tea.NewProgram(tui.NewModel(config), tea.WithAltScreen())
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
	// Database connection flags
	rootCmd.Flags().StringVarP(&dbDriver, "driver", "d", "postgres", "Database driver (mariadb, postgres)")
	rootCmd.Flags().StringVarP(&dbHost, "host", "", "", "Database host (default: localhost)")
	rootCmd.Flags().IntVarP(&dbPort, "port", "p", 0, "Database port (default: 5432 for postgres, 3306 for mariadb)")
	rootCmd.Flags().StringVarP(&dbUsername, "username", "u", "", "Database username (default: postgres for postgres, root for mariadb)")
	rootCmd.Flags().StringVarP(&dbPassword, "password", "P", "", "Database password")
	rootCmd.Flags().StringVarP(&dbName, "database", "n", "", "Database name (default: postgres for postgres, mysql for mariadb)")
	rootCmd.Flags().StringVarP(&dbSchema, "schema", "s", "", "Database schema (default: public for postgres)")

	// Mark driver as required
	rootCmd.MarkFlagRequired("driver")
}
