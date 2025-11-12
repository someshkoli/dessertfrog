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
		// Clear the form
		m.connInputDriver = ""
		m.connInputHost = ""
		m.connInputPort = ""
		m.connInputUsername = ""
		m.connInputPassword = ""
		m.connInputDatabase = ""
		m.connInputSchema = ""
		m.connInputCursor = 0
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
		m.connInputField++
		if m.connInputField > 6 { // 7 fields total (0-6)
			m.connInputField = 0
		}
		m.connInputCursor = len(m.getCurrentInputValue())
		return m, nil
	}

	if key == "shift+tab" {
		m.connInputField--
		if m.connInputField < 0 {
			m.connInputField = 6
		}
		m.connInputCursor = len(m.getCurrentInputValue())
		return m, nil
	}

	// Handle enter key - submit connection
	if key == "enter" {
		return m.submitNewConnection()
	}

	// Handle backspace
	if key == "backspace" {
		currentValue := m.getCurrentInputValue()
		if m.connInputCursor > 0 && m.connInputCursor <= len([]rune(currentValue)) {
			runes := []rune(currentValue)
			newValue := string(runes[:m.connInputCursor-1]) + string(runes[m.connInputCursor:])
			m.setCurrentInputValue(newValue)
			m.connInputCursor--
		}
		return m, nil
	}

	// Handle left/right arrow keys for cursor movement
	if key == "left" {
		if m.connInputCursor > 0 {
			m.connInputCursor--
		}
		return m, nil
	}

	if key == "right" {
		currentValue := m.getCurrentInputValue()
		if m.connInputCursor < len([]rune(currentValue)) {
			m.connInputCursor++
		}
		return m, nil
	}

	// Handle home/end keys
	if key == "home" {
		m.connInputCursor = 0
		return m, nil
	}

	if key == "end" {
		currentValue := m.getCurrentInputValue()
		m.connInputCursor = len([]rune(currentValue))
		return m, nil
	}

	// Handle regular character input
	if len(key) == 1 {
		currentValue := m.getCurrentInputValue()
		runes := []rune(currentValue)
		newValue := string(runes[:m.connInputCursor]) + key + string(runes[m.connInputCursor:])
		m.setCurrentInputValue(newValue)
		m.connInputCursor++
		return m, nil
	}

	return m, nil
}

// getCurrentInputValue returns the value of the currently focused input field
func (m Model) getCurrentInputValue() string {
	switch m.connInputField {
	case 0:
		return m.connInputDriver
	case 1:
		return m.connInputHost
	case 2:
		return m.connInputPort
	case 3:
		return m.connInputUsername
	case 4:
		return m.connInputPassword
	case 5:
		return m.connInputDatabase
	case 6:
		return m.connInputSchema
	default:
		return ""
	}
}

// setCurrentInputValue sets the value of the currently focused input field
func (m *Model) setCurrentInputValue(value string) {
	switch m.connInputField {
	case 0:
		m.connInputDriver = value
	case 1:
		m.connInputHost = value
	case 2:
		m.connInputPort = value
	case 3:
		m.connInputUsername = value
	case 4:
		m.connInputPassword = value
	case 5:
		m.connInputDatabase = value
	case 6:
		m.connInputSchema = value
	}
}

// submitNewConnection creates a new database connection from the input form
func (m Model) submitNewConnection() (Model, tea.Cmd) {
	// Validate inputs
	if m.connInputDriver == "" {
		m.connectionError = "Driver is required"
		return m, nil
	}

	if m.connInputHost == "" {
		m.connectionError = "Host is required"
		return m, nil
	}

	// Parse port
	var port int
	if m.connInputPort != "" {
		p, err := strconv.Atoi(m.connInputPort)
		if err != nil {
			m.connectionError = "Invalid port number"
			return m, nil
		}
		port = p
	} else {
		// Set default port based on driver
		switch strings.ToLower(m.connInputDriver) {
		case "postgres", "postgresql":
			port = 5432
		case "mariadb", "mysql":
			port = 3306
		case "clickhouse", "ch":
			port = 9000
		default:
			m.connectionError = "Unknown driver: " + m.connInputDriver
			return m, nil
		}
	}

	// Set default username if not provided
	username := m.connInputUsername
	if username == "" {
		switch strings.ToLower(m.connInputDriver) {
		case "postgres", "postgresql":
			username = "postgres"
		case "mariadb", "mysql":
			username = "root"
		case "clickhouse", "ch":
			username = "default"
		}
	}

	// Set default database if not provided
	database := m.connInputDatabase
	if database == "" {
		switch strings.ToLower(m.connInputDriver) {
		case "postgres", "postgresql":
			database = "postgres"
		case "mariadb", "mysql":
			database = "mysql"
		case "clickhouse", "ch":
			database = "default"
		}
	}

	// Set default schema for postgres
	schema := m.connInputSchema
	if schema == "" && (strings.ToLower(m.connInputDriver) == "postgres" || strings.ToLower(m.connInputDriver) == "postgresql") {
		schema = "public"
	}

	// Clear any previous error and close input form
	m.connectionError = ""
	m.connInputMode = false
	m.connectionStatus = Connecting

	// Create connection entry
	connEntry := connhistory.ConnectionEntry{
		Driver:   m.connInputDriver,
		Host:     m.connInputHost,
		Port:     port,
		Username: username,
		Password: m.connInputPassword,
		Database: database,
		Schema:   schema,
	}

	// Switch to the new connection
	return m, switchConnection(connEntry)
}
