package tui

import "github.com/charmbracelet/lipgloss"

// Styles for the TUI application
var (
	// Title and text styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	// Border styles
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#808080")).
			Padding(0, 1)

	activeBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	inactiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#4A4A4A")).
				Padding(0, 1)

	infoBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#808080")).
				Padding(0, 1)

	screenBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("#808080")).
				Padding(0, 1)

	statusBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#808080")).
			Padding(0, 1)

	// Connection status styles
	connectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	disconnectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)

	connectingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Bold(true)

	// Command mode styles
	commandLineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	// Table selection styles
	selectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)

	// Search popup styles
	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Background(lipgloss.Color("#1a1a1a"))

	searchInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#808080")).
				Padding(0, 1)

	activeSearchInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	inactiveSearchInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#4A4A4A")).
				Padding(0, 1)

	overlayStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#000000")).
			Foreground(lipgloss.Color("#FFFFFF"))

	// Ghost text style for autocomplete suggestions
	ghostTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Faint(true)

	// Status line styles
	statusLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2C2C2C")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 0).
			Inline(true)

	statusLineLeftStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA"))

	// Error box style
	errorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF0000")).
			Padding(2, 4).
			Foreground(lipgloss.Color("#FF6666")).
			Bold(true)
)
