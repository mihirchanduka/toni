package db

import (
	"database/sql"
	"fmt"
	"time"
	"toni/internal/model"
)

// ListRestaurants retrieves all restaurants with aggregate stats, optionally filtered.
func ListRestaurants(db *sql.DB, filter string) ([]model.RestaurantRow, error) {
	query := `
		SELECT
			r.id,
			r.name,
			COALESCE(r.address, ''),
			COALESCE(r.city, ''),
			COALESCE(r.neighborhood, ''),
			COALESCE(r.cuisine, ''),
			COALESCE(r.price_range, ''),
			AVG(v.rating) as avg_rating,
			COUNT(v.id) as visit_count,
			MAX(v.visited_on) as last_visit
		FROM restaurants r
		LEFT JOIN visits v ON r.id = v.restaurant_id
		WHERE (? = '' OR r.name LIKE '%' || ? || '%' OR r.city LIKE '%' || ? || '%')
		GROUP BY r.id
		ORDER BY r.name
	`

	rows, err := db.Query(query, filter, filter, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list restaurants: %w", err)
	}
	defer rows.Close()

	var results []model.RestaurantRow
	for rows.Next() {
		var r model.RestaurantRow
		var avgRating sql.NullFloat64
		var lastVisit sql.NullString
		if err := rows.Scan(&r.ID, &r.Name, &r.Address, &r.City, &r.Neighborhood, &r.Cuisine, &r.PriceRange, &avgRating, &r.VisitCount, &lastVisit); err != nil {
			return nil, fmt.Errorf("failed to scan restaurant row: %w", err)
		}
		if avgRating.Valid {
			r.AvgRating = &avgRating.Float64
		}
		if lastVisit.Valid {
			r.LastVisit = lastVisit.String
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating restaurant rows: %w", err)
	}

	return results, nil
}

// GetRestaurant retrieves a single restaurant by ID.
func GetRestaurant(db *sql.DB, id int64) (model.Restaurant, error) {
	query := `
		SELECT id, name, address, city, neighborhood, cuisine, price_range, latitude, longitude, place_id, created_at
		FROM restaurants
		WHERE id = ?
	`

	var r model.Restaurant
	var address, city, neighborhood, cuisine, priceRange, placeID sql.NullString
	var latitude, longitude sql.NullFloat64
	var createdAt string

	err := db.QueryRow(query, id).Scan(
		&r.ID, &r.Name, &address, &city, &neighborhood, &cuisine, &priceRange, &latitude, &longitude, &placeID, &createdAt,
	)
	if err != nil {
		return model.Restaurant{}, fmt.Errorf("failed to get restaurant: %w", err)
	}

	r.Address = address.String
	r.City = city.String
	r.Neighborhood = neighborhood.String
	r.Cuisine = cuisine.String
	r.PriceRange = priceRange.String
	r.PlaceID = placeID.String

	if latitude.Valid {
		r.Latitude = &latitude.Float64
	}
	if longitude.Valid {
		r.Longitude = &longitude.Float64
	}

	if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
		r.CreatedAt = t
	}

	return r, nil
}

// GetRestaurantWithStats retrieves a restaurant with all its visits.
func GetRestaurantWithStats(db *sql.DB, id int64) (model.RestaurantDetail, error) {
	restaurant, err := GetRestaurant(db, id)
	if err != nil {
		return model.RestaurantDetail{}, err
	}

	query := `
		SELECT id, restaurant_id, visited_on, rating, notes, would_return, created_at
		FROM visits
		WHERE restaurant_id = ?
		ORDER BY visited_on DESC
	`

	rows, err := db.Query(query, id)
	if err != nil {
		return model.RestaurantDetail{}, fmt.Errorf("failed to get visits: %w", err)
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
			return model.RestaurantDetail{}, fmt.Errorf("failed to scan visit: %w", err)
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

	return model.RestaurantDetail{
		Restaurant: restaurant,
		Visits:     visits,
	}, nil
}

// InsertRestaurant creates a new restaurant.
func InsertRestaurant(db *sql.DB, r model.NewRestaurant) (int64, error) {
	query := `
		INSERT INTO restaurants (name, address, city, neighborhood, cuisine, price_range, latitude, longitude, place_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
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

	result, err := db.Exec(query, r.Name, address, city, neighborhood, cuisine, priceRange, latitude, longitude, placeID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert restaurant: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// UpdateRestaurant updates an existing restaurant.
func UpdateRestaurant(db *sql.DB, r model.UpdateRestaurant) error {
	query := `
		UPDATE restaurants
		SET name = ?, address = ?, city = ?, neighborhood = ?, cuisine = ?, price_range = ?, latitude = ?, longitude = ?, place_id = ?
		WHERE id = ?
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

	_, err := db.Exec(query, r.Name, address, city, neighborhood, cuisine, priceRange, latitude, longitude, placeID, r.ID)
	if err != nil {
		return fmt.Errorf("failed to update restaurant: %w", err)
	}

	return nil
}

// DeleteRestaurant deletes a restaurant and all its visits.
func DeleteRestaurant(db *sql.DB, id int64) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM visits WHERE restaurant_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete visits: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM want_to_visit WHERE restaurant_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete want_to_visit entries: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM restaurants WHERE id = ?", id); err != nil {
		return fmt.Errorf("failed to delete restaurant: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SearchRestaurants returns restaurants matching a search query.
func SearchRestaurants(db *sql.DB, query string) ([]model.Restaurant, error) {
	sqlQuery := `
		SELECT id, name, address, city, neighborhood, cuisine, price_range, latitude, longitude, place_id, created_at
		FROM restaurants
		WHERE name LIKE '%' || ? || '%'
		ORDER BY name
		LIMIT 20
	`

	rows, err := db.Query(sqlQuery, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search restaurants: %w", err)
	}
	defer rows.Close()

	var results []model.Restaurant
	for rows.Next() {
		var r model.Restaurant
		var address, city, neighborhood, cuisine, priceRange, placeID sql.NullString
		var latitude, longitude sql.NullFloat64
		var createdAt string

		if err := rows.Scan(&r.ID, &r.Name, &address, &city, &neighborhood, &cuisine, &priceRange, &latitude, &longitude, &placeID, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan restaurant: %w", err)
		}

		r.Address = address.String
		r.City = city.String
		r.Neighborhood = neighborhood.String
		r.Cuisine = cuisine.String
		r.PriceRange = priceRange.String
		r.PlaceID = placeID.String

		if latitude.Valid {
			r.Latitude = &latitude.Float64
		}
		if longitude.Valid {
			r.Longitude = &longitude.Float64
		}

		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			r.CreatedAt = t
		}

		results = append(results, r)
	}

	return results, nil
}
