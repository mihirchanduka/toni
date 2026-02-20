package model

import "time"

// Restaurant represents a restaurant entity.
type Restaurant struct {
	ID           int64
	Name         string
	Address      string
	City         string
	Neighborhood string
	Cuisine      string
	PriceRange   string
	Latitude     *float64
	Longitude    *float64
	PlaceID      string
	CreatedAt    time.Time
}

// Visit represents a visit to a restaurant.
type Visit struct {
	ID           int64
	RestaurantID int64
	VisitedOn    string // ISO 8601 date (YYYY-MM-DD)
	Rating       *float64
	Notes        string
	WouldReturn  *bool
	CreatedAt    time.Time
}

// VisitRow represents a visit with joined restaurant data for list display.
type VisitRow struct {
	ID             int64
	VisitedOn      string
	RestaurantName string
	City           string
	Address        string
	PriceRange     string
	Rating         *float64
	WouldReturn    *bool
	Notes          string
	RestaurantID   int64
}

// RestaurantRow represents a restaurant with aggregate stats for list display.
type RestaurantRow struct {
	ID           int64
	Name         string
	Address      string
	City         string
	Neighborhood string
	Cuisine      string
	PriceRange   string
	AvgRating    *float64
	VisitCount   int
	LastVisit    string
}

// RestaurantDetail represents a restaurant with all its visits.
type RestaurantDetail struct {
	Restaurant Restaurant
	Visits     []Visit
}

// NewRestaurant represents data for creating a restaurant.
type NewRestaurant struct {
	Name         string
	Address      string
	City         string
	Neighborhood string
	Cuisine      string
	PriceRange   string
	Latitude     *float64
	Longitude    *float64
	PlaceID      string
}

// NewVisit represents data for creating a visit.
type NewVisit struct {
	RestaurantID int64
	VisitedOn    string
	Rating       *float64
	Notes        string
	WouldReturn  *bool
}

// UpdateRestaurant represents data for updating a restaurant.
type UpdateRestaurant struct {
	ID           int64
	Name         string
	Address      string
	City         string
	Neighborhood string
	Cuisine      string
	PriceRange   string
	Latitude     *float64
	Longitude    *float64
	PlaceID      string
}

// UpdateVisit represents data for updating a visit.
type UpdateVisit struct {
	ID           int64
	RestaurantID int64
	VisitedOn    string
	Rating       *float64
	Notes        string
	WouldReturn  *bool
}

// WantToVisit represents a place the user wants to visit.
type WantToVisit struct {
	ID           int64
	RestaurantID int64
	Notes        string
	Priority     *int // 1-5, 5 being highest priority
	CreatedAt    time.Time
}

// WantToVisitRow represents a want_to_visit with joined restaurant data for list display.
type WantToVisitRow struct {
	ID             int64
	RestaurantName string
	Address        string
	City           string
	Neighborhood   string
	Cuisine        string
	PriceRange     string
	Priority       *int
	Notes          string
	RestaurantID   int64
	CreatedAt      time.Time
}

// NewWantToVisit represents data for creating a want_to_visit entry.
type NewWantToVisit struct {
	RestaurantID int64
	Notes        string
	Priority     *int
}

// UpdateWantToVisit represents data for updating a want_to_visit entry.
type UpdateWantToVisit struct {
	ID           int64
	RestaurantID int64
	Notes        string
	Priority     *int
}
