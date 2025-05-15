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
	Published     string // formatted for display
	FeedTitle     string
	PublishedTime time.Time // used for sorting, not shown in template
}

// PageData holds the data for our templates
type PageData struct {
	Feeds     []Feed
	FeedItems []FeedItem
	Error     string
}

var (
	feeds     []Feed
	feedItems []FeedItem
	mutex     sync.Mutex
	templates *template.Template
	db        *bolt.DB
)

const (
	dbPath     = "rss_feeds.db"
	bucketName = "feeds"
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

	// Start the server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Feeds:     feeds,
		FeedItems: feedItems,
	}
	templates.ExecuteTemplate(w, "index.html", data)
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
		published := ""
		var pubTime time.Time
		if item.Published != "" {
			published = item.Published
			pubTime, _ = parseDate(item.Published)
		} else if item.Updated != "" {
			published = item.Updated
			pubTime, _ = parseDate(item.Updated)
		}
		// Format for display
		formatted := ""
		if !pubTime.IsZero() {
			formatted = pubTime.Format("Jan 2, 2006 15:04")
		} else {
			formatted = published
		}
		feedItems = append(feedItems, FeedItem{
			Title:         item.Title,
			Link:          item.Link,
			Description:   item.Description,
			Published:     formatted,
			FeedTitle:     feed.Title,
			PublishedTime: pubTime,
		})
	}
	sortFeedItemsByDate(feedItems)
	mutex.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshFeeds()
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func refreshFeeds() {
	mutex.Lock()
	defer mutex.Unlock()

	// Clear current items
	feedItems = []FeedItem{}

	// Refresh all feeds
	fp := gofeed.NewParser()
	for _, feed := range feeds {
		parsedFeed, err := fp.ParseURL(feed.URL)
		if err != nil {
			log.Printf("Error refreshing feed %s: %v", feed.URL, err)
			continue
		}

		for _, item := range parsedFeed.Items {
			published := ""
			var pubTime time.Time
			if item.Published != "" {
				published = item.Published
				pubTime, _ = parseDate(item.Published)
			} else if item.Updated != "" {
				published = item.Updated
				pubTime, _ = parseDate(item.Updated)
			}
			formatted := ""
			if !pubTime.IsZero() {
				formatted = pubTime.Format("Jan 2, 2006 15:04")
			} else {
				formatted = published
			}
			feedItems = append(feedItems, FeedItem{
				Title:         item.Title,
				Link:          item.Link,
				Description:   item.Description,
				Published:     formatted,
				FeedTitle:     parsedFeed.Title,
				PublishedTime: pubTime,
			})
		}
	}
	sortFeedItemsByDate(feedItems)
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
