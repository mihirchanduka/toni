package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS restaurants (
    id           INTEGER PRIMARY KEY,
    name         TEXT NOT NULL,
    address      TEXT,
    city         TEXT,
    neighborhood TEXT,
    cuisine      TEXT,
    price_range  TEXT CHECK(price_range IN ('$','$$','$$$','$$$$') OR price_range IS NULL),
    latitude     REAL,
    longitude    REAL,
    place_id     TEXT,
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS visits (
    id            INTEGER PRIMARY KEY,
    restaurant_id INTEGER NOT NULL REFERENCES restaurants(id),
    visited_on    TEXT,
    rating        REAL CHECK(rating BETWEEN 1 AND 10 OR rating IS NULL),
    notes         TEXT,
    would_return  INTEGER CHECK(would_return IN (0,1) OR would_return IS NULL),
    created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS want_to_visit (
    id            INTEGER PRIMARY KEY,
    restaurant_id INTEGER NOT NULL REFERENCES restaurants(id),
    notes         TEXT,
    priority      INTEGER CHECK(priority BETWEEN 1 AND 5 OR priority IS NULL),
    created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX IF NOT EXISTS idx_visits_restaurant_id ON visits(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_visits_visited_on ON visits(visited_on DESC);
CREATE INDEX IF NOT EXISTS idx_want_to_visit_restaurant_id ON want_to_visit(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_want_to_visit_priority ON want_to_visit(priority DESC);
`

// Open opens or creates the SQLite database and initializes the schema.
func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}
