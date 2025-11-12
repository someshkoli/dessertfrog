package tui

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/someshkoli/dessertfrog/pkg/connhistory"
	"github.com/someshkoli/dessertfrog/pkg/encryption"
)

// Init initializes the bubbletea model
func (m Model) Init() tea.Cmd {
	// First check if encryption needs to be set up
	return checkEncryptionSetup()
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case connectionSuccessMsg:
		m.connectionStatus = Connected
		m.connectionError = ""
		m.tablesLoading = true
		// After successful connection, save to history and fetch tables
		if m.connHistory != nil {
			err := m.connHistory.Add(
				m.dbConfig.Driver,
				m.dbConfig.Host,
				m.dbConfig.Port,
				m.dbConfig.Username,
				m.dbConfig.Password,
				m.dbConfig.Database,
				m.dbConfig.Schema,
			)
			if err != nil {
				// Log error but don't fail connection
				m = m.debugLog(fmt.Sprintf("Failed to save connection to history: %v", err))
			} else {
				m = m.debugLog("Connection saved to history successfully")
			}
		} else {
			m = m.debugLog("Connection history is nil, cannot save connection")
		}
		return m, fetchTables(m.driver)

	case connectionFailedMsg:
		m.connectionStatus = ConnectionFailed
		m.connectionError = msg.err.Error()
		return m, nil

	case connectionSwitchSuccessMsg:
		// Close old driver
		if m.driver != nil {
			_ = m.driver.Close()
		}

		// Update to new connection
		m.driver = msg.driver
		m.dbConfig = msg.dbConfig
		m.sqlHistory = msg.sqlHistory
		m.connectionStatus = Connected
		m.connectionError = ""

		// Reset view state
		m.tableViewMode = false
		m.currentViewTable = nil
		m.tableData = nil
		m.tableColumns = nil
		m.selectedRow = 0
		m.scrollOffset = 0
		m.historyStack = make([]HistoryState, 0)
		m.historyIndex = -1

		// Save to connection history
		if m.connHistory != nil {
			err := m.connHistory.Add(
				msg.dbConfig.Driver,
				msg.dbConfig.Host,
				msg.dbConfig.Port,
				msg.dbConfig.Username,
				msg.dbConfig.Password,
				msg.dbConfig.Database,
				msg.dbConfig.Schema,
			)
			if err != nil {
				m = m.debugLog(fmt.Sprintf("Failed to save switched connection to history: %v", err))
			}
		}

		// Fetch tables for new connection
		m.tablesLoading = true
		return m, fetchTables(m.driver)

	case connectionSwitchFailedMsg:
		m.connectionError = fmt.Sprintf("Connection failed: %v", msg.err)
		m.connectionStatus = Disconnected
		// If we have connection input values, reopen the form to show error and allow retry
		if m.connInputDriver != "" || m.connInputHost != "" {
			m.connInputMode = true
			return m, nil
		}
		// Otherwise reopen connection manager to show error and allow retry
		m.connManagerMode = true
		if m.connHistory != nil {
			m.filteredConnections = m.connHistory.GetAll()
		}
		return m, nil

	// Encryption setup messages
	case encryptionSetupRequiredMsg:
		// Show key selector popup for first-time setup
		m = m.debugLog("Encryption setup required - discovering keys")
		keys, err := encryption.DiscoverKeys()
		if err != nil {
			// If key discovery fails, continue without encryption
			m = m.debugLog(fmt.Sprintf("Key discovery failed: %v", err))
			// If no database driver, show connection manager, otherwise connect
			if m.driver != nil {
				m.connectionStatus = Connecting
				return m, connectToDatabase(m.driver)
			} else {
				// Open connection manager view
				m.connManagerMode = true
				if m.connHistory != nil {
					m.filteredConnections = m.connHistory.GetAll()
				}
				return m, nil
			}
		}
		m = m.debugLog(fmt.Sprintf("Found %d encryption keys", len(keys)))
		m.availableKeys = keys
		m.filteredKeys = keys
		m.keySelectorMode = true
		m.keySelectorInsertMode = true
		m.keySelectorSelected = 0
		m.keySelectorScroll = 0
		return m, nil

	case encryptionSetupCompleteMsg:
		// Encryption is set up, save the key and continue with connection
		m = m.debugLog(fmt.Sprintf("Encryption setup complete with key: %s", msg.key.Name))
		m.encryptionKey = msg.key
		m.encryptionConfig = &encryption.Config{
			KeyPath: msg.key.Path,
			KeyType: msg.key.Type,
		}

		// Reinitialize connection history with encryption
		m = m.debugLog("Initializing connection history with encryption")
		connHist, err := connhistory.NewHistoryWithEncryption(m.encryptionKey)
		if err != nil {
			// Check if passphrase is required
			if errors.Is(err, encryption.ErrPassphraseRequired) {
				// Show passphrase prompt
				m.debugLog("Passphrase required - showing prompt")
				m.passphrasePromptMode = true
				m.passphraseInput = ""
				m.passphraseCursor = 0
				m.passphraseKeyName = msg.key.Name
				m.passphraseKeyPath = msg.key.Path
				return m, nil
			}
			// Other error - continue without encryption
			m = m.debugLog(fmt.Sprintf("Failed to load connection history: %v", err))
			m.connectionError = fmt.Sprintf("Failed to load connection history: %v", err)
		} else {
			m.connHistory = connHist
			m.debugLog("Connection history initialized without passphrase")
		}

		// Continue with database connection or show connection manager
		if m.driver != nil {
			m.connectionStatus = Connecting
			return m, connectToDatabase(m.driver)
		} else {
			// Open connection manager view
			m.connManagerMode = true
			if m.connHistory != nil {
				m.filteredConnections = m.connHistory.GetAll()
			}
			return m, nil
		}

	case encryptionKeySelectedMsg:
		// User selected an encryption key
		m = m.debugLog(fmt.Sprintf("User selected encryption key: %s", msg.key.Name))
		m.encryptionKey = msg.key
		m.encryptionConfig = &encryption.Config{
			KeyPath: msg.key.Path,
			KeyType: msg.key.Type,
		}

		// Reinitialize connection history with encryption
		m = m.debugLog("Initializing connection history with selected key")
		connHist, err := connhistory.NewHistoryWithEncryption(m.encryptionKey)
		if err != nil {
			// Check if passphrase is required
			if errors.Is(err, encryption.ErrPassphraseRequired) {
				// Show passphrase prompt
				m.debugLog("Passphrase required for selected key")
				m.passphrasePromptMode = true
				m.passphraseInput = ""
				m.passphraseCursor = 0
				m.passphraseKeyName = msg.key.Name
				m.passphraseKeyPath = msg.key.Path
				return m, nil
			}
			// Other error - continue without encryption
			m.debugLog(fmt.Sprintf("Failed to initialize with selected key: %v", err))
			m.connectionError = fmt.Sprintf("Failed to load connection history: %v", err)
		} else {
			m.connHistory = connHist
			m.debugLog("Connection history initialized with selected key")
		}

		// Continue with database connection or show connection manager
		if m.driver != nil {
			m.connectionStatus = Connecting
			return m, connectToDatabase(m.driver)
		} else {
			// Open connection manager view
			m.connManagerMode = true
			if m.connHistory != nil {
				m.filteredConnections = m.connHistory.GetAll()
			}
			return m, nil
		}

	case sshKeyGenerationMsg:
		if msg.success {
			// SSH key generated successfully, use it
			m.availableKeys = append(m.availableKeys, *msg.key)
			m.filteredKeys = append(m.filteredKeys, *msg.key)
			m.encryptionKey = msg.key
			m.encryptionConfig = &encryption.Config{
				KeyPath: msg.key.Path,
				KeyType: msg.key.Type,
			}

			// Save config
			if err := encryption.SaveConfig(m.encryptionConfig); err != nil {
				m.connectionError = fmt.Sprintf("Failed to save encryption config: %v", err)
			}

			// Reinitialize connection history with encryption
			connHist, err := connhistory.NewHistoryWithEncryption(m.encryptionKey)
			if err == nil {
				m.connHistory = connHist
			}

			// Close key selector and continue
			m.keySelectorMode = false
			m.connectionStatus = Connecting
			return m, connectToDatabase(m.driver)
		} else {
			// SSH key generation failed
			m.connectionError = fmt.Sprintf("Failed to generate SSH key: %v", msg.err)
			return m, nil
		}

	case encryptionSetupErrorMsg:
		// Encryption setup failed, continue without encryption
		m.connectionError = fmt.Sprintf("Encryption setup failed: %v", msg.err)
		m.connectionStatus = Connecting
		return m, connectToDatabase(m.driver)

	case passphraseSubmittedMsg:
		// User submitted passphrase - reinitialize connection history with passphrase
		if msg.keychainError != nil {
			m = m.debugLog(fmt.Sprintf("Warning: Failed to save passphrase to keychain: %v", msg.keychainError))
		} else {
			m = m.debugLog("Passphrase saved to keychain successfully")
		}

		m = m.debugLog(fmt.Sprintf("Attempting to initialize connection history with passphrase (len=%d)", len(msg.passphrase)))
		connHist, err := connhistory.NewHistoryWithEncryptionAndPassphrase(m.encryptionKey, msg.passphrase)
		if err != nil {
			m.connectionError = fmt.Sprintf("Invalid passphrase or failed to initialize: %v", err)
			m = m.debugLog(fmt.Sprintf("Failed to load connection history with passphrase: %v", err))
			// Show passphrase prompt again
			m.passphrasePromptMode = true
			m.passphraseInput = ""
			m.passphraseCursor = 0
			return m, nil
		} else {
			m.connHistory = connHist
			m = m.debugLog("Connection history initialized successfully with passphrase")
		}

		// Continue with database connection or show connection manager
		if m.driver != nil {
			m.connectionStatus = Connecting
			return m, connectToDatabase(m.driver)
		} else {
			// Open connection manager view
			m.connManagerMode = true
			if m.connHistory != nil {
				m.filteredConnections = m.connHistory.GetAll()
			}
			return m, nil
		}

	case tablesLoadedMsg:
		m.tablesLoading = false
		m.tables = msg.tables
		m.allEntities = msg.allEntities
		m.tablesError = ""
		m.selectedRow = 0
		m.scrollOffset = 0

		// Update schema panel line count for first table
		m = m.updateSchemaPanelLineCount()

		// Start loading detailed schema for first table
		if len(m.tables) > 0 {
			m.schemaInfoLoading = true
			m.schemaInfo = nil
			return m, fetchSchemaInfo(m.driver, m.tables[0].SchemaName, m.tables[0].TableName)
		}

		// Push initial home state to history
		m = m.pushHistory()

		return m, nil

	case tablesLoadFailedMsg:
		m.tablesLoading = false
		m.tablesError = msg.err.Error()
		return m, nil

	case tableDataLoadedMsg:
		m = m.debugLog("tableDataLoadedMsg received: rows=%d, cols=%d", len(msg.rows), len(msg.columns))
		m = m.debugLog("  isNavigatingHistory=%v", m.isNavigatingHistory)
		m = m.debugLog("  Before: historyIndex=%d, stack_size=%d", m.historyIndex, len(m.historyStack))

		m.tableDataLoading = false
		m.tableColumns = msg.columns
		m.tableData = msg.rows
		m.allTableData = msg.rows // Store unfiltered data
		m.tableDataError = ""
		m.tableContentFilter = "" // Clear filter on new data load
		m.queryTime = msg.queryTime
		m.fetchTime = msg.fetchTime
		// Reset scroll and selection when new data is loaded
		m.tableDataScrollX = 0
		m.tableDataScrollY = 0
		m.selectedDataRow = 0
		m.selectedDataCol = 0
		// Note: tableDataOffset is managed by pagination handlers, not reset here

		// Push to history only if this is not a history navigation
		if !m.isNavigatingHistory {
			m = m.debugLog("  Not navigating history, pushing new state")
			m = m.pushHistory()
		} else {
			// Clear the flag after handling history navigation
			m = m.debugLog("  Navigating history, NOT pushing, clearing flag")
			m.isNavigatingHistory = false
		}

		m = m.debugLog("  After: historyIndex=%d, stack_size=%d", m.historyIndex, len(m.historyStack))

		return m, nil

	case tableDataLoadFailedMsg:
		m.tableDataLoading = false
		m.tableDataError = msg.err.Error()
		return m, nil

	case tableSchemaLoadedMsg:
		// Update currentViewTable with full schema (includes primary key info)
		if m.currentViewTable != nil && msg.schema != nil {
			// Only update if it's the same table
			if m.currentViewTable.SchemaName == msg.schema.SchemaName &&
				m.currentViewTable.TableName == msg.schema.TableName {
				m.currentViewTable = msg.schema
				m = m.debugLog("Table schema loaded: %d columns, PK columns: %v",
					len(msg.schema.Columns), getPrimaryKeyNames(msg.schema))
			}
		}
		return m, nil

	case tableSchemaLoadFailedMsg:
		// Schema load failed - not critical, log and continue
		m = m.debugLog("Failed to load table schema: %v", msg.err)
		return m, nil

	case schemaInfoLoadedMsg:
		// Schema info loaded successfully for schema panel
		m.schemaInfoLoading = false
		m.schemaInfo = msg.schema
		// Update line count with new detailed schema
		m = m.updateSchemaPanelLineCount()
		return m, nil

	case schemaInfoLoadFailedMsg:
		// Schema info load failed - not critical
		m.schemaInfoLoading = false
		m = m.debugLog("Failed to load schema info: %v", msg.err)
		return m, nil

	case sqlQueryResultMsg:
		// SQL query executed successfully
		m.tableViewMode = true
		m.isCustomQuery = true
		m.executedSQLQuery = msg.query
		m.currentViewTable = nil
		m.tableColumns = msg.columns
		m.tableData = msg.rows
		m.allTableData = msg.rows // Store unfiltered data
		m.tableContentFilter = "" // Clear filter
		m.tableDataLoading = false
		m.tableDataError = ""
		m.tableDataScrollX = 0
		m.tableDataScrollY = 0
		m.selectedDataRow = 0
		m.selectedDataCol = 0
		m.tableDataOffset = 0
		m.queryTime = msg.queryTime
		m.fetchTime = msg.fetchTime

		// Add to SQL history
		if m.sqlHistory != nil {
			if err := m.sqlHistory.Add(msg.query); err != nil {
				m = m.debugLog("Failed to add query to history: %v", err)
			}
		}

		// Push to navigation history
		if !m.isNavigatingHistory {
			m = m.debugLog("  SQL query result, pushing to history")
			m = m.pushHistory()
		} else {
			m = m.debugLog("  Navigating history, NOT pushing, clearing flag")
			m.isNavigatingHistory = false
		}

		return m, nil

	case sqlQueryFailedMsg:
		// SQL query failed - show query and complete error in table view
		m.tableDataLoading = false
		m.tableDataError = fmt.Sprintf("Query: %s\n\nError: %v", msg.query, msg.err)
		return m, nil

	case cellUpdateSuccessMsg:
		// Cell update successful - close edit popup and stay on current page
		m.cellEditMode = false
		m.cellEditValue = ""
		m.cellEditCursor = 0
		m.cellEditCommandMode = false
		m.cellEditCommand = ""
		// Update the cell value in the current data without refreshing
		if m.cellEditRowIdx >= 0 && m.cellEditRowIdx < len(m.tableData) {
			if m.cellEditColIdx >= 0 && m.cellEditColIdx < len(m.tableData[m.cellEditRowIdx]) {
				m.tableData[m.cellEditRowIdx][m.cellEditColIdx] = msg.newValue
			}
		}
		// Clear any previous errors
		m.tableDataError = ""
		return m, nil

	case cellUpdateFailedMsg:
		// Cell update failed - show error in edit popup or as notification
		// For now, we'll show it in the table error field and close the popup
		m.cellEditMode = false
		m.cellEditValue = ""
		m.cellEditCursor = 0
		m.cellEditCommandMode = false
		m.cellEditCommand = ""
		m.tableDataError = fmt.Sprintf("Update failed: %v", msg.err)
		return m, nil

	case clearClipboardMsgType:
		// Clear the clipboard notification message
		m.clipboardMessage = ""
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}
