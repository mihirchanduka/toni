package ui

import (
	"database/sql"
	"fmt"
	"strings"
	"toni/internal/db"
	"toni/internal/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// RestaurantFormModel represents the restaurant form.
type RestaurantFormModel struct {
	db           *sql.DB
	restaurantID int64
	focusedField int
	inputs       []textinput.Model
	error        string
}

// NewRestaurantFormModel creates a new restaurant form.
func NewRestaurantFormModel(database *sql.DB, restaurantID int64) *RestaurantFormModel {
	inputs := make([]textinput.Model, 6)

	// Name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Restaurant name"
	inputs[0].Focus()
	inputs[0].CharLimit = 100

	// Address
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Street address"
	inputs[1].CharLimit = 200

	// City
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "City"
	inputs[2].CharLimit = 100

	// Neighborhood
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Neighborhood"
	inputs[3].CharLimit = 100

	// Cuisine
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Cuisine"
	inputs[4].CharLimit = 100

	// Price Range
	inputs[5] = textinput.New()
	inputs[5].Placeholder = "$, $$, $$$, or $$$$"
	inputs[5].CharLimit = 4

	return &RestaurantFormModel{
		db:           database,
		restaurantID: restaurantID,
		focusedField: 0,
		inputs:       inputs,
	}
}

// LoadRestaurant loads an existing restaurant for editing.
func (m *RestaurantFormModel) LoadRestaurant(restaurant model.Restaurant) {
	m.restaurantID = restaurant.ID
	m.inputs[0].SetValue(restaurant.Name)
	m.inputs[1].SetValue(restaurant.Address)
	m.inputs[2].SetValue(restaurant.City)
	m.inputs[3].SetValue(restaurant.Neighborhood)
	m.inputs[4].SetValue(restaurant.Cuisine)
	m.inputs[5].SetValue(restaurant.PriceRange)
}

// Update handles input.
func (m RestaurantFormModel) Update(msg tea.KeyMsg) (RestaurantFormModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg {
			return model.FormCancelledMsg{}
		}
	case "ctrl+s":
		return m, m.save()
	case "tab":
		m.nextField()
		return m, nil
	case "shift+tab":
		m.prevField()
		return m, nil
	}

	// Update current input
	var cmd tea.Cmd
	m.inputs[m.focusedField], cmd = m.inputs[m.focusedField].Update(msg)
	return m, cmd
}

// View renders the form.
func (m *RestaurantFormModel) View(width, height int) string {
	var fields []string

	fields = append(fields, renderFormField("Name *", m.inputs[0], m.focusedField == 0))
	fields = append(fields, renderFormField("Address", m.inputs[1], m.focusedField == 1))
	fields = append(fields, renderFormField("City", m.inputs[2], m.focusedField == 2))
	fields = append(fields, renderFormField("Neighborhood", m.inputs[3], m.focusedField == 3))
	fields = append(fields, renderFormField("Cuisine", m.inputs[4], m.focusedField == 4))
	fields = append(fields, renderFormField("Price Range ($-$$$$)", m.inputs[5], m.focusedField == 5))

	if m.error != "" {
		fields = append(fields, "")
		fields = append(fields, ErrorStyle.Render(m.error))
	}

	formContent := strings.Join(fields, "\n\n")

	// Use PanelStyle with proper dimensions
	content := PanelStyle.
		Width(width - 4).
		Height(height - 4).
		Render(formContent)

	return content
}

func (m *RestaurantFormModel) nextField() {
	m.inputs[m.focusedField].Blur()
	m.focusedField = (m.focusedField + 1) % len(m.inputs)
	m.inputs[m.focusedField].Focus()
}

func (m *RestaurantFormModel) prevField() {
	m.inputs[m.focusedField].Blur()
	m.focusedField--
	if m.focusedField < 0 {
		m.focusedField = len(m.inputs) - 1
	}
	m.inputs[m.focusedField].Focus()
}

func (m *RestaurantFormModel) save() tea.Cmd {
	return func() tea.Msg {
		name := strings.TrimSpace(m.inputs[0].Value())
		if name == "" {
			return model.ErrorMsg{Err: fmt.Errorf("name is required")}
		}

		address := strings.TrimSpace(m.inputs[1].Value())
		city := strings.TrimSpace(m.inputs[2].Value())
		neighborhood := strings.TrimSpace(m.inputs[3].Value())
		cuisine := strings.TrimSpace(m.inputs[4].Value())
		priceRange := strings.TrimSpace(m.inputs[5].Value())

		// Validate price range
		if priceRange != "" {
			valid := false
			for _, p := range []string{"$", "$$", "$$$", "$$$$"} {
				if priceRange == p {
					valid = true
					break
				}
			}
			if !valid {
				return model.ErrorMsg{Err: fmt.Errorf("price range must be $, $$, $$$, or $$$$")}
			}
		}

		// Save
		if m.restaurantID > 0 {
			before, err := db.GetRestaurant(m.db, m.restaurantID)
			if err != nil {
				return model.ErrorMsg{Err: err}
			}
			err = db.UpdateRestaurant(m.db, model.UpdateRestaurant{
				ID:           m.restaurantID,
				Name:         name,
				Address:      address,
				City:         city,
				Neighborhood: neighborhood,
				Cuisine:      cuisine,
				PriceRange:   priceRange,
			})
			if err != nil {
				return model.ErrorMsg{Err: err}
			}
			return model.RestaurantSavedMsg{
				ID:        m.restaurantID,
				Operation: "update",
				Before:    &before,
				After: model.Restaurant{
					ID:           m.restaurantID,
					Name:         name,
					Address:      address,
					City:         city,
					Neighborhood: neighborhood,
					Cuisine:      cuisine,
					PriceRange:   priceRange,
					Latitude:     before.Latitude,
					Longitude:    before.Longitude,
					PlaceID:      before.PlaceID,
					CreatedAt:    before.CreatedAt,
				},
			}
		} else {
			id, err := db.InsertRestaurant(m.db, model.NewRestaurant{
				Name:         name,
				Address:      address,
				City:         city,
				Neighborhood: neighborhood,
				Cuisine:      cuisine,
				PriceRange:   priceRange,
			})
			if err != nil {
				return model.ErrorMsg{Err: err}
			}
			return model.RestaurantSavedMsg{
				ID:        id,
				Operation: "insert",
				After: model.Restaurant{
					ID:           id,
					Name:         name,
					Address:      address,
					City:         city,
					Neighborhood: neighborhood,
					Cuisine:      cuisine,
					PriceRange:   priceRange,
				},
			}
		}
	}
}
