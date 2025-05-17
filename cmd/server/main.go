// Package main is the entry point for the RSS reader application
package main

import (
	"html/template"
	"log"
	"net/http"

	"deel/internal/database"
	"deel/internal/feeds"
	"deel/internal/handlers"
)

func main() {
	// Initialize templates
	templates := template.Must(template.ParseFiles("templates/index.html"))

	// Initialize database
	db, err := database.NewDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize feed manager
	feedManager, err := feeds.NewManager(db)
	if err != nil {
		log.Fatalf("Failed to initialize feed manager: %v", err)
	}

	// Initialize handler
	handler := handlers.NewHandler(feedManager, templates)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Set up routes
	http.HandleFunc("/", handler.HandleIndex)
	http.HandleFunc("/add", handler.HandleAddFeed)
	http.HandleFunc("/refresh", handler.HandleRefresh)
	http.HandleFunc("/remove", handler.HandleRemoveFeed)
	http.HandleFunc("/toggle-read", handler.HandleToggleReadStatus)
	http.HandleFunc("/mark-all-read", handler.HandleMarkAllRead)

	// Start the server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
