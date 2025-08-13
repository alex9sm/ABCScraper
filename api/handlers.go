package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"ABCScraper/scrapers"

	"github.com/gorilla/mux"
)

// APIResponse represents the standard API response format
type APIResponse struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// SetupRoutes configures all API routes
func SetupRoutes(r *mux.Router) {
	// Health check endpoint
	r.HandleFunc("/health", healthHandler).Methods("GET")

	// API v1 routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Store scraping endpoint - matches your desired format
	api.HandleFunc("/stores/{zipcode}", scrapeStoresHandler).Methods("GET")

	// You can add more endpoints here as you expand
	// api.HandleFunc("/products/{productId}", scrapeProductHandler).Methods("GET")
	// api.HandleFunc("/categories/{category}", scrapeCategoryHandler).Methods("GET")
}

// Health check handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Status:    "success",
		Message:   "Scraper API is healthy and running",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Handler for scraping stores by zip code
func scrapeStoresHandler(w http.ResponseWriter, r *http.Request) {
	// Extract zipcode from URL path
	vars := mux.Vars(r)
	zipcode := vars["zipcode"]

	// Validate zipcode format (5 digits)
	if !isValidZipcode(zipcode) {
		sendErrorResponse(w, "Invalid zipcode format. Must be 5 digits.", http.StatusBadRequest)
		return
	}

	// Call your scraper function
	scrapedData, err := scrapers.ScrapeUserStore(zipcode)
	if err != nil {
		sendErrorResponse(w, "Failed to scrape store data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if scrapedData == nil {
		scrapedData = []scrapers.StoreResult{}
	}

	// Send successful response
	response := APIResponse{
		Status:    "success",
		Data:      scrapedData,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper function to validate zipcode format
func isValidZipcode(zipcode string) bool {
	// Match 5 digits exactly
	matched, _ := regexp.MatchString(`^\d{5}$`, zipcode)
	return matched
}

// Helper function to send error responses
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := APIResponse{
		Status:    "error",
		Message:   message,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Example handler for future expansion - uncomment when you add more scrapers
/*
func scrapeProductHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productId := vars["productId"]

	scrapedData, err := scrapers.ScrapeProduct(productId)
	if err != nil {
		sendErrorResponse(w, "Failed to scrape product data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := APIResponse{
		Status:    "success",
		Data:      scrapedData,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
*/
