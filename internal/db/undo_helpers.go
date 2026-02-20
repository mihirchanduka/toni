package db

import (
	"database/sql"
	"fmt"
	"time"
	"toni/internal/model"
)

func InsertRestaurantWithID(db *sql.DB, r model.Restaurant) error {
	query := `
		INSERT INTO restaurants (id, name, address, city, neighborhood, cuisine, price_range, latitude, longitude, place_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var address, city, neighborhood, cuisine, priceRange, placeID interface{}
	var latitude, longitude interface{}
	if r.Address != "" {
		address = r.Address
	}
	if r.City != "" {
		city = r.City
	}
	if r.Neighborhood != "" {
		neighborhood = r.Neighborhood
	}
	if r.Cuisine != "" {
		cuisine = r.Cuisine
	}
	if r.PriceRange != "" {
		priceRange = r.PriceRange
	}
	if r.Latitude != nil {
		latitude = *r.Latitude
	}
	if r.Longitude != nil {
		longitude = *r.Longitude
	}
	if r.PlaceID != "" {
		placeID = r.PlaceID
	}
	createdAt := time.Now().UTC().Format(time.RFC3339)
	if !r.CreatedAt.IsZero() {
		createdAt = r.CreatedAt.UTC().Format(time.RFC3339)
	}

	if _, err := db.Exec(query, r.ID, r.Name, address, city, neighborhood, cuisine, priceRange, latitude, longitude, placeID, createdAt); err != nil {
		return fmt.Errorf("failed to insert restaurant with id: %w", err)
	}
	return nil
}

func InsertVisitWithID(db *sql.DB, v model.Visit) error {
	query := `
		INSERT INTO visits (id, restaurant_id, visited_on, rating, notes, would_return, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	var visitedOn interface{}
	var rating, wouldReturn interface{}
	var notes interface{}
	if v.VisitedOn != "" {
		visitedOn = v.VisitedOn
	}
	if v.Rating != nil {
		rating = *v.Rating
	}
	if v.Notes != "" {
		notes = v.Notes
	}
	if v.WouldReturn != nil {
		if *v.WouldReturn {
			wouldReturn = 1
		} else {
			wouldReturn = 0
		}
	}
	createdAt := time.Now().UTC().Format(time.RFC3339)
	if !v.CreatedAt.IsZero() {
		createdAt = v.CreatedAt.UTC().Format(time.RFC3339)
	}

	if _, err := db.Exec(query, v.ID, v.RestaurantID, visitedOn, rating, notes, wouldReturn, createdAt); err != nil {
		return fmt.Errorf("failed to insert visit with id: %w", err)
	}
	return nil
}

func InsertWantToVisitWithID(db *sql.DB, w model.WantToVisit) error {
	query := `
		INSERT INTO want_to_visit (id, restaurant_id, notes, priority, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	var priority interface{}
	if w.Priority != nil {
		priority = *w.Priority
	}
	createdAt := time.Now().UTC().Format(time.RFC3339)
	if !w.CreatedAt.IsZero() {
		createdAt = w.CreatedAt.UTC().Format(time.RFC3339)
	}
	if _, err := db.Exec(query, w.ID, w.RestaurantID, w.Notes, priority, createdAt); err != nil {
		return fmt.Errorf("failed to insert want_to_visit with id: %w", err)
	}
	return nil
}

func GetVisitsByRestaurant(db *sql.DB, restaurantID int64) ([]model.Visit, error) {
	rows, err := db.Query(`
		SELECT id, restaurant_id, visited_on, rating, notes, would_return, created_at
		FROM visits
		WHERE restaurant_id = ?
		ORDER BY id
	`, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query visits by restaurant: %w", err)
	}
	defer rows.Close()

	var visits []model.Visit
	for rows.Next() {
		var v model.Visit
		var visitedOn sql.NullString
		var rating sql.NullFloat64
		var notes sql.NullString
		var wouldReturn sql.NullInt64
		var createdAt string
		if err := rows.Scan(&v.ID, &v.RestaurantID, &visitedOn, &rating, &notes, &wouldReturn, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan visit: %w", err)
		}
		v.VisitedOn = visitedOn.String
		if rating.Valid {
			r := rating.Float64
			v.Rating = &r
		}
		v.Notes = notes.String
		if wouldReturn.Valid {
			wr := wouldReturn.Int64 == 1
			v.WouldReturn = &wr
		}
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			v.CreatedAt = t
		}
		visits = append(visits, v)
	}
	return visits, rows.Err()
}

func GetWantToVisitByRestaurant(db *sql.DB, restaurantID int64) ([]model.WantToVisit, error) {
	rows, err := db.Query(`
		SELECT id, restaurant_id, notes, priority, created_at
		FROM want_to_visit
		WHERE restaurant_id = ?
		ORDER BY id
	`, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query want_to_visit by restaurant: %w", err)
	}
	defer rows.Close()

	var entries []model.WantToVisit
	for rows.Next() {
		var w model.WantToVisit
		var priority sql.NullInt64
		var createdAt string
		if err := rows.Scan(&w.ID, &w.RestaurantID, &w.Notes, &priority, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan want_to_visit: %w", err)
		}
		if priority.Valid {
			p := int(priority.Int64)
			w.Priority = &p
		}
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			w.CreatedAt = t
		}
		entries = append(entries, w)
	}
	return entries, rows.Err()
}
