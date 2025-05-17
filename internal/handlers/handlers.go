// Package handlers implements HTTP request handlers for the RSS reader
package handlers

import (
	"html/template"
	"log"
	"net/http"
	"sync"

	"deel/internal/feeds"
	"deel/internal/models"
)

// Handler encapsulates the dependencies for HTTP handlers
type Handler struct {
	FeedManager *feeds.Manager
	Templates   *template.Template
	Mutex       *sync.Mutex
}

// NewHandler creates a new Handler
func NewHandler(feedManager *feeds.Manager, templates *template.Template) *Handler {
	return &Handler{
		FeedManager: feedManager,
		Templates:   templates,
		Mutex:       &sync.Mutex{},
	}
}

// HandleIndex handles the index page request
func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	currentFilter := r.URL.Query().Get("filter") // read/unread filter
	if currentFilter == "" {
		currentFilter = "all" 
	}
	currentFeedURLFilter := r.URL.Query().Get("feedURL") // feed source filter

	itemsToDisplay := h.FeedManager.GetFilteredItems(currentFilter, currentFeedURLFilter)
	
	data := models.PageData{
		Feeds:          h.FeedManager.Feeds,
		FeedItems:      itemsToDisplay,
		Filter:         currentFilter,
		BaseURL:        r.URL.Path, 
		CurrentFeedURL: currentFeedURLFilter,
	}
	
	err := h.Templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// HandleAddFeed handles adding a new feed
func (h *Handler) HandleAddFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	feedURL := r.FormValue("feed_url")
	if feedURL == "" {
		data := models.PageData{
			Feeds:     h.FeedManager.Feeds,
			FeedItems: h.FeedManager.FeedItems,
			Error:     "Feed URL cannot be empty",
		}
		h.Templates.ExecuteTemplate(w, "index.html", data)
		return
	}

	h.Mutex.Lock()
	feed, err := h.FeedManager.AddFeed(feedURL)
	h.Mutex.Unlock()

	if err != nil {
		data := models.PageData{
			Feeds:     h.FeedManager.Feeds,
			FeedItems: h.FeedManager.FeedItems,
			Error:     "Failed to parse feed: " + err.Error(),
		}
		h.Templates.ExecuteTemplate(w, "index.html", data)
		return
	}

	// If feed is nil, it means the feed already exists
	if feed == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// HandleRefresh handles refreshing the feeds
func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	h.Mutex.Lock()
	h.FeedManager.RefreshFeeds()
	h.Mutex.Unlock()
	
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther) // Redirect back, preserving filters
}

// HandleRemoveFeed handles removing a feed
func (h *Handler) HandleRemoveFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	feedURL := r.FormValue("feed_url")
	if feedURL != "" {
		h.Mutex.Lock()
		h.FeedManager.RemoveFeed(feedURL)
		h.Mutex.Unlock()
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther) // Redirect back
}

// HandleToggleReadStatus handles toggling the read status of a feed item
func (h *Handler) HandleToggleReadStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusMethodNotAllowed)
		return
	}

	itemLink := r.FormValue("link")
	if itemLink == "" {
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther) // Redirect back if link is missing
		return
	}

	h.Mutex.Lock()
	err := h.FeedManager.ToggleReadStatus(itemLink)
	h.Mutex.Unlock()

	if err != nil {
		log.Printf("Error toggling read status for %s: %v", itemLink, err)
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther) // Redirect back to the previous page
}

// HandleMarkAllRead handles marking all feed items as read
func (h *Handler) HandleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusMethodNotAllowed)
		return
	}

	h.Mutex.Lock()
	h.FeedManager.MarkAllRead()
	h.Mutex.Unlock()

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
