package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"ABCScraper/api"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Get port from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create main router
	r := mux.NewRouter()

	// Setup API routes
	api.SetupRoutes(r)

	// Setup CORS for production
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, specify your React Native app's domains
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	})

	handler := c.Handler(r)

	// Start server
	fmt.Printf(" API running on port %s\n", port)
	fmt.Println(" Available endpoints:")
	fmt.Println("   GET  /health")
	fmt.Println("   GET  /api/v1/stores/{zipcode}")

	log.Fatal(http.ListenAndServe(":"+port, handler))
}
