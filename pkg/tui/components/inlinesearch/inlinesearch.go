// Package inlinesearch provides a reusable inline search bar component for Bubble Tea.
package inlinesearch

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Model is the inline search bar component.
type Model struct {
	input           textinput.Model
	ActiveStyle     lipgloss.Style
	InactiveStyle   lipgloss.Style
	SuggestionStyle lipgloss.Style
	focused         bool
	suggestion      string
}

// New creates a new inline search bar with the provided active and inactive styles.
func New(activeStyle, inactiveStyle lipgloss.Style) Model {
	ti := textinput.New()
	ti.Prompt = "Filter: "
	ti.CharLimit = 500
	ti.TextStyle = lipgloss.NewStyle()
	ti.PlaceholderStyle = lipgloss.NewStyle().Faint(true)

	return Model{
		input:           ti,
		ActiveStyle:     activeStyle,
		InactiveStyle:   inactiveStyle,
		SuggestionStyle: lipgloss.NewStyle().Faint(true),
	}
}

// Focus focuses the search input and switches to the active style.
func (m *Model) Focus() {
	m.focused = true
	m.input.Focus()
}

// Blur removes focus and switches to the inactive style.
func (m *Model) Blur() {
	m.focused = false
	m.input.Blur()
}

// Focused reports whether the component is focused.
func (m Model) Focused() bool {
	return m.focused
}

// Value returns the current search query text.
func (m Model) Value() string {
	return m.input.Value()
}

// SetValue sets the search query text.
func (m *Model) SetValue(s string) {
	m.input.SetValue(s)
}

// SetPrompt sets the prompt prefix shown before the input.
func (m *Model) SetPrompt(p string) {
	m.input.Prompt = p
}

// SetPlaceholder sets the placeholder text shown when the input is empty.
func (m *Model) SetPlaceholder(p string) {
	m.input.Placeholder = p
}

// SetWidth sets the display width of the text input field.
func (m *Model) SetWidth(w int) {
	m.input.Width = w
}

// CursorEnd moves the cursor to the end of the current text.
func (m *Model) CursorEnd() {
	m.input.CursorEnd()
}

// SetSuggestion sets the ghost-text autocomplete suggestion shown after the cursor.
func (m *Model) SetSuggestion(s string) {
	m.suggestion = s
}

// Update handles incoming Bubble Tea messages, forwarding them to the text input.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the component. Uses the active style when focused, inactive otherwise.
// Any pending suggestion is rendered after the input cursor in a faint style.
func (m Model) View() string {
	style := m.InactiveStyle
	if m.focused {
		style = m.ActiveStyle
	}
	content := m.input.View()
	if m.suggestion != "" {
		content += m.SuggestionStyle.Render(m.suggestion)
	}
	return style.Render(content)
}
