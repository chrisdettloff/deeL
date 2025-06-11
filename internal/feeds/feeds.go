// Package feeds handles operations related to RSS feeds
package feeds

import (
	"log"
	"sort"
	"time"

	"github.com/mmcdole/gofeed"

	"deel/internal/database" // Corrected import path
	"deel/internal/models"
	"deel/internal/utils"
)

// Manager handles feed operations
type Manager struct {
	DB        *database.DB // Changed db.DB to database.DB
	Feeds     []models.Feed
	FeedItems []models.FeedItem
}

// NewManager creates a new feed manager
func NewManager(db *database.DB) (*Manager, error) { // Changed db *db.DB to db *database.DB
	feeds, err := db.LoadFeeds()
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		DB:    db,
		Feeds: feeds,
	}

	// Initialize feed items
	manager.RefreshFeeds()

	return manager, nil
}

// RefreshFeeds updates the feed items from all feeds
func (m *Manager) RefreshFeeds() {
	m.FeedItems = []models.FeedItem{}

	fp := gofeed.NewParser()
	for _, feed := range m.Feeds {
		parsedFeed, err := fp.ParseURL(feed.URL)
		if err != nil {
			log.Printf("Error refreshing feed %s: %v", feed.URL, err)
			continue
		}

		for _, item := range parsedFeed.Items {
			var pubTime time.Time
			var formatted string

			if item.PublishedParsed != nil {
				pubTime = *item.PublishedParsed
				formatted = pubTime.Format("Jan 2, 2006 15:04")
			} else if item.UpdatedParsed != nil {
				pubTime = *item.UpdatedParsed
				formatted = pubTime.Format("Jan 2, 2006 15:04")
			} else if item.Published != "" {
				parsed, pErr := utils.ParseDate(item.Published)
				if pErr == nil {
					pubTime = parsed
					formatted = pubTime.Format("Jan 2, 2006 15:04")
				} else {
					formatted = item.Published
				}
			} else if item.Updated != "" {
				parsed, pErr := utils.ParseDate(item.Updated)
				if pErr == nil {
					pubTime = parsed
					formatted = pubTime.Format("Jan 2, 2006 15:04")
				} else {
					formatted = item.Updated
				}
			}

			m.FeedItems = append(m.FeedItems, models.FeedItem{
				Title:         item.Title,
				Link:          item.Link,
				Description:   item.Description,
				Published:     formatted,
				FeedTitle:     parsedFeed.Title,
				PublishedTime: pubTime,
				Read:          m.DB.GetFeedItemReadStatus(item.Link),
				Favorite:      m.DB.GetFeedItemFavoriteStatus(item.Link), // Add this line
				FeedURLOrigin: feed.URL,
			})
		}
	}
	m.SortFeedItemsByDate()
	m.UpdateUnreadCounts()
}

// AddFeed adds a new feed
func (m *Manager) AddFeed(feedURL string) (*models.Feed, error) {
	// Parse the feed to get its title
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return nil, err
	}

	// Check if feed already exists
	for _, f := range m.Feeds {
		if f.URL == feedURL {
			return nil, nil // Feed already exists
		}
	}

	// Add the new feed
	newFeed := models.Feed{
		URL:   feedURL,
		Title: feed.Title,
	}
	m.Feeds = append(m.Feeds, newFeed)

	// Save to database
	if err := m.DB.SaveFeed(newFeed); err != nil {
		log.Printf("Error saving feed to database: %v", err)
		return nil, err
	}

	// Add the feed items
	for _, item := range feed.Items {
		var pubTime time.Time
		var formatted string

		if item.PublishedParsed != nil {
			pubTime = *item.PublishedParsed
			formatted = pubTime.Format("Jan 2, 2006 15:04")
		} else if item.UpdatedParsed != nil {
			pubTime = *item.UpdatedParsed
			formatted = pubTime.Format("Jan 2, 2006 15:04")
		} else if item.Published != "" {
			parsed, pErr := utils.ParseDate(item.Published)
			if pErr == nil {
				pubTime = parsed
				formatted = pubTime.Format("Jan 2, 2006 15:04")
			} else {
				formatted = item.Published
			}
		} else if item.Updated != "" {
			parsed, pErr := utils.ParseDate(item.Updated)
			if pErr == nil {
				pubTime = parsed
				formatted = pubTime.Format("Jan 2, 2006 15:04")
			} else {
				formatted = item.Updated
			}
		}

		m.FeedItems = append(m.FeedItems, models.FeedItem{
			Title:         item.Title,
			Link:          item.Link,
			Description:   item.Description,
			Published:     formatted,
			FeedTitle:     feed.Title,
			PublishedTime: pubTime,
			Read:          m.DB.GetFeedItemReadStatus(item.Link),
			Favorite:      m.DB.GetFeedItemFavoriteStatus(item.Link), // Add this line
			FeedURLOrigin: newFeed.URL,
		})
	}
	m.SortFeedItemsByDate()
	m.UpdateUnreadCounts()

	return &newFeed, nil
}

// RemoveFeed removes a feed
func (m *Manager) RemoveFeed(feedURL string) error {
	for i, feed := range m.Feeds {
		if feed.URL == feedURL {
			// Remove from slice
			m.Feeds = append(m.Feeds[:i], m.Feeds[i+1:]...)
			// Remove from database
			if err := m.DB.RemoveFeed(feedURL); err != nil {
				log.Printf("Error removing feed from database: %v", err)
				return err
			}
			break
		}
	}

	// Refresh feed items to reflect the change
	m.RefreshFeeds()
	return nil
}

// ToggleReadStatus toggles the read/unread status of a single feed item
func (m *Manager) ToggleReadStatus(itemLink string) error {
	currentStatus := m.DB.GetFeedItemReadStatus(itemLink)
	newStatus := !currentStatus
	err := m.DB.SetFeedItemReadStatus(itemLink, newStatus)
	if err != nil {
		log.Printf("Error toggling read status for %s: %v", itemLink, err)
		return err
	}

	// Update in-memory feedItems for immediate reflection
	for i, item := range m.FeedItems {
		if item.Link == itemLink {
			m.FeedItems[i].Read = newStatus
			break
		}
	}
	m.UpdateUnreadCounts()
	return nil
}

// ToggleFavoriteStatus toggles the favorite status of a single feed item
func (m *Manager) ToggleFavoriteStatus(itemLink string) error {
	currentStatus := m.DB.GetFeedItemFavoriteStatus(itemLink)
	newStatus := !currentStatus
	err := m.DB.SetFeedItemFavoriteStatus(itemLink, newStatus)
	if err != nil {
		log.Printf("Error toggling favorite status for %s: %v", itemLink, err)
		return err
	}

	// Update in-memory feedItems for immediate reflection
	for i, item := range m.FeedItems {
		if item.Link == itemLink {
			m.FeedItems[i].Favorite = newStatus
			break
		}
	}
	// Note: Unlike read status, unread counts are not affected by favoriting.
	return nil
}

// MarkAllRead marks all currently unread feed items as read
func (m *Manager) MarkAllRead() error {
	for i, item := range m.FeedItems {
		if !item.Read {
			err := m.DB.SetFeedItemReadStatus(item.Link, true)
			if err != nil {
				log.Printf("Error marking item %s as read during mark all: %v", item.Link, err)
				// Continue trying to mark others
			}
			m.FeedItems[i].Read = true // Update in-memory representation
		}
	}
	m.UpdateUnreadCounts()
	return nil
}

// GetFilteredItems returns filtered feed items based on the provided criteria
func (m *Manager) GetFilteredItems(filter, feedURL string) []models.FeedItem {
	var itemsToDisplay []models.FeedItem
	
	// Start with all feed items
	workingItemsList := m.FeedItems

	// 1. Filter by feed URL if provided
	if feedURL != "" {
		var tempItems []models.FeedItem
		for _, item := range workingItemsList {
			if item.FeedURLOrigin == feedURL {
				tempItems = append(tempItems, item)
			}
		}
		workingItemsList = tempItems // Update working list
	}

	// 2. Then, filter by read/unread/favorite status on the (potentially feed-filtered) list
	if filter == "unread" {
		for _, item := range workingItemsList {
			if !item.Read {
				itemsToDisplay = append(itemsToDisplay, item)
			}
		}
	} else if filter == "favorites" { // Add this condition
		for _, item := range workingItemsList {
			if item.Favorite {
				itemsToDisplay = append(itemsToDisplay, item)
			}
		}
	} else { // "all" (or no read/unread filter)
		itemsToDisplay = make([]models.FeedItem, len(workingItemsList))
		copy(itemsToDisplay, workingItemsList)
	}
	
	return itemsToDisplay
}

// SortFeedItemsByDate sorts feedItems in place by PublishedTime (descending)
func (m *Manager) SortFeedItemsByDate() {
	sort.Slice(m.FeedItems, func(i, j int) bool {
		return m.FeedItems[i].PublishedTime.After(m.FeedItems[j].PublishedTime)
	})
}

// UpdateUnreadCounts calculates and updates the unread count for each feed
func (m *Manager) UpdateUnreadCounts() {
	// Reset all counts
	for i := range m.Feeds {
		m.Feeds[i].UnreadCount = 0
	}
	
	// Count unread items for each feed
	for _, item := range m.FeedItems {
		if !item.Read {
			for i := range m.Feeds {
				if m.Feeds[i].URL == item.FeedURLOrigin {
					m.Feeds[i].UnreadCount++
					break
				}
			}
		}
	}
}
