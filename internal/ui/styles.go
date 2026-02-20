package ui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	ColorBase    = lipgloss.Color("#1D221E")
	ColorSurface = lipgloss.Color("#2A332C")
	ColorMuted   = lipgloss.Color("#7E8C80")
	ColorText    = lipgloss.Color("#D6E0D3")
	ColorAccent  = lipgloss.Color("#8FA082")
	ColorGreen   = lipgloss.Color("#a6e3a1")
	ColorRed     = lipgloss.Color("#f38ba8")
	ColorYellow  = lipgloss.Color("#f9e2af")
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
			Padding(0, 1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorMuted)

	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorAccent).
				Bold(true).
				Padding(0, 1).
				Background(ColorSurface)

	SelectedRowStyle = lipgloss.NewStyle().
				Foreground(ColorBase).
				Background(ColorAccent).
				Bold(false)

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
