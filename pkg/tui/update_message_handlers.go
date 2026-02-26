package tui

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/someshkoli/dessertfrog/pkg/connhistory"
	"github.com/someshkoli/dessertfrog/pkg/encryption"
)

// handleEncryptionSetupRequired handles encryption setup requirement
func (m Model) handleEncryptionSetupRequired(msg encryptionSetupRequiredMsg) (tea.Model, tea.Cmd) {
	// Show key selector popup for first-time setup
	m = m.debugLog("Encryption setup required - discovering keys")
	keys, err := encryption.DiscoverKeys()
	if err != nil {
		// If key discovery fails, show key selector with empty list
		// User can press 'g' to generate a new key
		m = m.debugLog("Key discovery failed: %v - showing key selector with empty list", err)
		keys = []encryption.Key{} // Empty list
	}
	m = m.debugLog("Found %d encryption keys", len(keys))
	m.availableKeys = keys
	m.filteredKeys = keys
	m.keySelectorMode = true
	m.keySelectorInsertMode = false
	m.keySelectorSelected = 0
	m.keySelectorScroll = 0
	return m, nil
}

// handleEncryptionSetupComplete handles completion of encryption setup
func (m Model) handleEncryptionSetupComplete(msg encryptionSetupCompleteMsg) (tea.Model, tea.Cmd) {
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
			m.passphraseInput = m.makePassphraseInput()
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
}

// handleEncryptionKeySelected handles user selecting an encryption key
func (m Model) handleEncryptionKeySelected(msg encryptionKeySelectedMsg) (tea.Model, tea.Cmd) {
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
			m.passphraseInput = m.makePassphraseInput()
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
}

// handleSSHKeyGeneration handles SSH key generation result
func (m Model) handleSSHKeyGeneration(msg sshKeyGenerationMsg) (tea.Model, tea.Cmd) {
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

		// Continue with database connection or show connection manager
		if m.driver != nil {
			m.connectionStatus = Connecting
			return m, connectToDatabase(m.driver)
		}
		// Open connection manager view
		m.connManagerMode = true
		if m.connHistory != nil {
			m.filteredConnections = m.connHistory.GetAll()
		}
		return m, nil
	} else {
		// SSH key generation failed
		m.connectionError = fmt.Sprintf("Failed to generate SSH key: %v", msg.err)
		return m, nil
	}
}

// handleEncryptionDisabled handles user disabling encryption
func (m Model) handleEncryptionDisabled(msg encryptionDisabledMsg) (tea.Model, tea.Cmd) {
	// User chose to disable encryption - initialize without encryption
	m = m.debugLog("User disabled encryption")
	m.encryptionConfig = &encryption.Config{
		DisableEncryption: true,
	}

	// Initialize connection history without encryption
	connHist, err := connhistory.NewHistory()
	if err != nil {
		m = m.debugLog(fmt.Sprintf("Failed to initialize connection history: %v", err))
		m.connectionError = fmt.Sprintf("Failed to initialize connection history: %v", err)
	}
	m.connHistory = connHist
	m = m.debugLog("Connection history initialized without encryption")

	// Continue with database connection or show connection manager
	if m.driver != nil {
		m.connectionStatus = Connecting
		return m, connectToDatabase(m.driver)
	}
	// Open connection manager view
	m.connManagerMode = true
	if m.connHistory != nil {
		m.filteredConnections = m.connHistory.GetAll()
	}
	return m, nil
}

// handleEncryptionSetupError handles encryption setup error
func (m Model) handleEncryptionSetupError(msg encryptionSetupErrorMsg) (tea.Model, tea.Cmd) {
	// Encryption setup failed, continue without encryption
	m.connectionError = fmt.Sprintf("Encryption setup failed: %v", msg.err)
	m.connectionStatus = Connecting
	return m, connectToDatabase(m.driver)
}

// handlePassphraseSubmitted handles passphrase submission
func (m Model) handlePassphraseSubmitted(msg passphraseSubmittedMsg) (tea.Model, tea.Cmd) {
	// User submitted passphrase - reinitialize connection history with passphrase
	if msg.keychainError != nil {
		m = m.debugLog("Warning: Failed to save passphrase to keychain: %v", msg.keychainError)
	} else {
		m = m.debugLog("Passphrase saved to keychain successfully")
	}

	m = m.debugLog("Attempting to initialize connection history with passphrase (len=%d)", len(msg.passphrase))
	connHist, err := connhistory.NewHistoryWithEncryptionAndPassphrase(m.encryptionKey, msg.passphrase)
	if err != nil {
		m.connectionError = fmt.Sprintf("Invalid passphrase or failed to initialize: %v", err)
		m = m.debugLog("Failed to load connection history with passphrase: %v", err)
		// Show passphrase prompt again
		m.passphrasePromptMode = true
		m.passphraseInput = m.makePassphraseInput()
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
}

// handleTablesLoaded handles successful tables loading
func (m Model) handleTablesLoaded(msg tablesLoadedMsg) (tea.Model, tea.Cmd) {
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
}

// handleTablesLoadFailed handles failed tables loading
func (m Model) handleTablesLoadFailed(msg tablesLoadFailedMsg) (tea.Model, tea.Cmd) {
	m.tablesLoading = false
	m.tablesError = msg.err.Error()
	return m, nil
}

// handleTableDataLoaded handles successful table data loading
func (m Model) handleTableDataLoaded(msg tableDataLoadedMsg) (tea.Model, tea.Cmd) {
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
}

// handleTableDataLoadFailed handles failed table data loading
func (m Model) handleTableDataLoadFailed(msg tableDataLoadFailedMsg) (tea.Model, tea.Cmd) {
	m.tableDataLoading = false
	m.tableDataError = msg.err.Error()
	return m, nil
}

// handleTableSchemaLoaded handles successful table schema loading
func (m Model) handleTableSchemaLoaded(msg tableSchemaLoadedMsg) (tea.Model, tea.Cmd) {
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
}

// handleTableSchemaLoadFailed handles failed table schema loading
func (m Model) handleTableSchemaLoadFailed(msg tableSchemaLoadFailedMsg) (tea.Model, tea.Cmd) {
	// Schema load failed - not critical, log and continue
	m = m.debugLog("Failed to load table schema: %v", msg.err)
	return m, nil
}

// handleSchemaInfoLoaded handles successful schema info loading
func (m Model) handleSchemaInfoLoaded(msg schemaInfoLoadedMsg) (tea.Model, tea.Cmd) {
	// Schema info loaded successfully for schema panel
	m.schemaInfoLoading = false
	m.schemaInfo = msg.schema
	// Update line count with new detailed schema
	m = m.updateSchemaPanelLineCount()
	return m, nil
}

// handleSchemaInfoLoadFailed handles failed schema info loading
func (m Model) handleSchemaInfoLoadFailed(msg schemaInfoLoadFailedMsg) (tea.Model, tea.Cmd) {
	// Schema info load failed - not critical
	m.schemaInfoLoading = false
	m = m.debugLog("Failed to load schema info: %v", msg.err)
	return m, nil
}

// handleSQLQueryResult handles successful SQL query execution
func (m Model) handleSQLQueryResult(msg sqlQueryResultMsg) (tea.Model, tea.Cmd) {
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
}

// handleSQLQueryFailed handles failed SQL query execution
func (m Model) handleSQLQueryFailed(msg sqlQueryFailedMsg) (tea.Model, tea.Cmd) {
	// SQL query failed - show query and complete error in table view
	m.tableDataLoading = false
	m.tableDataError = fmt.Sprintf("Query: %s\n\nError: %v", msg.query, msg.err)
	return m, nil
}

func (m Model) handleCellUpdateSuccess(msg cellUpdateSuccessMsg) (tea.Model, tea.Cmd) {
	m.cellEditMode = false
	m.cellEditCommandMode = false
	m.cellEditCommand = ""
	if m.cellEditRowIdx >= 0 && m.cellEditRowIdx < len(m.tableData) {
		if m.cellEditColIdx >= 0 && m.cellEditColIdx < len(m.tableData[m.cellEditRowIdx]) {
			m.tableData[m.cellEditRowIdx][m.cellEditColIdx] = msg.newValue
		}
	}
	m.tableDataError = ""
	return m, nil
}

func (m Model) handleCellUpdateFailed(msg cellUpdateFailedMsg) (tea.Model, tea.Cmd) {
	m.cellEditMode = false
	m.cellEditCommandMode = false
	m.cellEditCommand = ""
	m.tableDataError = fmt.Sprintf("Update failed: %v", msg.err)
	return m, nil
}

// handleRowsDeleteSuccess handles successful row deletion
func (m Model) handleRowsDeleteSuccess(msg rowsDeleteSuccessMsg) (tea.Model, tea.Cmd) {
	m = m.debugLog("Rows deleted successfully: %d rows affected", msg.rowsAffected)

	// Show success message
	m.clipboardMessage = fmt.Sprintf("✓ Deleted %d row(s)", msg.rowsAffected)

	// Refresh table data to show updated rows
	if m.currentViewTable != nil {
		m.tableDataLoading = true
		return m, tea.Batch(
			fetchTableData(m.driver, m.currentViewTable.SchemaName, m.currentViewTable.TableName, m.tableDataOffset),
			clearClipboardMessage(),
		)
	}

	return m, clearClipboardMessage()
}

// handleRowsDeleteFailed handles failed row deletion
func (m Model) handleRowsDeleteFailed(msg rowsDeleteFailedMsg) (tea.Model, tea.Cmd) {
	m = m.debugLog("Row deletion failed: %v", msg.err)

	// Show error message
	m.tableDataError = fmt.Sprintf("Delete failed: %v", msg.err)

	// Clear the deleted rows tracking on failure
	m.deletedRows = make(map[int]bool)
	m.deletedRowsCount = 0

	return m, nil
}

// handleClearClipboard handles clearing clipboard message
func (m Model) handleClearClipboard(msg clearClipboardMsgType) (tea.Model, tea.Cmd) {
	// Clear the clipboard notification message
	m.clipboardMessage = ""
	return m, nil
}

func (m Model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.sqlQueryInput.Width = msg.Width - 20
	return m, nil
}
