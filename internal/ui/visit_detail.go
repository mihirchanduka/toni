package ui

import (
	"strings"
	"toni/internal/model"
	"toni/internal/util"

	"github.com/charmbracelet/lipgloss"
)

// VisitDetailModel represents the visit detail screen.
type VisitDetailModel struct {
	visit      model.Visit
	restaurant model.Restaurant
}

// NewVisitDetailModel creates a new visit detail model.
func NewVisitDetailModel(visit model.Visit, restaurant model.Restaurant) *VisitDetailModel {
	return &VisitDetailModel{
		visit:      visit,
		restaurant: restaurant,
	}
}

// View renders the visit detail.
func (m *VisitDetailModel) View(width, height int) string {
	var sections []string

	// Keyboard shortcuts in top right corner
	shortcuts := HelpDescStyle.Render("e edit  h back")

	// Main info section
	var fields []string
	fields = append(fields, renderField("Restaurant", m.restaurant.Name))
	fields = append(fields, renderField("City", m.restaurant.City))
	fields = append(fields, renderField("Visited On", util.FormatDate(m.visit.VisitedOn)))

	// Rating with colored star
	ratingValue := util.FormatRating(m.visit.Rating)
	if m.visit.Rating != nil {
		ratingValue = lipgloss.NewStyle().Foreground(ColorYellow).Render(util.FormatRatingWithStar(m.visit.Rating))
	}
	fields = append(fields, LabelStyle.Render("Rating:")+" "+ratingValue)

	// Would Return with colored symbol
	returnValue := util.FormatWouldReturn(m.visit.WouldReturn)
	if m.visit.WouldReturn != nil {
		symbol := util.FormatWouldReturnSymbol(m.visit.WouldReturn)
		color := ColorRed
		if *m.visit.WouldReturn {
			color = ColorGreen
		}
		returnValue = lipgloss.NewStyle().Foreground(color).Render(symbol) + "  " + returnValue
	}
	fields = append(fields, LabelStyle.Render("Would Return?")+" "+returnValue)

	sections = append(sections, strings.Join(fields, "\n"))

	// Divider
	divider := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(strings.Repeat("─", width-8))
	sections = append(sections, divider)

	// Notes section
	if m.visit.Notes != "" {
		sections = append(sections, LabelStyle.Render("Notes:"))
		sections = append(sections, NormalRowStyle.Render(m.visit.Notes))
	} else {
		sections = append(sections, HelpDescStyle.Render("No notes for this visit"))
	}

	content := PanelStyle.
		Width(width - 4).
		Render(strings.Join(sections, "\n\n"))

	// Add shortcuts at the top
	header := lipgloss.NewStyle().
		Width(width - 4).
		Align(lipgloss.Right).
		Render(shortcuts)

	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

func renderField(label, value string) string {
	if value == "" {
		value = "—"
	}
	return LabelStyle.Render(label+":") + " " + NormalRowStyle.Render(value)
}
