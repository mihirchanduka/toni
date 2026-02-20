package ui

import (
	"strings"
	"toni/internal/model"

	"github.com/charmbracelet/lipgloss"
)

// RenderHelp renders context-sensitive help footer.
func RenderHelp(screen model.Screen, mode model.Mode, width int) string {
	if mode == model.ModeInsert {
		return renderFormHelp(width)
	}

	switch screen {
	case model.ScreenVisits:
		return renderVisitsHelp(width)
	case model.ScreenRestaurants:
		return renderRestaurantsHelp(width)
	case model.ScreenWantToVisit:
		return renderWantToVisitHelp(width)
	case model.ScreenVisitDetail:
		return renderVisitDetailHelp(width)
	case model.ScreenRestaurantDetail:
		return renderRestaurantDetailHelp(width)
	case model.ScreenWantToVisitDetail:
		return renderWantToVisitDetailHelp(width)
	default:
		return renderDefaultHelp(width)
	}
}

func renderVisitsHelp(width int) string {
	keys := []string{
		helpKey("j/k", "navigate"),
		helpKey("tab", "next col"),
		helpKey("s/S", "sort"),
		helpKey("n/N", "filter"),
		helpKey("a", "add visit"),
		helpKey("r", "restaurants"),
		helpKey("w", "want to visit"),
		helpKey("enter", "details"),
		helpKey("/", "jump col"),
	}
	return renderHelpLine(keys, width)
}

func renderRestaurantsHelp(width int) string {
	keys := []string{
		helpKey("j/k", "navigate"),
		helpKey("tab", "next col"),
		helpKey("s/S", "sort"),
		helpKey("c/C", "hide/show col"),
		helpKey("a", "add restaurant"),
		helpKey("v", "log visit"),
		helpKey("w", "want to visit"),
		helpKey("enter", "details"),
		helpKey("u/ctrl+r", "undo/redo"),
	}
	return renderHelpLine(keys, width)
}

func renderWantToVisitHelp(width int) string {
	keys := []string{
		helpKey("j/k", "navigate"),
		helpKey("tab", "next col"),
		helpKey("s/S", "sort"),
		helpKey("n/N", "filter"),
		helpKey("a", "add place"),
		helpKey("v", "visits"),
		helpKey("r", "restaurants"),
		helpKey("enter", "details"),
		helpKey("/", "jump col"),
	}
	return renderHelpLine(keys, width)
}

func renderWantToVisitDetailHelp(width int) string {
	keys := []string{
		helpKey("h/esc", "back"),
		helpKey("c", "mark visited"),
		helpKey("e", "edit"),
		helpKey("d", "delete"),
	}
	return renderHelpLine(keys, width)
}

func renderVisitDetailHelp(width int) string {
	keys := []string{
		helpKey("h/esc", "back"),
		helpKey("e", "edit"),
		helpKey("d", "delete"),
	}
	return renderHelpLine(keys, width)
}

func renderRestaurantDetailHelp(width int) string {
	keys := []string{
		helpKey("h/esc", "back"),
		helpKey("v", "add visit"),
		helpKey("e", "edit"),
		helpKey("d", "delete"),
	}
	return renderHelpLine(keys, width)
}

func renderFormHelp(width int) string {
	keys := []string{
		helpKey("tab", "next field"),
		helpKey("shift+tab", "prev field"),
		helpKey("ctrl+s", "save"),
		helpKey("esc", "cancel"),
	}
	return renderHelpLine(keys, width)
}

func renderDefaultHelp(width int) string {
	keys := []string{
		helpKey("j/k", "navigate"),
		helpKey("h/l", "back/select"),
		helpKey("q", "quit"),
	}
	return renderHelpLine(keys, width)
}

func helpKey(key, desc string) string {
	return HelpKeyStyle.Render(key) + " " + HelpDescStyle.Render(desc)
}

func renderHelpLine(keys []string, width int) string {
	line := strings.Join(keys, "  ")
	return FooterStyle.Width(width).Render(line)
}

// RenderFullHelp renders the full help screen.
func RenderFullHelp(width, height int) string {
	content := lipgloss.NewStyle().
		Width(width-4).
		Height(height-6).
		Padding(1, 2)

	sections := []string{
		titleSection("Navigation (Nav Mode)"),
		helpSection([]helpItem{
			{"j / ↓", "Move down"},
			{"k / ↑", "Move up"},
			{"h / ← / b", "Go back / parent"},
			{"l / → / enter", "Open / select"},
			{"tab / shift+tab", "Cycle active column"},
			{"/ then 1-9", "Jump to column"},
			{"s / S", "Sort active column asc/desc"},
			{"c / C", "Hide active column / show all"},
			{"n / N", "Filter by selected value / clear"},
			{"gg", "Jump to top"},
			{"G", "Jump to bottom"},
			{"ctrl+d", "Half page down"},
			{"ctrl+u", "Half page up"},
			{"u / ctrl+r", "Undo / redo"},
			{"esc", "Cancel / close"},
			{"q", "Quit (from top-level)"},
			{"?", "Toggle help"},
		}),
		titleSection("Visits Screen"),
		helpSection([]helpItem{
			{"a", "Quick-add visit"},
			{"r", "Go to restaurants"},
			{"w", "Go to want to visit"},
			{"enter / l", "Open visit detail"},
		}),
		titleSection("Restaurants Screen"),
		helpSection([]helpItem{
			{"a", "Add restaurant"},
			{"v", "Log visit for selected"},
			{"w", "Go to want to visit"},
			{"enter / l", "Open restaurant detail"},
			{"b / h", "Back to visits"},
		}),
		titleSection("Want to Visit Screen"),
		helpSection([]helpItem{
			{"a", "Add place to list"},
			{"c", "Mark as visited (from detail)"},
			{"v", "Go to visits"},
			{"r", "Go to restaurants"},
			{"enter / l", "Open detail"},
		}),
		titleSection("Forms (Insert/Edit Mode)"),
		helpSection([]helpItem{
			{"tab", "Next field"},
			{"shift+tab", "Previous field"},
			{"ctrl+s", "Save"},
			{"esc", "Cancel"},
		}),
	}

	helpText := content.Render(strings.Join(sections, "\n\n"))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		TitleStyle.Width(width).Render("Help"),
		helpText,
		FooterStyle.Width(width).Render(HelpKeyStyle.Render("esc")+" "+HelpDescStyle.Render("close help")),
	)
}

type helpItem struct {
	key  string
	desc string
}

func titleSection(title string) string {
	return LabelStyle.Render(title)
}

func helpSection(items []helpItem) string {
	var lines []string
	for _, item := range items {
		lines = append(lines, "  "+HelpKeyStyle.Render(item.key)+" - "+HelpDescStyle.Render(item.desc))
	}
	return strings.Join(lines, "\n")
}
