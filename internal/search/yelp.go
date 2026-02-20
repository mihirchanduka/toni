package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const yelpAPIBase = "https://api.yelp.com/v3"

// YelpClient wraps the Yelp Fusion API.
type YelpClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewYelpClient creates a new Yelp Fusion API client.
func NewYelpClient(apiKey string) *YelpClient {
	return &YelpClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// Suggestion represents a restaurant autocomplete result.
type Suggestion struct {
	Name         string
	City         string
	Neighborhood string
	Address      string
	Cuisine      string
	PriceRange   string // $, $$, $$$, $$$$
	Latitude     float64
	Longitude    float64
	PlaceID      string // Yelp business ID
}

// Autocomplete searches for restaurant businesses matching the query.
// Uses Yelp's Business Search API with partial name matching.
func (c *YelpClient) Autocomplete(ctx context.Context, query, location string) ([]Suggestion, error) {
	if query == "" {
		return []Suggestion{}, nil
	}

	// Build request URL - using Business Search for better results
	params := url.Values{}
	params.Set("term", query)
	params.Set("categories", "restaurants,food")
	params.Set("limit", "8")
	params.Set("sort_by", "best_match")

	// Add location if provided, otherwise default to broad search
	if location != "" {
		params.Set("location", location)
	} else {
		// Default to New York if no location specified
		params.Set("location", "New York, NY")
	}

	reqURL := fmt.Sprintf("%s/businesses/search?%s", yelpAPIBase, params.Encode())

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return []Suggestion{}, fmt.Errorf("request creation failed: %w", err)
	}

	// Yelp Fusion API authentication
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return []Suggestion{}, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	// Non-2xx response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return []Suggestion{}, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	// Parse response
	var result businessSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []Suggestion{}, fmt.Errorf("JSON decode error: %w", err)
	}

	// Convert to suggestions
	suggestions := make([]Suggestion, 0, len(result.Businesses))
	for _, business := range result.Businesses {
		suggestion := Suggestion{
			Name:      business.Name,
			PlaceID:   business.ID,
			Latitude:  business.Coordinates.Latitude,
			Longitude: business.Coordinates.Longitude,
		}

		// Location details
		if business.Location != nil {
			suggestion.Address = business.Location.Address1
			suggestion.City = business.Location.City
			// Use state as neighborhood proxy if available
			if business.Location.State != "" {
				suggestion.Neighborhood = business.Location.State
			}
		}

		// Price range - Yelp already uses $ format!
		if business.Price != "" {
			suggestion.PriceRange = business.Price
		}

		// Cuisine from first category
		if len(business.Categories) > 0 {
			suggestion.Cuisine = business.Categories[0].Title
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// GetBusinessDetails fetches full details for a business by its Yelp ID.
func (c *YelpClient) GetBusinessDetails(ctx context.Context, businessID string) (*Suggestion, error) {
	reqURL := fmt.Sprintf("%s/businesses/%s", yelpAPIBase, businessID)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}

	// Authentication
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	// Parse response
	var business businessDetail
	if err := json.NewDecoder(resp.Body).Decode(&business); err != nil {
		return nil, fmt.Errorf("JSON decode error: %w", err)
	}

	// Convert to Suggestion
	suggestion := &Suggestion{
		Name:      business.Name,
		PlaceID:   business.ID,
		Latitude:  business.Coordinates.Latitude,
		Longitude: business.Coordinates.Longitude,
	}

	// Location details
	if business.Location != nil {
		suggestion.Address = business.Location.Address1
		suggestion.City = business.Location.City
		if business.Location.State != "" {
			suggestion.Neighborhood = business.Location.State
		}
	}

	// Price range
	if business.Price != "" {
		suggestion.PriceRange = business.Price
	}

	// Cuisine from first category
	if len(business.Categories) > 0 {
		suggestion.Cuisine = business.Categories[0].Title
	}

	return suggestion, nil
}

// API response types

type businessSearchResponse struct {
	Businesses []businessDetail `json:"businesses"`
	Total      int              `json:"total"`
}

type businessDetail struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	ImageURL    string      `json:"image_url"`
	URL         string      `json:"url"`
	Rating      float64     `json:"rating"`
	ReviewCount int         `json:"review_count"`
	Price       string      `json:"price"`
	Categories  []category  `json:"categories"`
	Coordinates coordinates `json:"coordinates"`
	Location    *location   `json:"location"`
	Phone       string      `json:"phone"`
	IsClosed    bool        `json:"is_closed"`
}

type category struct {
	Alias string `json:"alias"`
	Title string `json:"title"`
}

type coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type location struct {
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	Address3 string `json:"address3"`
	City     string `json:"city"`
	State    string `json:"state"`
	ZipCode  string `json:"zip_code"`
	Country  string `json:"country"`
}
