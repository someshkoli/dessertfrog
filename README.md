# dessertfrog

A terminal UI database browser for SQL style datastores.

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
| `--driver` | `-d` | `postgres` | Database driver (postgres, mariadb) |
| `--host` | | `localhost` | Database host |
| `--port` | `-p` | 5432/3306 | Database port (depends on driver) |
| `--username` | `-u` | postgres/root | Database username (depends on driver) |
| `--password` | `-P` | | Database password |
| `--database` | `-n` | postgres/mysql | Database name (depends on driver) |
| `--schema` | `-s` | public | Database schema (PostgreSQL only) |

### Keyboard Shortcuts

#### Table List View - `↑/↓` or `k/j` - Navigate tables
- `Enter` - View table data
- `/` - Open search popup
- `Ctrl+F` - Inline search
- `s` - OpenSQL query editor
- `q` - Quit

#### Table Data View
- `↑/↓/←/→` or `h/j/k/l` - Navigate cells
- `v` - View cell value (opens popup for large values)
- `V` - View record (opens popup for entire row)
- `y` - Copy cell value to clipboard
- `Y` - Copy record (entire row) to clipboard
- `w/b` - Jump forward/backward between cells
- `i` - Edit cell value
- `Esc` - Return to table list
- `/` - Search within table data
- `n/p` - Next/previous page
- `s` - OpenSQL query editor

#### Cell/Record Popup
- `↑/↓` or `k/j` - Scroll (or navigate JSON tree)
- `Space` or `Enter` - Toggle JSON node expansion
- `y` - Copy value to clipboard
- `Esc` - Close popup

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
