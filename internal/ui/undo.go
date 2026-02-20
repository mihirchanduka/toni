package ui

import (
	"fmt"
	"toni/internal/db"
	"toni/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

type undoAction struct {
	label string
	undo  func() error
	redo  func() error
}

type undoAppliedMsg struct {
	err       error
	action    undoAction
	direction string // undo, redo
}

func (m *Model) pushUndoAction(action undoAction) {
	m.undoStack = append(m.undoStack, action)
	m.redoStack = nil
}

func (m *Model) undoCmd() tea.Cmd {
	if len(m.undoStack) == 0 {
		return nil
	}
	action := m.undoStack[len(m.undoStack)-1]
	m.undoStack = m.undoStack[:len(m.undoStack)-1]
	return func() tea.Msg {
		err := action.undo()
		return undoAppliedMsg{err: err, action: action, direction: "undo"}
	}
}

func (m *Model) redoCmd() tea.Cmd {
	if len(m.redoStack) == 0 {
		return nil
	}
	action := m.redoStack[len(m.redoStack)-1]
	m.redoStack = m.redoStack[:len(m.redoStack)-1]
	return func() tea.Msg {
		err := action.redo()
		return undoAppliedMsg{err: err, action: action, direction: "redo"}
	}
}

func (m *Model) buildVisitSaveAction(msg model.VisitSavedMsg) *undoAction {
	switch msg.Operation {
	case "insert":
		after := msg.After
		return &undoAction{
			label: "visit saved",
			undo: func() error {
				return db.DeleteVisit(m.db, after.ID)
			},
			redo: func() error {
				return db.InsertVisitWithID(m.db, after)
			},
		}
	case "update":
		if msg.Before == nil {
			return nil
		}
		before := *msg.Before
		after := msg.After
		return &undoAction{
			label: "visit updated",
			undo: func() error {
				return db.UpdateVisit(m.db, visitToUpdate(before))
			},
			redo: func() error {
				return db.UpdateVisit(m.db, visitToUpdate(after))
			},
		}
	default:
		return nil
	}
}

func (m *Model) buildRestaurantSaveAction(msg model.RestaurantSavedMsg) *undoAction {
	switch msg.Operation {
	case "insert":
		after := msg.After
		return &undoAction{
			label: "restaurant saved",
			undo: func() error {
				return db.DeleteRestaurant(m.db, after.ID)
			},
			redo: func() error {
				return db.InsertRestaurantWithID(m.db, after)
			},
		}
	case "update":
		if msg.Before == nil {
			return nil
		}
		before := *msg.Before
		after := msg.After
		return &undoAction{
			label: "restaurant updated",
			undo: func() error {
				return db.UpdateRestaurant(m.db, restaurantToUpdate(before))
			},
			redo: func() error {
				return db.UpdateRestaurant(m.db, restaurantToUpdate(after))
			},
		}
	default:
		return nil
	}
}

func (m *Model) buildWantToVisitSaveAction(msg model.WantToVisitSavedMsg) *undoAction {
	switch msg.Operation {
	case "insert":
		after := msg.After
		return &undoAction{
			label: "want_to_visit saved",
			undo: func() error {
				return db.DeleteWantToVisit(m.db, after.ID)
			},
			redo: func() error {
				return db.InsertWantToVisitWithID(m.db, after)
			},
		}
	case "update":
		if msg.Before == nil {
			return nil
		}
		before := *msg.Before
		after := msg.After
		return &undoAction{
			label: "want_to_visit updated",
			undo: func() error {
				return db.UpdateWantToVisit(m.db, wantToVisitToUpdate(before))
			},
			redo: func() error {
				return db.UpdateWantToVisit(m.db, wantToVisitToUpdate(after))
			},
		}
	default:
		return nil
	}
}

func (m *Model) buildDeleteVisitAction(msg model.DeleteVisitMsg) undoAction {
	deleted := msg.Deleted
	return undoAction{
		label: "visit deleted",
		undo: func() error {
			return db.InsertVisitWithID(m.db, deleted)
		},
		redo: func() error {
			return db.DeleteVisit(m.db, deleted.ID)
		},
	}
}

func (m *Model) buildDeleteWantToVisitAction(msg model.DeleteWantToVisitMsg) undoAction {
	deleted := msg.Deleted
	return undoAction{
		label: "want_to_visit deleted",
		undo: func() error {
			return db.InsertWantToVisitWithID(m.db, deleted)
		},
		redo: func() error {
			return db.DeleteWantToVisit(m.db, deleted.ID)
		},
	}
}

func (m *Model) buildDeleteRestaurantAction(msg model.DeleteRestaurantMsg) undoAction {
	deleted := msg.Deleted
	visits := append([]model.Visit(nil), msg.DeletedVisits...)
	entries := append([]model.WantToVisit(nil), msg.DeletedWantToVisit...)
	return undoAction{
		label: "restaurant deleted",
		undo: func() error {
			if err := db.InsertRestaurantWithID(m.db, deleted); err != nil {
				return err
			}
			for _, e := range entries {
				if err := db.InsertWantToVisitWithID(m.db, e); err != nil {
					return err
				}
			}
			for _, v := range visits {
				if err := db.InsertVisitWithID(m.db, v); err != nil {
					return err
				}
			}
			return nil
		},
		redo: func() error {
			return db.DeleteRestaurant(m.db, deleted.ID)
		},
	}
}

func (m *Model) buildConvertAction(msg model.ConvertToVisitMsg) undoAction {
	deleted := msg.Deleted
	return undoAction{
		label: "converted want_to_visit",
		undo: func() error {
			return db.InsertWantToVisitWithID(m.db, deleted)
		},
		redo: func() error {
			return db.DeleteWantToVisit(m.db, deleted.ID)
		},
	}
}

func (m *Model) reloadCurrentTopLevelCmd() tea.Cmd {
	switch m.screen {
	case model.ScreenVisits, model.ScreenVisitDetail, model.ScreenVisitForm:
		return loadVisitsCmd(m.db, "")
	case model.ScreenRestaurants, model.ScreenRestaurantDetail, model.ScreenRestaurantForm:
		return loadRestaurantsCmd(m.db, "")
	case model.ScreenWantToVisit, model.ScreenWantToVisitDetail, model.ScreenWantToVisitForm:
		return loadWantToVisitCmd(m.db)
	default:
		return loadVisitsCmd(m.db, "")
	}
}

func visitToUpdate(v model.Visit) model.UpdateVisit {
	return model.UpdateVisit{
		ID:           v.ID,
		RestaurantID: v.RestaurantID,
		VisitedOn:    v.VisitedOn,
		Rating:       v.Rating,
		Notes:        v.Notes,
		WouldReturn:  v.WouldReturn,
	}
}

func restaurantToUpdate(r model.Restaurant) model.UpdateRestaurant {
	return model.UpdateRestaurant{
		ID:           r.ID,
		Name:         r.Name,
		Address:      r.Address,
		City:         r.City,
		Neighborhood: r.Neighborhood,
		Cuisine:      r.Cuisine,
		PriceRange:   r.PriceRange,
		Latitude:     r.Latitude,
		Longitude:    r.Longitude,
		PlaceID:      r.PlaceID,
	}
}

func wantToVisitToUpdate(w model.WantToVisit) model.UpdateWantToVisit {
	return model.UpdateWantToVisit{
		ID:           w.ID,
		RestaurantID: w.RestaurantID,
		Notes:        w.Notes,
		Priority:     w.Priority,
	}
}

func (m *Model) applyUndoResult(msg undoAppliedMsg) tea.Cmd {
	if msg.err != nil {
		m.error = fmt.Sprintf("%s failed: %v", msg.direction, msg.err)
		return nil
	}

	if msg.direction == "undo" {
		m.redoStack = append(m.redoStack, msg.action)
		m.info = "Undid: " + msg.action.label
	} else {
		m.undoStack = append(m.undoStack, msg.action)
		m.info = "Redid: " + msg.action.label
	}
	m.error = ""
	return m.reloadCurrentTopLevelCmd()
}
