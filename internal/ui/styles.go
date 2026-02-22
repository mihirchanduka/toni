package ui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	ColorBase       = lipgloss.AdaptiveColor{Light: "#F4F6F2", Dark: "#151A14"}
	ColorSurface    = lipgloss.AdaptiveColor{Light: "#E8EDE5", Dark: "#1E251D"}
	ColorSurfaceAlt = lipgloss.AdaptiveColor{Light: "#DCE4D7", Dark: "#273026"}
	ColorMuted      = lipgloss.AdaptiveColor{Light: "#6E7B65", Dark: "#A8B3A2"}
	ColorText       = lipgloss.AdaptiveColor{Light: "#243024", Dark: "#E8ECE5"}
	ColorAccent     = lipgloss.AdaptiveColor{Light: "#8FA082", Dark: "#A5B69A"}
	ColorOnAccent   = lipgloss.AdaptiveColor{Light: "#1B2818", Dark: "#102015"}
	ColorGreen      = lipgloss.AdaptiveColor{Light: "#7B9372", Dark: "#97B089"}
	ColorRed        = lipgloss.AdaptiveColor{Light: "#B8695D", Dark: "#D28A7D"}
	ColorYellow     = lipgloss.AdaptiveColor{Light: "#A4935D", Dark: "#CFC08A"}
)

// Styles
var (
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorBase)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			Padding(0, 1)

	HeaderBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted)

	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Bold(true).
				Padding(0, 1)

	TableSeparatorStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Faint(true)

	TableDividerStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Faint(true)

	SelectedRowStyle = lipgloss.NewStyle().
				Foreground(ColorOnAccent).
				Background(ColorAccent).
				Bold(true)

	NormalRowStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(ColorMuted)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccent)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Padding(0, 1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Padding(0, 1)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	InputStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorSurface).
			Padding(0, 1)

	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2)

	ActiveBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorAccent).
				Padding(1, 2)

	PanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2)

	BreadcrumbStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	BreadcrumbActiveStyle = lipgloss.NewStyle().
				Foreground(ColorAccent)

	EmptyStateStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true).
			Padding(2, 4)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)
)
