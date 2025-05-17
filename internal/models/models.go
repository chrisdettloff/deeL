// Package models defines data structures for the RSS reader application
package models

import "time"

// Feed represents an RSS feed
type Feed struct {
	URL   string
	Title string
}

// FeedItem represents an item from an RSS feed
type FeedItem struct {
	Title         string
	Link          string
	Description   string
	Published     string    // formatted for display
	FeedTitle     string
	PublishedTime time.Time // used for sorting, not shown in template
	Read          bool      // true if read, false if unread
	FeedURLOrigin string    // URL of the feed this item came from
}

// PageData holds the data for our templates
type PageData struct {
	Feeds          []Feed
	FeedItems      []FeedItem
	Error          string
	Filter         string // "all" or "unread"
	BaseURL        string // e.g., "/"
	CurrentFeedURL string // To highlight the active feed filter
}
