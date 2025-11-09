package tui

import (
	"github.com/someshkoli/dessertfrog/pkg/config"
)

// ConvertConfigKeyBindings converts config key bindings to TUI key bindings
func ConvertConfigKeyBindings(cfg *config.Config) map[string][]KeyBinding {
	result := make(map[string][]KeyBinding)

	if len(cfg.KeyBindings.Global) > 0 {
		result["global"] = convertBindingList(cfg.KeyBindings.Global)
	}
	if len(cfg.KeyBindings.Normal) > 0 {
		result["normal"] = convertBindingList(cfg.KeyBindings.Normal)
	}
	if len(cfg.KeyBindings.TableView) > 0 {
		result["table_view"] = convertBindingList(cfg.KeyBindings.TableView)
	}
	if len(cfg.KeyBindings.Search) > 0 {
		result["search"] = convertBindingList(cfg.KeyBindings.Search)
	}
	if len(cfg.KeyBindings.InlineSearch) > 0 {
		result["inline_search"] = convertBindingList(cfg.KeyBindings.InlineSearch)
	}
	if len(cfg.KeyBindings.CommandMode) > 0 {
		result["command_mode"] = convertBindingList(cfg.KeyBindings.CommandMode)
	}
	if len(cfg.KeyBindings.CellEdit) > 0 {
		result["cell_edit"] = convertBindingList(cfg.KeyBindings.CellEdit)
	}
	if len(cfg.KeyBindings.SQLQuery) > 0 {
		result["sql_query"] = convertBindingList(cfg.KeyBindings.SQLQuery)
	}
	if len(cfg.KeyBindings.CellPopup) > 0 {
		result["cell_popup"] = convertBindingList(cfg.KeyBindings.CellPopup)
	}
	if len(cfg.KeyBindings.RecordView) > 0 {
		result["record_view"] = convertBindingList(cfg.KeyBindings.RecordView)
	}
	if len(cfg.KeyBindings.DebugPanel) > 0 {
		result["debug_panel"] = convertBindingList(cfg.KeyBindings.DebugPanel)
	}

	return result
}

// convertBindingList converts a list of config bindings to TUI bindings
func convertBindingList(configBindings []config.KeyBindingConfig) []KeyBinding {
	result := make([]KeyBinding, 0, len(configBindings))
	for _, binding := range configBindings {
		result = append(result, KeyBinding{
			Key:     binding.Key,
			Command: Command(binding.Command),
		})
	}
	return result
}
