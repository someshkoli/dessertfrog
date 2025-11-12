package config

// KeyBindingConfig represents a single key binding in the config
type KeyBindingConfig struct {
	Key     string `yaml:"key"`
	Command string `yaml:"command"`
}

// KeyBindingsConfig holds all key bindings configuration
type KeyBindingsConfig struct {
	Global       []KeyBindingConfig `yaml:"global,omitempty"`
	Normal       []KeyBindingConfig `yaml:"normal,omitempty"`
	TableView    []KeyBindingConfig `yaml:"table_view,omitempty"`
	Search       []KeyBindingConfig `yaml:"search,omitempty"`
	InlineSearch []KeyBindingConfig `yaml:"inline_search,omitempty"`
	CommandMode  []KeyBindingConfig `yaml:"command_mode,omitempty"`
	CellEdit     []KeyBindingConfig `yaml:"cell_edit,omitempty"`
	SQLQuery     []KeyBindingConfig `yaml:"sql_query,omitempty"`
	CellPopup    []KeyBindingConfig `yaml:"cell_popup,omitempty"`
	RecordView   []KeyBindingConfig `yaml:"record_view,omitempty"`
	DebugPanel   []KeyBindingConfig `yaml:"debug_panel,omitempty"`
}
