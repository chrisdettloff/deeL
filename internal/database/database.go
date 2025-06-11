// Package database handles database operations for the RSS reader
package database

import (
	"encoding/json"
	"log"
	"time"

	bolt "go.etcd.io/bbolt"

	"deel/internal/models"
)

const (
	// DBPath is the path to the database file
	DBPath = "rss_feeds.db"
	
	// BucketName is the name of the bucket for feeds
	BucketName = "feeds"
	
	// FeedItemStatusBucketName is the name of the bucket for feed item statuses
	FeedItemStatusBucketName = "feedItemStatus"

	// FeedItemFavoriteBucketName is the name of the bucket for feed item favorite statuses
	FeedItemFavoriteBucketName = "feedItemFavorite"
)

// DB wraps the bolt database
type DB struct {
	*bolt.DB
}

// NewDB initializes and returns a new database connection
func NewDB() (*DB, error) {
	db, err := bolt.Open(DBPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	// Create buckets if they don't exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BucketName))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(FeedItemStatusBucketName))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(FeedItemFavoriteBucketName))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &DB{DB: db}, nil
}

// LoadFeeds loads all feeds from the database
func (db *DB) LoadFeeds() ([]models.Feed, error) {
	var feeds []models.Feed

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))

		return b.ForEach(func(k, v []byte) error {
			var feed models.Feed
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

	return feeds, err
}

// SaveFeed saves a feed to the database
func (db *DB) SaveFeed(feed models.Feed) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))

		encoded, err := json.Marshal(feed)
		if err != nil {
			return err
		}

		return b.Put([]byte(feed.URL), encoded)
	})
}

// RemoveFeed removes a feed from the database
func (db *DB) RemoveFeed(feedURL string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		return b.Delete([]byte(feedURL))
	})
}

// GetFeedItemReadStatus retrieves the read status of a feed item from the database.
// It defaults to false (unread) if the item is not found.
func (db *DB) GetFeedItemReadStatus(link string) bool {
	var isRead bool
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(FeedItemStatusBucketName))
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

// SetFeedItemReadStatus sets the read status of a feed item in the database.
func (db *DB) SetFeedItemReadStatus(link string, read bool) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(FeedItemStatusBucketName))
		val := "false"
		if read {
			val = "true"
		}
		return b.Put([]byte(link), []byte(val))
	})
}

// GetFeedItemFavoriteStatus retrieves the favorite status of a feed item from the database.
// It defaults to false (not favorited) if the item is not found.
func (db *DB) GetFeedItemFavoriteStatus(link string) bool {
	var isFavorite bool
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(FeedItemFavoriteBucketName))
		val := b.Get([]byte(link))
		if val != nil {
			isFavorite = (string(val) == "true")
		}
		return nil
	})
	if err != nil {
		log.Printf("Error getting favorite status for %s: %v", link, err)
		return false // Default to false on error
	}
	return isFavorite
}

// SetFeedItemFavoriteStatus sets the favorite status of a feed item in the database.
func (db *DB) SetFeedItemFavoriteStatus(link string, favorite bool) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(FeedItemFavoriteBucketName))
		val := "false"
		if favorite {
			val = "true"
		}
		return b.Put([]byte(link), []byte(val))
	})
}
