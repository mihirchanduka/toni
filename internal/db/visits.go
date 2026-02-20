package db

import (
	"database/sql"
	"fmt"
	"time"
	"toni/internal/model"
)

// ListVisits retrieves all visits with restaurant info, optionally filtered.
func ListVisits(db *sql.DB, filter string) ([]model.VisitRow, error) {
	query := `
		SELECT
			v.id,
			COALESCE(v.visited_on, ''),
			r.name,
			COALESCE(r.city, ''),
			COALESCE(r.address, ''),
			COALESCE(r.price_range, ''),
			v.rating,
			v.would_return,
			COALESCE(v.notes, ''),
			v.restaurant_id
		FROM visits v
		JOIN restaurants r ON v.restaurant_id = r.id
		WHERE (? = '' OR r.name LIKE '%' || ? || '%' OR v.notes LIKE '%' || ? || '%')
		ORDER BY v.visited_on DESC
	`

	rows, err := db.Query(query, filter, filter, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list visits: %w", err)
	}
	defer rows.Close()

	var results []model.VisitRow
	for rows.Next() {
		var v model.VisitRow
		var rating sql.NullFloat64
		var wouldReturn sql.NullInt64

		if err := rows.Scan(&v.ID, &v.VisitedOn, &v.RestaurantName, &v.City, &v.Address, &v.PriceRange, &rating, &wouldReturn, &v.Notes, &v.RestaurantID); err != nil {
			return nil, fmt.Errorf("failed to scan visit row: %w", err)
		}

		if rating.Valid {
			r := rating.Float64
			v.Rating = &r
		}
		if wouldReturn.Valid {
			wr := wouldReturn.Int64 == 1
			v.WouldReturn = &wr
		}

		results = append(results, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating visit rows: %w", err)
	}

	return results, nil
}

// GetVisit retrieves a single visit by ID.
func GetVisit(db *sql.DB, id int64) (model.Visit, error) {
	query := `
		SELECT id, restaurant_id, visited_on, rating, notes, would_return, created_at
		FROM visits
		WHERE id = ?
	`

	var v model.Visit
	var visitedOn sql.NullString
	var rating sql.NullFloat64
	var notes sql.NullString
	var wouldReturn sql.NullInt64
	var createdAt string

	err := db.QueryRow(query, id).Scan(
		&v.ID, &v.RestaurantID, &visitedOn, &rating, &notes, &wouldReturn, &createdAt,
	)
	if err != nil {
		return model.Visit{}, fmt.Errorf("failed to get visit: %w", err)
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

	return v, nil
}

// InsertVisit creates a new visit.
func InsertVisit(db *sql.DB, v model.NewVisit) (int64, error) {
	query := `
		INSERT INTO visits (restaurant_id, visited_on, rating, notes, would_return)
		VALUES (?, ?, ?, ?, ?)
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

	result, err := db.Exec(query, v.RestaurantID, visitedOn, rating, notes, wouldReturn)
	if err != nil {
		return 0, fmt.Errorf("failed to insert visit: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// UpdateVisit updates an existing visit.
func UpdateVisit(db *sql.DB, v model.UpdateVisit) error {
	query := `
		UPDATE visits
		SET restaurant_id = ?, visited_on = ?, rating = ?, notes = ?, would_return = ?
		WHERE id = ?
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

	_, err := db.Exec(query, v.RestaurantID, visitedOn, rating, notes, wouldReturn, v.ID)
	if err != nil {
		return fmt.Errorf("failed to update visit: %w", err)
	}

	return nil
}

// DeleteVisit deletes a visit.
func DeleteVisit(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM visits WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete visit: %w", err)
	}
	return nil
}
