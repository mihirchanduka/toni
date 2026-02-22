package util

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// FormatDate formats a date string (YYYY-MM-DD) for display.
func FormatDate(date string) string {
	date = strings.TrimSpace(date)
	if date == "" {
		return "Unknown"
	}
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("Jan 02, 2006")
}

// FormatDateHuman formats a date with humanized relative display.
// "Today", "Yesterday", "3d ago", "Jan 15", "Jan 15 '24"
func FormatDateHuman(date string) string {
	date = strings.TrimSpace(date)
	if date == "" {
		return "Unknown"
	}
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dateDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	diff := today.Sub(dateDay)
	days := int(diff.Hours() / 24)

	switch {
	case days == 0:
		return "Today"
	case days == 1:
		return "Yesterday"
	case days > 1 && days < 7:
		return fmt.Sprintf("%dd ago", days)
	case t.Year() == now.Year():
		return t.Format("Jan 02")
	default:
		return t.Format("Jan 02 '06")
	}
}

// FormatRating formats a rating as "8.5/10" or "—" if nil.
func FormatRating(rating *float64) string {
	if rating == nil {
		return "—"
	}
	return formatRatingNumber(*rating) + "/10"
}

// FormatRatingWithStar formats a rating as "8.5 ★" for display.
func FormatRatingWithStar(rating *float64) string {
	if rating == nil {
		return "—"
	}
	return formatRatingNumber(*rating) + " ★"
}

// FormatRatingStars formats a rating as stars (e.g., "★★★★☆").
func FormatRatingStars(rating *float64) string {
	if rating == nil {
		return "—"
	}
	// Convert 1-10 to 1-5 stars, rounded to nearest whole star.
	stars := int(math.Round(*rating / 2.0))
	if stars < 0 {
		stars = 0
	}
	if stars > 5 {
		stars = 5
	}
	result := ""
	for i := 0; i < 5; i++ {
		if i < stars {
			result += "★"
		} else {
			result += "☆"
		}
	}
	return result
}

// FormatWouldReturn formats the would_return boolean.
func FormatWouldReturn(wouldReturn *bool) string {
	if wouldReturn == nil {
		return "—"
	}
	if *wouldReturn {
		return "Yes"
	}
	return "No"
}

// FormatWouldReturnSymbol formats would_return as symbols: ✓, ✗, or –
func FormatWouldReturnSymbol(wouldReturn *bool) string {
	if wouldReturn == nil {
		return "–"
	}
	if *wouldReturn {
		return "✓"
	}
	return "✗"
}

// FormatAvgRating formats an average rating.
func FormatAvgRating(avg *float64) string {
	if avg == nil {
		return "—"
	}
	return fmt.Sprintf("%.1f", *avg)
}

// TodayISO returns today's date in ISO 8601 format (YYYY-MM-DD).
func TodayISO() string {
	return time.Now().Format("2006-01-02")
}

// ValidateDate validates a date string in YYYY-MM-DD format.
func ValidateDate(date string) error {
	_, err := time.Parse("2006-01-02", date)
	return err
}

func formatRatingNumber(v float64) string {
	// Keep one decimal at most, but avoid trailing .0 for whole values.
	s := strconv.FormatFloat(v, 'f', 1, 64)
	s = strings.TrimSuffix(s, ".0")
	return s
}

// ParseVisitDateInput parses flexible user input and normalizes to ISO (YYYY-MM-DD).
// Empty input is allowed and returns "".
func ParseVisitDateInput(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", nil
	}

	layouts := []string{
		"2006-01-02",
		"January 2, 2006",
		"Jan 2, 2006",
		"1/2/2006",
		"01/02/2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}

	return "", fmt.Errorf("invalid date format")
}

// TruncateString truncates a string to maxLen and adds "..." if needed.
func TruncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
