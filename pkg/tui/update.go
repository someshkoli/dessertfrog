package tui

import (
	tea "github.com/charmbracelet/bubbletea"
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
		return m.handleConnectionSuccess(msg)
	case connectionFailedMsg:
		return m.handleConnectionFailed(msg)
	case connectionSwitchSuccessMsg:
		return m.handleConnectionSwitchSuccess(msg)
	case connectionSwitchFailedMsg:
		return m.handleConnectionSwitchFailed(msg)
	case encryptionSetupRequiredMsg:
		return m.handleEncryptionSetupRequired(msg)
	case encryptionSetupCompleteMsg:
		return m.handleEncryptionSetupComplete(msg)
	case encryptionKeySelectedMsg:
		return m.handleEncryptionKeySelected(msg)
	case sshKeyGenerationMsg:
		return m.handleSSHKeyGeneration(msg)
	case encryptionDisabledMsg:
		return m.handleEncryptionDisabled(msg)
	case encryptionSetupErrorMsg:
		return m.handleEncryptionSetupError(msg)
	case passphraseSubmittedMsg:
		return m.handlePassphraseSubmitted(msg)
	case tablesLoadedMsg:
		return m.handleTablesLoaded(msg)
	case tablesLoadFailedMsg:
		return m.handleTablesLoadFailed(msg)
	case tableDataLoadedMsg:
		return m.handleTableDataLoaded(msg)
	case tableDataLoadFailedMsg:
		return m.handleTableDataLoadFailed(msg)
	case tableSchemaLoadedMsg:
		return m.handleTableSchemaLoaded(msg)
	case tableSchemaLoadFailedMsg:
		return m.handleTableSchemaLoadFailed(msg)
	case schemaInfoLoadedMsg:
		return m.handleSchemaInfoLoaded(msg)
	case schemaInfoLoadFailedMsg:
		return m.handleSchemaInfoLoadFailed(msg)
	case sqlQueryResultMsg:
		return m.handleSQLQueryResult(msg)
	case sqlQueryFailedMsg:
		return m.handleSQLQueryFailed(msg)
	case cellUpdateSuccessMsg:
		return m.handleCellUpdateSuccess(msg)
	case cellUpdateFailedMsg:
		return m.handleCellUpdateFailed(msg)
	case rowsDeleteSuccessMsg:
		return m.handleRowsDeleteSuccess(msg)
	case rowsDeleteFailedMsg:
		return m.handleRowsDeleteFailed(msg)
	case clearClipboardMsgType:
		return m.handleClearClipboard(msg)
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)
	}

	return m, nil
}
