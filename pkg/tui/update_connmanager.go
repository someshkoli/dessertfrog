package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/someshkoli/dessertfrog/pkg/connhistory"
	"github.com/someshkoli/dessertfrog/pkg/driver"
	"github.com/someshkoli/dessertfrog/pkg/helpers"
	"github.com/someshkoli/dessertfrog/pkg/sqlhistory"
)

// handleConnectionManagerKeys handles key presses in connection manager mode
func (m Model) handleConnectionManagerKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	// Handle escape key - different behavior based on mode
	if key == "esc" {
		if m.connManagerInsertMode {
			// In insert mode, escape goes to normal mode
			m.connManagerInsertMode = false
			return m, nil
		} else {
			// In normal mode, escape closes the popup
			m.connManagerMode = false
			m.connManagerFilter = ""
			m.connManagerSelected = 0
			m.connManagerScroll = 0
			m.connManagerInsertMode = true // Reset to insert mode for next time
			return m, nil
		}
	}

	// Handle 'i' key - enter insert mode from normal mode
	if key == "i" && !m.connManagerInsertMode {
		m.connManagerInsertMode = true
		return m, nil
	}

	// Handle 'q' key - always closes in normal mode
	if key == "q" && !m.connManagerInsertMode {
		m.connManagerMode = false
		m.connManagerFilter = ""
		m.connManagerSelected = 0
		m.connManagerScroll = 0
		m.connManagerInsertMode = true // Reset to insert mode for next time
		return m, nil
	}

	// Insert mode - all keys go to filter
	if m.connManagerInsertMode {
		switch key {
		case "backspace":
			// Remove last character from filter
			if len(m.connManagerFilter) > 0 {
				m.connManagerFilter = m.connManagerFilter[:len(m.connManagerFilter)-1]
				// Update filtered connections
				if m.connHistory != nil {
					m.filteredConnections = m.connHistory.Filter(m.connManagerFilter)
				}
				// Reset selection
				m.connManagerSelected = 0
				m.connManagerScroll = 0
			}
			return m, nil

		case "enter":
			// In insert mode, enter still selects and switches
			if len(m.filteredConnections) > 0 && m.connManagerSelected < len(m.filteredConnections) {
				selectedConn := m.filteredConnections[m.connManagerSelected]

				// Close connection manager
				m.connManagerMode = false
				m.connManagerFilter = ""
				m.connManagerInsertMode = true // Reset for next time

				// Switch to the selected connection
				return m, switchConnection(selectedConn)
			}
			return m, nil

		default:
			// Add character to filter (if it's a printable character)
			if len(key) == 1 {
				m.connManagerFilter += key
				// Update filtered connections
				if m.connHistory != nil {
					m.filteredConnections = m.connHistory.Filter(m.connManagerFilter)
				}
				// Reset selection
				m.connManagerSelected = 0
				m.connManagerScroll = 0
			}
			return m, nil
		}
	}

	// Normal/Navigate mode - hjkl and other navigation keys
	switch key {
	case "j", "down":
		// Navigate down
		if len(m.filteredConnections) > 0 {
			m.connManagerSelected++
			if m.connManagerSelected >= len(m.filteredConnections) {
				m.connManagerSelected = len(m.filteredConnections) - 1
			}

			// Auto-scroll
			maxVisible := helpers.Min(20, m.height-14)
			if m.connManagerSelected >= m.connManagerScroll+maxVisible {
				m.connManagerScroll = m.connManagerSelected - maxVisible + 1
			}
		}
		return m, nil

	case "k", "up":
		// Navigate up
		if len(m.filteredConnections) > 0 {
			m.connManagerSelected--
			if m.connManagerSelected < 0 {
				m.connManagerSelected = 0
			}

			// Auto-scroll
			if m.connManagerSelected < m.connManagerScroll {
				m.connManagerScroll = m.connManagerSelected
			}
		}
		return m, nil

	case "g":
		// Go to top
		m.connManagerSelected = 0
		m.connManagerScroll = 0
		return m, nil

	case "G":
		// Go to bottom
		if len(m.filteredConnections) > 0 {
			m.connManagerSelected = len(m.filteredConnections) - 1
			maxVisible := helpers.Min(20, m.height-14)
			m.connManagerScroll = helpers.Max(0, len(m.filteredConnections)-maxVisible)
		}
		return m, nil

	case "enter":
		// Select connection and switch to it
		if len(m.filteredConnections) > 0 && m.connManagerSelected < len(m.filteredConnections) {
			selectedConn := m.filteredConnections[m.connManagerSelected]

			// Close connection manager
			m.connManagerMode = false
			m.connManagerFilter = ""
			m.connManagerInsertMode = true // Reset for next time

			// Switch to the selected connection
			return m, switchConnection(selectedConn)
		}
		return m, nil

	default:
		return m, nil
	}
}

// switchConnection switches to a different database connection
func switchConnection(conn connhistory.ConnectionEntry) tea.Cmd {
	return func() tea.Msg {
		// Create driver configuration
		driverConfig := &driver.Config{
			Host:     conn.Host,
			Port:     conn.Port,
			Database: conn.Database,
			Schema:   conn.Schema,
			User:     conn.Username,
			Password: conn.Password,
			SSLMode:  "disable", // Default
		}

		// Create driver based on connection type
		var drv driver.Driver
		switch conn.Driver {
		case "postgres", "postgresql":
			drv = driver.NewPostgresDriver(driverConfig)
		case "clickhouse", "ch":
			drv = driver.NewClickHouseDriver(driverConfig)
		default:
			drv = driver.NewPostgresDriver(driverConfig)
		}

		// Try to connect
		ctx := context.Background()
		if err := drv.Connect(ctx); err != nil {
			return connectionSwitchFailedMsg{err: err}
		}

		// Initialize SQL history for new connection
		sqlHist, _ := sqlhistory.NewHistory(
			conn.Driver,
			conn.Host,
			conn.Port,
			conn.Database,
			conn.Schema,
			conn.Username,
			1000,
		)

		return connectionSwitchSuccessMsg{
			driver: drv,
			dbConfig: DBConfig{
				Driver:   conn.Driver,
				Host:     conn.Host,
				Port:     conn.Port,
				Username: conn.Username,
				Password: conn.Password,
				Database: conn.Database,
				Schema:   conn.Schema,
			},
			sqlHistory: sqlHist,
		}
	}
}

// connectionSwitchSuccessMsg is sent when connection switch succeeds
type connectionSwitchSuccessMsg struct {
	driver     driver.Driver
	dbConfig   DBConfig
	sqlHistory *sqlhistory.History
}

// connectionSwitchFailedMsg is sent when connection switch fails
type connectionSwitchFailedMsg struct {
	err error
}
