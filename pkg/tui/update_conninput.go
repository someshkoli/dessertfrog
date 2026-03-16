package tui

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/someshkoli/dessertfrog/pkg/connhistory"
)

// handleConnectionInputKeys handles key presses in connection input mode
func (m Model) handleConnectionInputKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	// Handle escape key - cancel and go back to connection manager
	if key == "esc" {
		m.connInputMode = false
		// Reset the form
		m.connInputs = m.makeConnectionInputs()
		m.connInputField = 0
		m.connectionError = ""
		// Go back to connection manager or disconnected state
		if m.connectionStatus == Disconnected {
			// Stay in disconnected state, will show connection manager view
			if m.connHistory != nil {
				m.filteredConnections = m.connHistory.GetAll()
			}
		} else {
			// We have a connection, reopen connection manager popup
			m.connManagerMode = true
			if m.connHistory != nil {
				m.filteredConnections = m.connHistory.GetAll()
			}
		}
		return m, nil
	}

	// Handle tab and shift+tab for field navigation
	if key == "tab" {
		m.connInputs[m.connInputField].Blur()
		m.connInputField++
		if m.connInputField > 7 { // 8 fields total (0-7)
			m.connInputField = 0
		}
		m.connInputs[m.connInputField].Focus()
		return m, nil
	}

	if key == "shift+tab" {
		m.connInputs[m.connInputField].Blur()
		m.connInputField--
		if m.connInputField < 0 {
			m.connInputField = 7
		}
		m.connInputs[m.connInputField].Focus()
		return m, nil
	}

	// Handle enter key - submit connection
	if key == "enter" {
		return m.submitNewConnection()
	}

	// Handle vim-style navigation in normal mode (ctrl+[ acts as escape from input)
	if key == "ctrl+[" {
		// Treat as escape
		return m.handleConnectionInputKeys(tea.KeyMsg{Type: tea.KeyEsc})
	}

	// Pass other keys to the focused textinput
	var cmd tea.Cmd
	m.connInputs[m.connInputField], cmd = m.connInputs[m.connInputField].Update(msg)
	return m, cmd
}

// submitNewConnection creates a new database connection from the input form
func (m Model) submitNewConnection() (Model, tea.Cmd) {
	// Get values from textinputs
	driver := strings.TrimSpace(m.connInputs[0].Value())
	host := strings.TrimSpace(m.connInputs[1].Value())
	portStr := strings.TrimSpace(m.connInputs[2].Value())
	username := strings.TrimSpace(m.connInputs[3].Value())
	password := m.connInputs[4].Value() // Don't trim password
	database := strings.TrimSpace(m.connInputs[5].Value())
	schema := strings.TrimSpace(m.connInputs[6].Value())
	sslMode := strings.TrimSpace(m.connInputs[7].Value())

	// Validate inputs
	if driver == "" {
		m.connectionError = "Driver is required"
		return m, nil
	}

	// Set default host if empty
	if host == "" {
		host = "localhost"
	}

	// Parse port
	var port int
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			m.connectionError = "Invalid port number"
			return m, nil
		}
		port = p
	} else {
		// Set default port based on driver
		switch strings.ToLower(driver) {
		case "postgres", "postgresql":
			port = 5432
		case "mariadb", "mysql":
			port = 3306
		case "clickhouse", "ch":
			port = 9000
		default:
			m.connectionError = "Unknown driver: " + driver
			return m, nil
		}
	}

	// Set default username if not provided
	if username == "" {
		switch strings.ToLower(driver) {
		case "postgres", "postgresql":
			username = "postgres"
		case "mariadb", "mysql":
			username = "root"
		case "clickhouse", "ch":
			username = "default"
		}
	}

	// Set default database if not provided
	if database == "" {
		switch strings.ToLower(driver) {
		case "postgres", "postgresql":
			database = "postgres"
		case "mariadb", "mysql":
			database = "mysql"
		case "clickhouse", "ch":
			database = "default"
		}
	}

	// Set default schema for postgres
	if schema == "" && (strings.ToLower(driver) == "postgres" || strings.ToLower(driver) == "postgresql") {
		schema = "public"
	}

	// Set default SSL mode
	if sslMode == "" {
		sslMode = "disable"
	}

	// Clear any previous error and close input form
	m.connectionError = ""
	m.connInputMode = false
	m.connectionStatus = Connecting

	// Create connection entry
	connEntry := connhistory.ConnectionEntry{
		Driver:   driver,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
		Schema:   schema,
		SSLMode:  sslMode,
	}

	// Switch to the new connection
	return m, switchConnection(connEntry)
}
