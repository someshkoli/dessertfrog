# dessertfrog

A terminal UI database browser for SQL style datastores.
<img width="870" height="567" alt="home" src="https://github.com/user-attachments/assets/6749f022-55f5-4585-bc75-562c65d0c14e" />
<img width="986" height="495" alt="search" src="https://github.com/user-attachments/assets/31fc4f07-53bb-44e8-b8a4-91a8e2ed6e86" />
<img width="1064" height="603" alt="table" src="https://github.com/user-attachments/assets/1d13223a-c18b-431d-a557-3fd959ea0d9e" />
<img width="611" height="399" alt="json-walker" src="https://github.com/user-attachments/assets/0e8d962d-14ad-43e5-a210-603dd8a0d8e5" />

## Features

- Browse database tables, views, materialized views, functions, and triggers
- View table schemas with column details (data types, nullability, defaults, constraints)
- Browse table data with pagination (500 rows per page)
- Edit cell values inline
- Search and filter tables by name
- Search within table data (content search)
- JSON viewer with expandable/collapsible tree structure
- Copy values to clipboard
- Navigate between tables and maintain history
- Customizable key bindings via configuration file
- Debug mode with real-time logging

## Installation

```bash
go install github.com/someshkoli/dessertfrog@latest
```

Or build from source:

```bash
git clone https://github.com/someshkoli/dessertfrog.git
cd dessertfrog
go build
```

## Usage

```bash
dessertfrog --driver postgres --host localhost --port 5432 --username postgres --password yourpass --database mydb --schema public
```

### Command Line Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config-file` | `-c` | `~/.config/dessertfrog/config.yaml` | Path to configuration file |
| `--driver` | `-d` | `postgres` | Database driver (postgres, mariadb) |
| `--host` | | `localhost` | Database host |
| `--port` | `-p` | 5432/3306 | Database port (depends on driver) |
| `--username` | `-u` | postgres/root | Database username (depends on driver) |
| `--password` | `-P` | | Database password |
| `--database` | `-n` | postgres/mysql | Database name (depends on driver) |
| `--schema` | `-s` | public | Database schema (PostgreSQL only) |

### Keyboard Shortcuts

#### Table List View
- `↑/↓` or `k/j` - Navigate tables
- `Ctrl+D` - Page down (jump down by half viewport)
- `Ctrl+U` - Page up (jump up by half viewport)
- `g/G` - Jump to first/last table
- `Enter` - View table data
- `/` - Inline search
- `Ctrl+P` - Open search popup (search all entities)
- `Ctrl+H` - Navigate back in history
- `Ctrl+L` - Navigate forward in history
- `s` - Open SQL query editor
- `q` - Quit

#### Table Data View
- `↑/↓/←/→` or `h/j/k/l` - Navigate cells
- `Ctrl+D` - Page down (jump down by half viewport)
- `Ctrl+U` - Page up (jump up by half viewport)
- `g/G` - Jump to first/last row
- `w/b` - Jump forward/backward between cells
- `v` - View cell value (opens popup for large values)
- `V` - View record (opens popup for entire row)
- `i` - Edit cell value
- `y` - Copy cell value to clipboard
- `Y` - Copy record (entire row) to clipboard
- `n/p` - Next/previous page (500 rows)
- `/` - Search within table data
- `Ctrl+P` - Open search popup to switch tables
- `Ctrl+H` - Navigate back in history
- `Ctrl+L` - Navigate forward in history
- `s` - Open SQL query editor
- `Esc` - Clear filter or return to table list
- `q` - Quit

#### Cell Value Popup
- `↑/↓` or `k/j` - Scroll (or navigate JSON tree)
- `g/G` - Jump to first/last item
- `h/l` or `←/→` - Collapse/expand JSON node
- `Enter` - Toggle JSON node expansion
- `Esc` or `q` - Close popup

#### Record View Popup
- `↑/↓` or `k/j` - Navigate through fields
- `g/G` - Jump to first/last field
- `v` - View selected field value (opens cell popup)
- `Esc` or `q` - Close popup

#### Debug Mode
- `F12` - Toggle debug overlay
- `F11` - Clear debug logs (when debug mode is active)
- `F10` - Focus/unfocus debug panel (enables navigation)

When debug panel is focused:
- `Tab` - Switch between App State and Debug Logs sections
- `j/k` or `↑/↓` - Navigate through log entries
- `g/G` - Jump to first/last log entry
- `Enter` or `Space` - Open detail popup for selected item
- `Esc` - Unfocus debug panel

The debug overlay displays:
- Connection information
- Current view state and mode flags
- Table/data statistics
- Navigation history info
- Window dimensions
- Real-time debug logs with timestamps
- Visual selection highlighting when focused
- Expandable detail view for any log entry or full app state

## Configuration

dessertfrog supports configuration through a YAML file. By default, the application looks for a config file at `~/.config/dessertfrog/config.yaml`. If the file doesn't exist, built-in defaults are used.

### Generating a Configuration File

You can generate a configuration file with all default key bindings using the built-in command:

```bash
# Generate config at default location (~/.config/dessertfrog/config.yaml)
dessertfrog generate-config

# Generate config at a custom location
dessertfrog generate-config --output /path/to/config.yaml

# Overwrite existing config file
dessertfrog generate-config --force
```

### Using a Custom Configuration File

You can specify a custom config file location when starting the application:

```bash
dessertfrog --config-file /path/to/config.yaml --driver postgres ...
```

Alternatively, you can manually copy the sample configuration file from the repository:

```bash
mkdir -p ~/.config/dessertfrog
cp config.yaml.sample ~/.config/dessertfrog/config.yaml
```

### Key Binding Customization

The configuration file allows you to customize key bindings without modifying the code. Key bindings defined in the config file override the default bindings.

Example configuration structure:

```yaml
keybindings:
  # Global bindings (work in all modes)
  global:
    - key: "f12"
      command: "toggle_debug"
    - key: "ctrl+c"
      command: "quit"

  # Normal mode (table list view)
  normal:
    - key: "j"
      command: "navigate_down"
    - key: "ctrl+h"
      command: "history_back"

  # Table view mode
  table_view:
    - key: "ctrl+d"
      command: "page_down"
    - key: "ctrl+u"
      command: "page_up"
```

**Available Commands:**
- Navigation: `navigate_up`, `navigate_down`, `navigate_left`, `navigate_right`, `page_up`, `page_down`, `go_to_top`, `go_to_bottom`, `history_back`, `history_forward`
- Actions: `confirm`, `cancel`, `quit`, `back`
- View: `open_table`, `open_search`, `open_sql_query`, `open_command_mode`, `inline_search`
- Table: `next_page`, `previous_page`, `edit_cell`, `copy_cell_value`, `copy_row`, `open_cell_popup`, `open_record_view`, `filter_content`
- Debug: `toggle_debug`, `clear_debug_logs`, `toggle_debug_focus`
- Popup: `toggle_json_node`, `switch_debug_section`

**Key Format Examples:**
- Single keys: `"j"`, `"k"`, `"g"`, `"G"`, `"enter"`, `"esc"`, `"tab"`
- Arrow keys: `"up"`, `"down"`, `"left"`, `"right"`
- Ctrl combinations: `"ctrl+h"`, `"ctrl+l"`, `"ctrl+d"`, `"ctrl+p"`
- Alt combinations: `"alt+left"`, `"alt+right"`
- Function keys: `"f1"`, `"f12"`

**Binding Modes:**
- `global` - Work in any mode
- `normal` - Table list view
- `table_view` - Table data view
- `search` - Search popup
- `inline_search` - Inline search mode
- `command_mode` - Command input
- `cell_edit` - Cell editing
- `sql_query` - SQL query input
- `cell_popup` - Cell value popup
- `record_view` - Record view popup
- `debug_panel` - Debug panel

## Supported Databases

- PostgreSQL
- MariaDB (not supported yet)

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [clipboard](https://github.com/atotto/clipboard) - Clipboard operations

## License

See LICENSE file for details.
