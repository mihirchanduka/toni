package model

// Bubble Tea message types

// ErrorMsg represents an error message.
type ErrorMsg struct {
	Err error
}

// VisitsLoadedMsg is sent when visits are loaded.
type VisitsLoadedMsg struct {
	Visits []VisitRow
}

// RestaurantsLoadedMsg is sent when restaurants are loaded.
type RestaurantsLoadedMsg struct {
	Restaurants []RestaurantRow
}

// VisitDetailLoadedMsg is sent when a visit detail is loaded.
type VisitDetailLoadedMsg struct {
	Visit      Visit
	Restaurant Restaurant
}

// RestaurantDetailLoadedMsg is sent when a restaurant detail is loaded.
type RestaurantDetailLoadedMsg struct {
	Detail RestaurantDetail
}

// VisitSavedMsg is sent when a visit is successfully saved.
type VisitSavedMsg struct {
	ID        int64
	Operation string // insert, update
	Before    *Visit
	After     Visit
}

// RestaurantSavedMsg is sent when a restaurant is successfully saved.
type RestaurantSavedMsg struct {
	ID        int64
	Operation string // insert, update
	Before    *Restaurant
	After     Restaurant
}

// RestaurantSearchResultsMsg is sent when restaurant search completes.
type RestaurantSearchResultsMsg struct {
	Restaurants []Restaurant
}

// FormCancelledMsg is sent when a form is cancelled.
type FormCancelledMsg struct{}

// DeleteVisitMsg is sent to delete a visit.
type DeleteVisitMsg struct {
	ID      int64
	Deleted Visit
}

// DeleteRestaurantMsg is sent to delete a restaurant.
type DeleteRestaurantMsg struct {
	ID                 int64
	Deleted            Restaurant
	DeletedVisits      []Visit
	DeletedWantToVisit []WantToVisit
}

// WantToVisitLoadedMsg is sent when want_to_visit list is loaded.
type WantToVisitLoadedMsg struct {
	WantToVisit []WantToVisitRow
}

// WantToVisitSavedMsg is sent when a want_to_visit is successfully saved.
type WantToVisitSavedMsg struct {
	ID        int64
	Operation string // insert, update
	Before    *WantToVisit
	After     WantToVisit
}

// DeleteWantToVisitMsg is sent to delete a want_to_visit entry.
type DeleteWantToVisitMsg struct {
	ID      int64
	Deleted WantToVisit
}

// ConvertToVisitMsg is sent to convert want_to_visit to actual visit.
type ConvertToVisitMsg struct {
	WantToVisitID int64
	RestaurantID  int64
	Deleted       WantToVisit
}

// Screen represents different app screens.
type Screen int

const (
	ScreenVisits Screen = iota
	ScreenRestaurants
	ScreenWantToVisit
	ScreenVisitDetail
	ScreenRestaurantDetail
	ScreenWantToVisitDetail
	ScreenVisitForm
	ScreenRestaurantForm
	ScreenWantToVisitForm
)

// Mode represents the current interaction mode.
type Mode int

const (
	ModeNav Mode = iota
	ModeInsert
)
