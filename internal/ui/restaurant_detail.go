package ui

import (
	"fmt"
	"strings"
	"toni/internal/model"
	"toni/internal/util"

	"github.com/charmbracelet/lipgloss"
)

// RestaurantDetailModel represents the restaurant detail screen.
type RestaurantDetailModel struct {
	detail model.RestaurantDetail
}

// NewRestaurantDetailModel creates a new restaurant detail model.
func NewRestaurantDetailModel(detail model.RestaurantDetail) *RestaurantDetailModel {
	return &RestaurantDetailModel{
		detail: detail,
	}
}

// View renders the restaurant detail.
func (m *RestaurantDetailModel) View(width, height int) string {
	r := m.detail.Restaurant

	// Keyboard shortcuts
	shortcuts := HelpDescStyle.Render("v add visit  e edit  h back")
	header := lipgloss.NewStyle().
		Width(width - 4).
		Align(lipgloss.Right).
		Render(shortcuts)

	var sections []string

	// Restaurant info
	var fields []string
	fields = append(fields, renderField("Name", r.Name))
	if r.Address != "" {
		fields = append(fields, renderField("Address", r.Address))
	}
	fields = append(fields, renderField("City", r.City))
	fields = append(fields, renderField("Neighborhood", r.Neighborhood))
	fields = append(fields, renderField("Cuisine", r.Cuisine))
	fields = append(fields, renderField("Price Range", r.PriceRange))

	// Visit count summary
	visitCountText := fmt.Sprintf("Visited %d times", len(m.detail.Visits))
	if len(m.detail.Visits) == 1 {
		visitCountText = "Visited once"
	} else if len(m.detail.Visits) == 0 {
		visitCountText = "Not visited yet"
	}
	fields = append(fields, LabelStyle.Render("Visits:")+" "+NormalRowStyle.Render(visitCountText))

	sections = append(sections, strings.Join(fields, "\n"))

	// Divider
	divider := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(strings.Repeat("─", width-8))
	sections = append(sections, divider)

	// Visits section
	if len(m.detail.Visits) > 0 {
		sections = append(sections, LabelStyle.Render("Visit History:"))
		sections = append(sections, m.renderVisitsTimeline(width))
	} else {
		sections = append(sections, HelpDescStyle.Render("No visits logged yet. Press 'v' to add one!"))
	}

	info := PanelStyle.
		Width(width - 4).
		Render(strings.Join(sections, "\n\n"))

	return lipgloss.JoinVertical(lipgloss.Left, header, info)
}

// renderVisitsTimeline renders visits as a compact inline timeline
func (m *RestaurantDetailModel) renderVisitsTimeline(width int) string {
	var entries []string

	for _, v := range m.detail.Visits {
		date := util.FormatDateHuman(v.VisitedOn)
		rating := util.FormatRating(v.Rating)

		entry := fmt.Sprintf("%s → %s", date, rating)
		entries = append(entries, entry)

		// Show max 5 in timeline, rest in table
		if len(entries) >= 5 {
			break
		}
	}

	timeline := strings.Join(entries, "  ·  ")

	// If more than 5 visits, show full table
	if len(m.detail.Visits) > 5 {
		timeline += "\n\n" + m.renderVisitsTable(width)
	}

	return NormalRowStyle.Render(timeline)
}

func (m *RestaurantDetailModel) renderVisitsTable(width int) string {
	dateWidth := 12
	ratingWidth := 10
	returnWidth := 10

	headerStyle := TableHeaderStyle.Bold(true)
	header := renderTableRow(
		[]string{formatHeaderLabel("date"), formatHeaderLabel("rating"), formatHeaderLabel("return")},
		[]int{dateWidth, ratingWidth, returnWidth},
		headerStyle,
	)
	divider := renderTableDivider([]int{dateWidth, ratingWidth, returnWidth})

	var rows []string
	for i, v := range m.detail.Visits {
		if i >= 5 { // Already showed first 5 in timeline
			style := NormalRowStyle

			ratingCell := util.FormatRating(v.Rating)
			if v.Rating != nil {
				ratingCell = lipgloss.NewStyle().Foreground(ColorYellow).Render(util.FormatRatingWithStar(v.Rating))
			}

			returnCell := util.FormatWouldReturnSymbol(v.WouldReturn)
			if v.WouldReturn != nil {
				color := ColorRed
				if *v.WouldReturn {
					color = ColorGreen
				}
				returnCell = lipgloss.NewStyle().Foreground(color).Render(returnCell)
			}

			cells := []string{
				util.FormatDateHuman(v.VisitedOn),
				ratingCell,
				returnCell,
			}
			rows = append(rows, renderTableRow(cells, []int{dateWidth, ratingWidth, returnWidth}, style))
		}
	}

	if len(rows) == 0 {
		return ""
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		divider,
		strings.Join(rows, "\n"),
	)
}
