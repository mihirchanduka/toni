package ui

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
	"toni/internal/db"
	"toni/internal/model"
	"toni/internal/search"
	"toni/internal/util"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type wtvAutocompleteResultMsg struct {
	seq     int
	results []search.Suggestion
	err     error
}

type wtvDebounceTick struct {
	seq int
}

// WantToVisitFormModel represents the want_to_visit form.
type WantToVisitFormModel struct {
	db             *sql.DB
	searchClient   *search.YelpClient
	wantToVisitID  int64
	restaurantID   int64
	focusedField   int
	inputs         []textinput.Model
	restaurantName string
	error          string

	// Autocomplete state
	searchSeq     int
	searchResults []search.Suggestion
	searchCursor  int
	showDropdown  bool
	searching     bool
	searchSpinner spinner.Model
}

// NewWantToVisitFormModel creates a new want_to_visit form.
func NewWantToVisitFormModel(database *sql.DB, yelpClient *search.YelpClient, restaurantID int64) *WantToVisitFormModel {
	inputs := make([]textinput.Model, 3)

	// Restaurant name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Restaurant name"
	inputs[0].Focus()
	inputs[0].CharLimit = 100

	// Priority
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "1-5 (5 = highest)"
	inputs[1].CharLimit = 1

	// Notes
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Why you want to visit..."
	inputs[2].CharLimit = 500

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := &WantToVisitFormModel{
		db:            database,
		searchClient:  yelpClient,
		restaurantID:  restaurantID,
		focusedField:  0,
		inputs:        inputs,
		searchSpinner: sp,
	}

	// If restaurant ID is provided, load the name
	if restaurantID > 0 {
		restaurant, err := db.GetRestaurant(database, restaurantID)
		if err == nil {
			m.restaurantName = restaurant.Name
			m.inputs[0].SetValue(restaurant.Name)
			m.focusedField = 1
			m.inputs[0].Blur()
			m.inputs[1].Focus()
		}
	}

	return m
}

// LoadWantToVisit loads an existing want_to_visit for editing.
func (m *WantToVisitFormModel) LoadWantToVisit(wtv model.WantToVisit) {
	m.wantToVisitID = wtv.ID
	m.restaurantID = wtv.RestaurantID

	// Load restaurant name
	restaurant, err := db.GetRestaurant(m.db, wtv.RestaurantID)
	if err == nil {
		m.restaurantName = restaurant.Name
		m.inputs[0].SetValue(restaurant.Name)
	}

	if wtv.Priority != nil {
		m.inputs[1].SetValue(strconv.Itoa(*wtv.Priority))
	}
	m.inputs[2].SetValue(wtv.Notes)
}

// Update handles input.
func (m WantToVisitFormModel) Update(msg tea.Msg) (WantToVisitFormModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case wtvDebounceTick:
		if msg.seq == m.searchSeq && m.searchClient != nil {
			query := m.inputs[0].Value()
			return m, m.doSearch(query, msg.seq)
		}
		return m, nil
	case wtvAutocompleteResultMsg:
		if msg.seq == m.searchSeq {
			m.searching = false
			if msg.err != nil {
				m.error = fmt.Sprintf("Search error: %v", msg.err)
				m.showDropdown = false
			} else {
				m.error = ""
				m.searchResults = msg.results
				m.searchCursor = 0
				m.showDropdown = len(msg.results) > 0
			}
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.searchSpinner, cmd = m.searchSpinner.Update(msg)
		return m, cmd
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if m.showDropdown && m.focusedField == 0 {
		switch keyMsg.String() {
		case "esc":
			m.showDropdown = false
			return m, nil
		case "j", "down":
			if m.searchCursor < len(m.searchResults)-1 {
				m.searchCursor++
			}
			return m, nil
		case "k", "up":
			if m.searchCursor > 0 {
				m.searchCursor--
			}
			return m, nil
		case "enter", "tab":
			if m.searchCursor < len(m.searchResults) {
				m.selectSuggestion(m.searchResults[m.searchCursor])
				m.showDropdown = false
				m.nextField()
			}
			return m, nil
		}
	}

	switch keyMsg.String() {
	case "esc":
		if m.showDropdown {
			m.showDropdown = false
			return m, nil
		}
		return m, func() tea.Msg {
			return model.FormCancelledMsg{}
		}
	case "ctrl+s":
		return m, m.save()
	case "tab":
		m.nextField()
		return m, nil
	case "shift+tab":
		m.showDropdown = false
		m.prevField()
		return m, nil
	}

	// Update current input
	var cmd tea.Cmd
	m.inputs[m.focusedField], cmd = m.inputs[m.focusedField].Update(keyMsg)
	cmds = append(cmds, cmd)

	// If restaurant text changed from selected/prefilled name, clear stale ID.
	if m.focusedField == 0 {
		currentName := strings.TrimSpace(m.inputs[0].Value())
		if currentName != m.restaurantName {
			m.restaurantID = 0
		}
	}

	// Trigger autocomplete on restaurant field changes.
	if m.focusedField == 0 && m.searchClient != nil {
		query := m.inputs[0].Value()
		if len(query) >= 2 {
			m.searchSeq++
			seq := m.searchSeq
			m.searching = true
			m.showDropdown = false
			cmds = append(cmds, m.searchSpinner.Tick)
			cmds = append(cmds, tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
				return wtvDebounceTick{seq: seq}
			}))
		} else {
			m.showDropdown = false
			m.searchResults = nil
			m.searching = false
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the form.
func (m *WantToVisitFormModel) View(width, height int) string {
	var fields []string

	restaurantField := renderFormField("Restaurant *", m.inputs[0], m.focusedField == 0)
	if m.showDropdown && len(m.searchResults) > 0 {
		dropdown := m.renderDropdown(width - 8)
		restaurantField = lipgloss.JoinVertical(lipgloss.Left, restaurantField, dropdown)
	} else if m.searching && m.focusedField == 0 {
		searching := HelpDescStyle.Render(m.searchSpinner.View() + " Searching...")
		restaurantField = lipgloss.JoinVertical(lipgloss.Left, restaurantField, searching)
	}
	fields = append(fields, restaurantField)
	fields = append(fields, renderFormField("Priority (1-5)", m.inputs[1], m.focusedField == 1))
	fields = append(fields, renderFormField("Notes", m.inputs[2], m.focusedField == 2))

	if m.error != "" {
		fields = append(fields, "")
		fields = append(fields, ErrorStyle.Render(m.error))
	}

	formContent := strings.Join(fields, "\n\n")

	content := PanelStyle.
		Width(width - 4).
		Height(height - 4).
		Render(formContent)

	return content
}

func (m *WantToVisitFormModel) doSearch(query string, seq int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		results, err := m.searchClient.Autocomplete(ctx, query, "")
		return wtvAutocompleteResultMsg{seq: seq, results: results, err: err}
	}
}

func (m *WantToVisitFormModel) selectSuggestion(suggestion search.Suggestion) {
	m.inputs[0].SetValue(suggestion.Name)
	m.restaurantName = suggestion.Name

	restaurants, err := db.SearchRestaurants(m.db, suggestion.Name)
	if err == nil {
		if existingID, ok := findExactRestaurantID(restaurants, suggestion.Name); ok {
			m.restaurantID = existingID
			return
		}
	}

	newRest := model.NewRestaurant{
		Name:         suggestion.Name,
		Address:      suggestion.Address,
		City:         suggestion.City,
		Neighborhood: suggestion.Neighborhood,
		Cuisine:      suggestion.Cuisine,
		PriceRange:   suggestion.PriceRange,
		Latitude:     &suggestion.Latitude,
		Longitude:    &suggestion.Longitude,
		PlaceID:      suggestion.PlaceID,
	}
	id, err := db.InsertRestaurant(m.db, newRest)
	if err == nil {
		m.restaurantID = id
		m.restaurantName = suggestion.Name
	}
}

func (m *WantToVisitFormModel) renderDropdown(width int) string {
	var items []string

	for i, result := range m.searchResults {
		style := NormalRowStyle
		if i == m.searchCursor {
			style = SelectedRowStyle
		}

		left := util.TruncateString(result.Name, 40)
		if result.City != "" {
			left += "  ·  " + result.City
		}

		right := ""
		if result.Cuisine != "" {
			right = HelpDescStyle.Render(result.Cuisine)
		}

		availableWidth := width - 4
		padding := max(0, availableWidth-lipgloss.Width(left)-lipgloss.Width(right))
		line := style.Width(availableWidth).Render(left + strings.Repeat(" ", padding) + right)
		items = append(items, line)
	}

	if len(items) == 0 {
		items = append(items, HelpDescStyle.Render("No results"))
	}

	return BorderStyle.
		Width(width).
		Render(strings.Join(items, "\n"))
}

func (m *WantToVisitFormModel) nextField() {
	m.inputs[m.focusedField].Blur()
	m.focusedField = (m.focusedField + 1) % len(m.inputs)
	m.inputs[m.focusedField].Focus()
}

func (m *WantToVisitFormModel) prevField() {
	m.inputs[m.focusedField].Blur()
	m.focusedField--
	if m.focusedField < 0 {
		m.focusedField = len(m.inputs) - 1
	}
	m.inputs[m.focusedField].Focus()
}

func (m *WantToVisitFormModel) save() tea.Cmd {
	return func() tea.Msg {
		// Validate and parse inputs
		restaurantName := strings.TrimSpace(m.inputs[0].Value())
		if restaurantName == "" {
			return model.ErrorMsg{Err: fmt.Errorf("restaurant name is required")}
		}

		// Find or create restaurant
		var restaurantID int64
		if m.restaurantID > 0 {
			restaurantID = m.restaurantID
		} else {
			restaurants, err := db.SearchRestaurants(m.db, restaurantName)
			if err != nil {
				return model.ErrorMsg{Err: err}
			}

			if existingID, ok := findExactRestaurantID(restaurants, restaurantName); ok {
				restaurantID = existingID
			} else {
				// Create new restaurant
				id, err := db.InsertRestaurant(m.db, model.NewRestaurant{
					Name: restaurantName,
				})
				if err != nil {
					return model.ErrorMsg{Err: err}
				}
				restaurantID = id
			}
		}

		var priority *int
		priorityStr := strings.TrimSpace(m.inputs[1].Value())
		if priorityStr != "" {
			p, err := strconv.Atoi(priorityStr)
			if err != nil || p < 1 || p > 5 {
				return model.ErrorMsg{Err: fmt.Errorf("priority must be between 1 and 5")}
			}
			priority = &p
		}

		notes := strings.TrimSpace(m.inputs[2].Value())

		// Save
		if m.wantToVisitID > 0 {
			before, err := db.GetWantToVisit(m.db, m.wantToVisitID)
			if err != nil {
				return model.ErrorMsg{Err: err}
			}
			err = db.UpdateWantToVisit(m.db, model.UpdateWantToVisit{
				ID:           m.wantToVisitID,
				RestaurantID: restaurantID,
				Notes:        notes,
				Priority:     priority,
			})
			if err != nil {
				return model.ErrorMsg{Err: err}
			}
			return model.WantToVisitSavedMsg{
				ID:        m.wantToVisitID,
				Operation: "update",
				Before:    &before,
				After: model.WantToVisit{
					ID:           m.wantToVisitID,
					RestaurantID: restaurantID,
					Notes:        notes,
					Priority:     priority,
					CreatedAt:    before.CreatedAt,
				},
			}
		} else {
			id, err := db.InsertWantToVisit(m.db, model.NewWantToVisit{
				RestaurantID: restaurantID,
				Notes:        notes,
				Priority:     priority,
			})
			if err != nil {
				return model.ErrorMsg{Err: err}
			}
			return model.WantToVisitSavedMsg{
				ID:        id,
				Operation: "insert",
				After: model.WantToVisit{
					ID:           id,
					RestaurantID: restaurantID,
					Notes:        notes,
					Priority:     priority,
				},
			}
		}
	}
}
