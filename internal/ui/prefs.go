package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TablePrefs stores per-table UI preferences.
type TablePrefs struct {
	SortKey       string   `json:"sort_key"`
	SortDesc      bool     `json:"sort_desc"`
	HiddenColumns []string `json:"hidden_columns"`
	ActiveColumn  string   `json:"active_column"`
}

// UIPreferences stores persisted app preferences.
type UIPreferences struct {
	Visits      TablePrefs `json:"visits"`
	Restaurants TablePrefs `json:"restaurants"`
	WantToVisit TablePrefs `json:"want_to_visit"`
}

func defaultUIPreferences() UIPreferences {
	return UIPreferences{}
}

func prefsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}
	return filepath.Join(home, ".toni", "ui_prefs.json"), nil
}

func loadUIPreferences() UIPreferences {
	path, err := prefsPath()
	if err != nil {
		return defaultUIPreferences()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return defaultUIPreferences()
	}

	var prefs UIPreferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return defaultUIPreferences()
	}
	return prefs
}

func saveUIPreferences(prefs UIPreferences) error {
	path, err := prefsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create prefs dir: %w", err)
	}

	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal prefs: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write prefs: %w", err)
	}
	return nil
}
