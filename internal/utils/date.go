// Package utils provides utility functions for the RSS reader
package utils

import (
	"time"
)

// ParseDate attempts to parse a date string using common RSS/Atom formats
func ParseDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}
	
	for _, f := range formats {
		t, err := time.Parse(f, dateStr)
		if err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, nil // fallback: return zero time, no error
}
