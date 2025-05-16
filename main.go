package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	bolt "go.etcd.io/bbolt"
)

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

var (
	feeds     []Feed
	feedItems []FeedItem
	mutex     sync.Mutex
	templates *template.Template
	db        *bolt.DB
)

const (
	dbPath                   = "rss_feeds.db"
	bucketName               = "feeds"
	feedItemStatusBucketName = "feedItemStatus" // New bucket for read statuses
)

func init() {
	templates = template.Must(template.ParseFiles("templates/index.html"))

	// Open the database
	var err error
	db, err = bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Create bucket if it doesn't exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		// Create bucket for feed item statuses if it doesn't exist
		_, err = tx.CreateBucketIfNotExists([]byte(feedItemStatusBucketName))
		return err
	})
	if err != nil {
		log.Fatalf("Failed to create bucket: %v", err)
	}

	// Load feeds from database
	loadFeedsFromDB()

	// Initialize feed items
	refreshFeeds()
}

func loadFeedsFromDB() {
	feeds = []Feed{} // Clear the feeds slice

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))

		return b.ForEach(func(k, v []byte) error {
			var feed Feed
			if err := json.Unmarshal(v, &feed); err != nil {
				return err
			}
			feeds = append(feeds, feed)
			return nil
		})
	})

	if err != nil {
		log.Printf("Error loading feeds from database: %v", err)
	}
}

func saveFeedToDB(feed Feed) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))

		encoded, err := json.Marshal(feed)
		if err != nil {
			return err
		}

		return b.Put([]byte(feed.URL), encoded)
	})
}

func removeFeedFromDB(feedURL string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Delete([]byte(feedURL))
	})
}

// getFeedItemReadStatus retrieves the read status of a feed item from the database.
// It defaults to false (unread) if the item is not found.
func getFeedItemReadStatus(link string) bool {
	var isRead bool
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(feedItemStatusBucketName))
		val := b.Get([]byte(link))
		if val != nil && string(val) == "true" {
			isRead = true
		}
		return nil
	})
	if err != nil {
		log.Printf("Error getting read status for %s: %v", link, err)
		return false // Default to unread on error
	}
	return isRead
}

// setFeedItemReadStatus stores the read status of a feed item in the database.
func setFeedItemReadStatus(link string, read bool) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(feedItemStatusBucketName))
		val := "false"
		if read {
			val = "true"
		}
		return b.Put([]byte(link), []byte(val))
	})
}

func main() {
	defer db.Close()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Set up routes
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/add", handleAddFeed)
	http.HandleFunc("/refresh", handleRefresh)
	http.HandleFunc("/remove", handleRemoveFeed)
	http.HandleFunc("/toggle-read", handleToggleReadStatus) // New handler
	http.HandleFunc("/mark-all-read", handleMarkAllRead)   // New handler

	// Start the server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	currentFilter := r.URL.Query().Get("filter") // read/unread filter
	if currentFilter == "" {
		currentFilter = "all" 
	}
	currentFeedURLFilter := r.URL.Query().Get("feedURL") // feed source filter

	var itemsToDisplay []FeedItem
	
	// Start with all feed items
	workingItemsList := feedItems

	// 1. Filter by feed URL if provided
	if currentFeedURLFilter != "" {
		var tempItems []FeedItem
		for _, item := range workingItemsList {
			if item.FeedURLOrigin == currentFeedURLFilter {
				tempItems = append(tempItems, item)
			}
		}
		workingItemsList = tempItems // Update working list
	}

	// 2. Then, filter by read/unread status on the (potentially feed-filtered) list
	if currentFilter == "unread" {
		for _, item := range workingItemsList {
			if !item.Read {
				itemsToDisplay = append(itemsToDisplay, item)
			}
		}
	} else { // "all" (or no read/unread filter)
		itemsToDisplay = make([]FeedItem, len(workingItemsList))
		copy(itemsToDisplay, workingItemsList)
	}
	
	// The items are sorted by date when feeds are added/refreshed.
	// This order should be preserved through the filtering steps.

	data := PageData{
		Feeds:          feeds,
		FeedItems:      itemsToDisplay,
		Filter:         currentFilter,
		BaseURL:        r.URL.Path, 
		CurrentFeedURL: currentFeedURLFilter,
	}
	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

func handleAddFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	feedURL := r.FormValue("feed_url")
	if feedURL == "" {
		data := PageData{
			Feeds:     feeds,
			FeedItems: feedItems,
			Error:     "Feed URL cannot be empty",
		}
		templates.ExecuteTemplate(w, "index.html", data)
		return
	}

	// Parse the feed to get its title
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		data := PageData{
			Feeds:     feeds,
			FeedItems: feedItems,
			Error:     "Failed to parse feed: " + err.Error(),
		}
		templates.ExecuteTemplate(w, "index.html", data)
		return
	}

	mutex.Lock()
	// Check if feed already exists
	for _, f := range feeds {
		if f.URL == feedURL {
			mutex.Unlock()
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	// Add the new feed
	newFeed := Feed{
		URL:   feedURL,
		Title: feed.Title,
	}
	feeds = append(feeds, newFeed)

	// Save to database
	if err := saveFeedToDB(newFeed); err != nil {
		log.Printf("Error saving feed to database: %v", err)
	}

	// Add the feed items
	for _, item := range feed.Items {
		// var published string // Not used directly here, formatted is used
		var pubTime time.Time
		var formatted string

		if item.PublishedParsed != nil {
			pubTime = *item.PublishedParsed
			formatted = pubTime.Format("Jan 2, 2006 15:04")
		} else if item.UpdatedParsed != nil {
			pubTime = *item.UpdatedParsed
			formatted = pubTime.Format("Jan 2, 2006 15:04")
		} else if item.Published != "" {
			parsed, pErr := parseDate(item.Published)
			if pErr == nil {
				pubTime = parsed
				formatted = pubTime.Format("Jan 2, 2006 15:04")
			} else {
				formatted = item.Published 
			}
		} else if item.Updated != "" {
			parsed, pErr := parseDate(item.Updated)
			if pErr == nil {
				pubTime = parsed
				formatted = pubTime.Format("Jan 2, 2006 15:04")
			} else {
				formatted = item.Updated 
			}
		}

		feedItems = append(feedItems, FeedItem{
			Title:         item.Title,
			Link:          item.Link,
			Description:   item.Description,
			Published:     formatted,
			FeedTitle:     feed.Title,
			PublishedTime: pubTime,
			Read:          getFeedItemReadStatus(item.Link), 
			FeedURLOrigin: newFeed.URL,                      
		})
	}
	sortFeedItemsByDate(feedItems)
	mutex.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshFeeds()
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther) // Redirect back, preserving filters
}

func handleRemoveFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	feedURL := r.FormValue("feed_url")
	if feedURL != "" {
		mutex.Lock()
		for i, feed := range feeds {
			if feed.URL == feedURL {
				// Remove from slice
				feeds = append(feeds[:i], feeds[i+1:]...)
				// Remove from database
				if err := removeFeedFromDB(feedURL); err != nil {
					log.Printf("Error removing feed from database: %v", err)
				}
				break
			}
		}
		mutex.Unlock()

		// Refresh feed items to reflect the change
		refreshFeeds()
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther) // Redirect back
}

func refreshFeeds() {
	mutex.Lock()
	defer mutex.Unlock()

	feedItems = []FeedItem{}

	fp := gofeed.NewParser()
	for _, feed := range feeds {
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
				parsed, pErr := parseDate(item.Published)
				if pErr == nil {
					pubTime = parsed
					formatted = pubTime.Format("Jan 2, 2006 15:04")
				} else {
					formatted = item.Published 
				}
		} else if item.Updated != "" {
				parsed, pErr := parseDate(item.Updated)
				if pErr == nil {
					pubTime = parsed
					formatted = pubTime.Format("Jan 2, 2006 15:04")
				} else {
					formatted = item.Updated 
				}
			}
			
			feedItems = append(feedItems, FeedItem{
				Title:         item.Title,
				Link:          item.Link,
				Description:   item.Description,
				Published:     formatted,
				FeedTitle:     parsedFeed.Title, 
				PublishedTime: pubTime,
				Read:          getFeedItemReadStatus(item.Link),
				FeedURLOrigin: feed.URL, 
			})
		}
	}
	sortFeedItemsByDate(feedItems)
}

// handleToggleReadStatus toggles the read/unread status of a single feed item.
func handleToggleReadStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusMethodNotAllowed)
		return
	}

	itemLink := r.FormValue("link")
	if itemLink == "" {
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther) // Redirect back if link is missing
		return
	}

	mutex.Lock()
	currentStatus := getFeedItemReadStatus(itemLink)
	newStatus := !currentStatus
	err := setFeedItemReadStatus(itemLink, newStatus)
	if err != nil {
		log.Printf("Error toggling read status for %s: %v", itemLink, err)
		// Optionally, add error handling to show to user via PageData.Error
	}

	// Update in-memory feedItems for immediate reflection
	for i, item := range feedItems {
		if item.Link == itemLink {
			feedItems[i].Read = newStatus
			break
		}
	}
	mutex.Unlock() // Unlock before redirect

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther) // Redirect back to the previous page
}

// handleMarkAllRead marks all currently unread feed items as read.
func handleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusMethodNotAllowed)
		return
	}

	mutex.Lock()
	for i, item := range feedItems {
		if !item.Read {
			err := setFeedItemReadStatus(item.Link, true)
			if err != nil {
				log.Printf("Error marking item %s as read during mark all: %v", item.Link, err)
				// Continue trying to mark others
			}
			feedItems[i].Read = true // Update in-memory representation
		}
	}
	mutex.Unlock() // Unlock before redirect

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

// parseDate attempts to parse a date string using common RSS/Atom formats
func parseDate(dateStr string) (time.Time, error) {
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

// sortFeedItemsByDate sorts feedItems in place by PublishedTime (descending)
func sortFeedItemsByDate(items []FeedItem) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].PublishedTime.After(items[j].PublishedTime)
	})
}
