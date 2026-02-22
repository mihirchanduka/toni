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

// Message types for autocomplete
type autocompleteResultMsg struct {
	seq     int
	results []search.Suggestion
	err     error
}

type debounceTick struct {
	seq int
}

// VisitFormModel represents the visit form.
type VisitFormModel struct {
	db             *sql.DB
	searchClient   *search.YelpClient
	visitID        int64
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

// NewVisitFormModel creates a new visit form.
func NewVisitFormModel(database *sql.DB, yelpClient *search.YelpClient, restaurantID int64) *VisitFormModel {
	inputs := make([]textinput.Model, 5)

	// Restaurant name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Search restaurant..."
	inputs[0].Focus()
	inputs[0].CharLimit = 100

	// Date
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "June 20, 2025 (optional)"
	inputs[1].CharLimit = 32

	// Rating
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "1-10 (decimals ok)"
	inputs[2].CharLimit = 4

	// Would Return
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "y/n"
	inputs[3].CharLimit = 1

	// Notes
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Your notes..."
	inputs[4].CharLimit = 500

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := &VisitFormModel{
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

// LoadVisit loads an existing visit for editing.
func (m *VisitFormModel) LoadVisit(visit model.Visit) {
	m.visitID = visit.ID
	m.restaurantID = visit.RestaurantID

	// Load restaurant name
	restaurant, err := db.GetRestaurant(m.db, visit.RestaurantID)
	if err == nil {
		m.restaurantName = restaurant.Name
		m.inputs[0].SetValue(restaurant.Name)
	}

	m.inputs[1].SetValue(visit.VisitedOn)
	if visit.VisitedOn != "" {
		if t, err := time.Parse("2006-01-02", visit.VisitedOn); err == nil {
			m.inputs[1].SetValue(t.Format("January 2, 2006"))
		}
	}
	if visit.Rating != nil {
		m.inputs[2].SetValue(strconv.FormatFloat(*visit.Rating, 'f', -1, 64))
	}
	if visit.WouldReturn != nil {
		if *visit.WouldReturn {
			m.inputs[3].SetValue("y")
		} else {
			m.inputs[3].SetValue("n")
		}
	}
	m.inputs[4].SetValue(visit.Notes)
}

// Update handles all messages.
func (m VisitFormModel) Update(msg tea.Msg) (VisitFormModel, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle async messages first
	switch msg := msg.(type) {
	case debounceTick:
		if msg.seq == m.searchSeq && m.searchClient != nil {
			query := m.inputs[0].Value()
			return m, m.doSearch(query, msg.seq)
		}
		return m, nil
	case autocompleteResultMsg:
		if msg.seq == m.searchSeq {
			m.searching = false
			if msg.err != nil {
				m.error = fmt.Sprintf("Search error: %v", msg.err)
				m.showDropdown = false
			} else {
				m.error = "" // Clear any previous errors
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

	// Handle keyboard input
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Handle dropdown navigation when visible
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

	// Handle form navigation
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
		if !m.showDropdown {
			m.nextField()
			return m, nil
		}
	case "shift+tab":
		m.showDropdown = false
		m.prevField()
		return m, nil
	}

	// Update current input
	var cmd tea.Cmd
	m.inputs[m.focusedField], cmd = m.inputs[m.focusedField].Update(keyMsg)
	cmds = append(cmds, cmd)

	// If restaurant text changed from the selected/prefilled name, clear stale ID.
	if m.focusedField == 0 {
		currentName := strings.TrimSpace(m.inputs[0].Value())
		if currentName != m.restaurantName {
			m.restaurantID = 0
		}
	}

	// Trigger autocomplete on restaurant field changes
	if m.focusedField == 0 && m.searchClient != nil {
		query := m.inputs[0].Value()
		if len(query) >= 2 {
			m.searchSeq++
			seq := m.searchSeq
			m.searching = true
			m.showDropdown = false // Clear dropdown while searching

			// Start spinner and debounce
			cmds = append(cmds, m.searchSpinner.Tick)
			cmds = append(cmds, tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
				return debounceTick{seq: seq}
			}))
		} else {
			m.showDropdown = false
			m.searchResults = nil
			m.searching = false
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *VisitFormModel) doSearch(query string, seq int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		results, err := m.searchClient.Autocomplete(ctx, query, "")
		return autocompleteResultMsg{seq: seq, results: results, err: err}
	}
}

func (m *VisitFormModel) selectSuggestion(suggestion search.Suggestion) {
	m.inputs[0].SetValue(suggestion.Name)
	m.restaurantName = suggestion.Name

	// Try to find existing restaurant or create with autocomplete data
	restaurants, err := db.SearchRestaurants(m.db, suggestion.Name)
	if err == nil {
		if existingID, ok := findExactRestaurantID(restaurants, suggestion.Name); ok {
			// Restaurant already exists, use its ID
			m.restaurantID = existingID
			return
		}
	}

	{
		// Create new restaurant with autocomplete data including location
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
}

// View renders the form.
func (m *VisitFormModel) View(width, height int) string {
	var fields []string

	useSearchSidebar := m.shouldUseSearchSidebar(width)

	// Restaurant field
	restaurantField := renderFormField("Restaurant *", m.inputs[0], m.focusedField == 0)
	if !useSearchSidebar && m.showDropdown && len(m.searchResults) > 0 {
		dropdown := m.renderDropdown(width - 8)
		restaurantField = lipgloss.JoinVertical(lipgloss.Left, restaurantField, dropdown)
	} else if !useSearchSidebar && m.searching && m.focusedField == 0 {
		searching := HelpDescStyle.Render(m.searchSpinner.View() + " Searching...")
		restaurantField = lipgloss.JoinVertical(lipgloss.Left, restaurantField, searching)
	}
	fields = append(fields, restaurantField)

	fields = append(fields, renderFormField("Visit Date (optional)", m.inputs[1], m.focusedField == 1))
	fields = append(fields, renderFormField("Rating (1-10, optional)", m.inputs[2], m.focusedField == 2))
	fields = append(fields, renderFormField("Would Return? (y/n)", m.inputs[3], m.focusedField == 3))
	fields = append(fields, renderFormField("Notes", m.inputs[4], m.focusedField == 4))

	if m.error != "" {
		fields = append(fields, "")
		fields = append(fields, ErrorStyle.Render(m.error))
	}

	formContent := strings.Join(fields, "\n\n")
	if useSearchSidebar {
		leftWidth := max(44, (width-14)*55/100)
		rightWidth := max(30, (width-14)-leftWidth)
		left := lipgloss.NewStyle().Width(leftWidth).Render(formContent)
		right := m.renderSearchSidebar(rightWidth, height-10)
		formContent = lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	}

	content := PanelStyle.
		Width(width - 4).
		Height(height - 4).
		Render(formContent)

	return content
}

func (m *VisitFormModel) renderDropdown(width int) string {
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

func (m *VisitFormModel) shouldUseSearchSidebar(width int) bool {
	if width < 110 || m.focusedField != 0 {
		return false
	}
	return m.showDropdown || m.searching
}

func (m *VisitFormModel) renderSearchSidebar(width, availableHeight int) string {
	title := LabelStyle.Render("Search Results")

	if m.searching {
		body := HelpDescStyle.Render(m.searchSpinner.View() + " Searching...")
		return BorderStyle.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))
	}

	if len(m.searchResults) == 0 {
		body := HelpDescStyle.Render("Type at least 2 characters to search.")
		return BorderStyle.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))
	}

	maxRows := min(12, max(5, availableHeight/2))
	var items []string
	for i := 0; i < len(m.searchResults) && i < maxRows; i++ {
		result := m.searchResults[i]
		style := NormalRowStyle
		if i == m.searchCursor {
			style = SelectedRowStyle
		}

		left := util.TruncateString(result.Name, max(10, width-14))
		if result.City != "" {
			left += "  ·  " + result.City
		}

		right := ""
		if result.Cuisine != "" {
			right = HelpDescStyle.Render(result.Cuisine)
		}

		lineWidth := max(10, width-8)
		padding := max(0, lineWidth-lipgloss.Width(left)-lipgloss.Width(right))
		line := style.Width(lineWidth).Render(left + strings.Repeat(" ", padding) + right)
		items = append(items, line)
	}

	help := HelpDescStyle.Render("j/k move  enter/tab select  esc close")
	body := lipgloss.JoinVertical(lipgloss.Left, items...)
	return BorderStyle.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", help))
}

func (m *VisitFormModel) nextField() {
	m.inputs[m.focusedField].Blur()
	m.focusedField = (m.focusedField + 1) % len(m.inputs)
	m.inputs[m.focusedField].Focus()
	m.showDropdown = false
}

func (m *VisitFormModel) prevField() {
	m.inputs[m.focusedField].Blur()
	m.focusedField--
	if m.focusedField < 0 {
		m.focusedField = len(m.inputs) - 1
	}
	m.inputs[m.focusedField].Focus()
}

func (m *VisitFormModel) save() tea.Cmd {
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

		dateInput := strings.TrimSpace(m.inputs[1].Value())
		date, err := util.ParseVisitDateInput(dateInput)
		if err != nil {
			return model.ErrorMsg{Err: fmt.Errorf("invalid date format (e.g. June 20, 2025)")}
		}

		var rating *float64
		ratingStr := strings.TrimSpace(m.inputs[2].Value())
		if ratingStr != "" {
			r, err := strconv.ParseFloat(ratingStr, 64)
			if err != nil || r < 1 || r > 10 {
				return model.ErrorMsg{Err: fmt.Errorf("rating must be between 1 and 10 (decimals allowed)")}
			}
			rating = &r
		}

		var wouldReturn *bool
		wrStr := strings.TrimSpace(strings.ToLower(m.inputs[3].Value()))
		if wrStr != "" {
			var wr bool
			switch wrStr {
			case "y", "yes", "1":
				wr = true
			case "n", "no", "0":
				wr = false
			default:
				return model.ErrorMsg{Err: fmt.Errorf("would return must be y/yes/1 or n/no/0")}
			}
			wouldReturn = &wr
		}

		notes := strings.TrimSpace(m.inputs[4].Value())

		// Save
		if m.visitID > 0 {
			before, err := db.GetVisit(m.db, m.visitID)
			if err != nil {
				return model.ErrorMsg{Err: err}
			}
			err = db.UpdateVisit(m.db, model.UpdateVisit{
				ID:           m.visitID,
				RestaurantID: restaurantID,
				VisitedOn:    date,
				Rating:       rating,
				Notes:        notes,
				WouldReturn:  wouldReturn,
			})
			if err != nil {
				return model.ErrorMsg{Err: err}
			}
			return model.VisitSavedMsg{
				ID:        m.visitID,
				Operation: "update",
				Before:    &before,
				After: model.Visit{
					ID:           m.visitID,
					RestaurantID: restaurantID,
					VisitedOn:    date,
					Rating:       rating,
					Notes:        notes,
					WouldReturn:  wouldReturn,
				},
			}
		} else {
			id, err := db.InsertVisit(m.db, model.NewVisit{
				RestaurantID: restaurantID,
				VisitedOn:    date,
				Rating:       rating,
				Notes:        notes,
				WouldReturn:  wouldReturn,
			})
			if err != nil {
				return model.ErrorMsg{Err: err}
			}
			return model.VisitSavedMsg{
				ID:        id,
				Operation: "insert",
				After: model.Visit{
					ID:           id,
					RestaurantID: restaurantID,
					VisitedOn:    date,
					Rating:       rating,
					Notes:        notes,
					WouldReturn:  wouldReturn,
				},
			}
		}
	}
}

func findExactRestaurantID(restaurants []model.Restaurant, restaurantName string) (int64, bool) {
	target := strings.ToLower(strings.TrimSpace(restaurantName))
	for _, r := range restaurants {
		if strings.ToLower(strings.TrimSpace(r.Name)) == target {
			return r.ID, true
		}
	}
	return 0, false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func renderFormField(label string, input textinput.Model, focused bool) string {
	style := BorderStyle
	if focused {
		style = ActiveBorderStyle
	}

	field := lipgloss.JoinVertical(
		lipgloss.Left,
		LabelStyle.Render(label),
		input.View(),
	)

	return style.Render(field)
}
