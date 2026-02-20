package ui

import "github.com/charmbracelet/bubbles/key"

// GState represents the state for "gg" navigation.
type GState int

const (
	GStateIdle GState = iota
	GStateFirstG
)

// KeyMap defines all keybindings for nav mode.
type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	Select       key.Binding
	Back         key.Binding
	Top          key.Binding
	Bottom       key.Binding
	HalfPageDown key.Binding
	HalfPageUp   key.Binding
	Search       key.Binding
	Quit         key.Binding
	Help         key.Binding
	Add          key.Binding
	Edit         key.Binding
	Delete       key.Binding
	Restaurants  key.Binding
	Visits       key.Binding
	LogVisit     key.Binding
	NextColumn   key.Binding
	PrevColumn   key.Binding
	SortAsc      key.Binding
	SortDesc     key.Binding
	HideColumn   key.Binding
	ShowColumns  key.Binding
	FilterValue  key.Binding
	ClearFilter  key.Binding
	ColumnJump   key.Binding
	Undo         key.Binding
	Redo         key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "back"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "open"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("b", "esc"),
			key.WithHelp("b/esc", "back"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("gg", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "½ page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "½ page up"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Restaurants: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restaurants"),
		),
		Visits: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "visits"),
		),
		LogVisit: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "log visit"),
		),
		NextColumn: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next col"),
		),
		PrevColumn: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev col"),
		),
		SortAsc: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort asc"),
		),
		SortDesc: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "sort desc"),
		),
		HideColumn: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "hide col"),
		),
		ShowColumns: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "show cols"),
		),
		FilterValue: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "filter value"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "clear filter"),
		),
		ColumnJump: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "jump col"),
		),
		Undo: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "undo"),
		),
		Redo: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "redo"),
		),
	}
}

// FormKeyMap defines keybindings for insert/edit mode.
type FormKeyMap struct {
	NextField key.Binding
	PrevField key.Binding
	Save      key.Binding
	Cancel    key.Binding
}

// DefaultFormKeyMap returns the default form keybindings.
func DefaultFormKeyMap() FormKeyMap {
	return FormKeyMap{
		NextField: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		PrevField: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev field"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}
