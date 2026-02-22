package db

import (
	"database/sql"
	"fmt"
	"time"
	"toni/internal/model"
)

// GetWantToVisitList returns all want_to_visit entries with restaurant info.
func GetWantToVisitList(db *sql.DB, orderBy string) ([]model.WantToVisitRow, error) {
	if orderBy == "" {
		orderBy = "w.priority DESC NULLS LAST, w.created_at DESC"
	}

	query := fmt.Sprintf(`
		SELECT
			w.id,
			r.name,
			COALESCE(r.address, ''),
			COALESCE(r.city, ''),
			COALESCE(r.neighborhood, ''),
			COALESCE(r.cuisine, ''),
			COALESCE(r.price_range, ''),
			w.priority,
			COALESCE(w.notes, ''),
			w.restaurant_id,
			w.created_at
		FROM want_to_visit w
		JOIN restaurants r ON w.restaurant_id = r.id
		ORDER BY %s
	`, orderBy)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.WantToVisitRow
	for rows.Next() {
		var row model.WantToVisitRow
		var createdAt string
		if err := rows.Scan(
			&row.ID,
			&row.RestaurantName,
			&row.Address,
			&row.City,
			&row.Neighborhood,
			&row.Cuisine,
			&row.PriceRange,
			&row.Priority,
			&row.Notes,
			&row.RestaurantID,
			&createdAt,
		); err != nil {
			return nil, err
		}
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			row.CreatedAt = t
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

// GetWantToVisit returns a single want_to_visit entry by ID.
func GetWantToVisit(db *sql.DB, id int64) (model.WantToVisit, error) {
	var wtv model.WantToVisit
	var createdAt string
	var notes sql.NullString
	var priority sql.NullInt64
	err := db.QueryRow(`
		SELECT id, restaurant_id, notes, priority, created_at
		FROM want_to_visit
		WHERE id = ?
	`, id).Scan(&wtv.ID, &wtv.RestaurantID, &notes, &priority, &createdAt)
	if err != nil {
		return wtv, err
	}
	if notes.Valid {
		wtv.Notes = notes.String
	}
	if priority.Valid {
		p := int(priority.Int64)
		wtv.Priority = &p
	}

	if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
		wtv.CreatedAt = t
	}

	return wtv, nil
}

// InsertWantToVisit creates a new want_to_visit entry.
func InsertWantToVisit(db *sql.DB, wtv model.NewWantToVisit) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO want_to_visit (restaurant_id, notes, priority)
		VALUES (?, ?, ?)
	`, wtv.RestaurantID, wtv.Notes, wtv.Priority)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateWantToVisit updates an existing want_to_visit entry.
func UpdateWantToVisit(db *sql.DB, wtv model.UpdateWantToVisit) error {
	_, err := db.Exec(`
		UPDATE want_to_visit
		SET restaurant_id = ?, notes = ?, priority = ?
		WHERE id = ?
	`, wtv.RestaurantID, wtv.Notes, wtv.Priority, wtv.ID)

	return err
}

// DeleteWantToVisit deletes a want_to_visit entry.
func DeleteWantToVisit(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM want_to_visit WHERE id = ?", id)
	return err
}

// ConvertWantToVisitToVisit deletes a want_to_visit entry and returns its restaurant ID.
// The caller should then create a visit form with this restaurant pre-filled.
func ConvertWantToVisitToVisit(db *sql.DB, wtvID int64) (int64, error) {
	wtv, err := GetWantToVisit(db, wtvID)
	if err != nil {
		return 0, err
	}

	// Delete the want_to_visit entry
	if err := DeleteWantToVisit(db, wtvID); err != nil {
		return 0, err
	}

	return wtv.RestaurantID, nil
}
