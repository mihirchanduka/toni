package ui

import (
	"fmt"
	"strings"
	"toni/internal/model"

	"github.com/charmbracelet/lipgloss"
)

// WantToVisitDetailModel represents the want_to_visit detail screen.
type WantToVisitDetailModel struct {
	entry      model.WantToVisit
	restaurant model.Restaurant
}

// NewWantToVisitDetailModel creates a new want_to_visit detail model.
func NewWantToVisitDetailModel(entry model.WantToVisit, restaurant model.Restaurant) *WantToVisitDetailModel {
	return &WantToVisitDetailModel{
		entry:      entry,
		restaurant: restaurant,
	}
}

// View renders the want_to_visit detail.
func (m *WantToVisitDetailModel) View(width, height int) string {
	// Keyboard shortcuts
	shortcuts := HelpDescStyle.Render("c mark visited  e edit  d delete  h back")
	header := lipgloss.NewStyle().
		Width(width - 4).
		Align(lipgloss.Right).
		Render(shortcuts)

	var sections []string

	// Restaurant info
	var fields []string
	fields = append(fields, renderField("Restaurant", m.restaurant.Name))
	if m.restaurant.Address != "" {
		fields = append(fields, renderField("Address", m.restaurant.Address))
	}
	fields = append(fields, renderField("City", m.restaurant.City))
	fields = append(fields, renderField("Neighborhood", m.restaurant.Neighborhood))
	fields = append(fields, renderField("Cuisine", m.restaurant.Cuisine))
	fields = append(fields, renderField("Price Range", m.restaurant.PriceRange))

	// Priority
	priorityText := "Not set"
	if m.entry.Priority != nil {
		priorityText = fmt.Sprintf("%d/5", *m.entry.Priority)
		color := ColorMuted
		if *m.entry.Priority >= 4 {
			color = ColorRed
		} else if *m.entry.Priority >= 3 {
			color = ColorYellow
		}
		priorityText = lipgloss.NewStyle().Foreground(color).Render(priorityText)
	}
	fields = append(fields, LabelStyle.Render("Priority:")+" "+priorityText)

	sections = append(sections, strings.Join(fields, "\n"))

	// Notes section
	if m.entry.Notes != "" {
		divider := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render(strings.Repeat("─", width-8))
		sections = append(sections, divider)
		sections = append(sections, LabelStyle.Render("Notes:"))
		sections = append(sections, NormalRowStyle.Render(m.entry.Notes))
	}

	info := PanelStyle.
		Width(width - 4).
		Render(strings.Join(sections, "\n\n"))

	return lipgloss.JoinVertical(lipgloss.Left, header, info)
}
