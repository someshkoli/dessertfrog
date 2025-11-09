package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/someshkoli/dessertfrog/pkg/config"
	"github.com/someshkoli/dessertfrog/pkg/tui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var generateConfigCmd = &cobra.Command{
	Use:   "generate-config",
	Short: "Generate a sample configuration file",
	Long: `Generate a sample configuration file with all available options and defaults.

By default, this will create ~/.config/dessertfrog/config.yaml
You can specify a custom output location with the --output flag.

If the file already exists, use --force to overwrite it.`,
	Run: func(cmd *cobra.Command, args []string) {
		outputPath, _ := cmd.Flags().GetString("output")
		force, _ := cmd.Flags().GetBool("force")

		// Use default path if not specified
		if outputPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Error getting home directory: %v\n", err)
				os.Exit(1)
			}
			outputPath = filepath.Join(homeDir, ".config", "dessertfrog", "config.yaml")
		}

		// Check if file already exists
		if _, err := os.Stat(outputPath); err == nil && !force {
			fmt.Printf("Config file already exists at: %s\n", outputPath)
			fmt.Println("Use --force to overwrite it")
			os.Exit(1)
		}

		// Get default key bindings
		defaults := tui.DefaultKeyBindings()

		// Convert to config format
		cfg := config.Config{
			KeyBindings: config.KeyBindingsConfig{
				Global:       convertToConfigBindings(defaults.Global),
				Normal:       convertToConfigBindings(defaults.Normal),
				TableView:    convertToConfigBindings(defaults.TableView),
				Search:       convertToConfigBindings(defaults.Search),
				InlineSearch: convertToConfigBindings(defaults.InlineSearch),
				CommandMode:  convertToConfigBindings(defaults.CommandMode),
				CellEdit:     convertToConfigBindings(defaults.CellEdit),
				SQLQuery:     convertToConfigBindings(defaults.SQLQuery),
				CellPopup:    convertToConfigBindings(defaults.CellPopup),
				RecordView:   convertToConfigBindings(defaults.RecordView),
				DebugPanel:   convertToConfigBindings(defaults.DebugPanel),
			},
		}

		// Marshal to YAML
		data, err := yaml.Marshal(&cfg)
		if err != nil {
			fmt.Printf("Error marshaling config to YAML: %v\n", err)
			os.Exit(1)
		}

		// Create directory if it doesn't exist
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating config directory: %v\n", err)
			os.Exit(1)
		}

		// Write the config
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			fmt.Printf("Error writing config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Sample configuration file generated at: %s\n", outputPath)
		fmt.Println("\nEdit this file to customize your key bindings and other settings.")
	},
}

// convertToConfigBindings converts TUI key bindings to config key bindings
func convertToConfigBindings(bindings []tui.KeyBinding) []config.KeyBindingConfig {
	result := make([]config.KeyBindingConfig, len(bindings))
	for i, binding := range bindings {
		result[i] = config.KeyBindingConfig{
			Key:     binding.Key,
			Command: string(binding.Command),
		}
	}
	return result
}

func init() {
	rootCmd.AddCommand(generateConfigCmd)
	generateConfigCmd.Flags().StringP("output", "o", "", "Output path for the config file (default: ~/.config/dessertfrog/config.yaml)")
	generateConfigCmd.Flags().BoolP("force", "f", false, "Overwrite existing config file")
}
