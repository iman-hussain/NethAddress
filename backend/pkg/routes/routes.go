package routes

import (
	"encoding/json"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/handlers"
)

// Router holds all HTTP handlers
type Router struct {
	propertyHandler     *handlers.PropertyHandler
	legacySearchHandler *handlers.LegacySearchHandler
}

// NewRouter creates a new router with all handlers
func NewRouter(
	propertyHandler *handlers.PropertyHandler,
	legacySearchHandler *handlers.LegacySearchHandler,
) *Router {
	return &Router{
		propertyHandler:     propertyHandler,
		legacySearchHandler: legacySearchHandler,
	}
}

// SetupRoutes configures all HTTP routes
func (router *Router) SetupRoutes(mux *http.ServeMux) {
	// Health check
	mux.HandleFunc("/healthz", handleHealthCheck)
	mux.HandleFunc("/", handleRoot)

	// Legacy endpoint (backward compatibility)
	mux.HandleFunc("/search", router.legacySearchHandler.HandleSearch)

	// New comprehensive API endpoints - longest paths first for proper matching
	mux.HandleFunc("/api/property/analysis", router.propertyHandler.HandleGetFullAnalysis)
	mux.HandleFunc("/api/property/scores", router.propertyHandler.HandleGetPropertyScores)
	mux.HandleFunc("/api/property/recommendations", router.propertyHandler.HandleGetRecommendations)
	mux.HandleFunc("/api/property", router.propertyHandler.HandleGetPropertyData)
}

// Health check endpoint
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "addressiq-backend",
	})
}

// Root endpoint
func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "AddressIQ API",
		"version": "2.0",
		"endpoints": map[string]string{
			"GET /healthz":                      "Health check",
			"GET /search":                       "Legacy search endpoint",
			"GET /api/property":                 "Get comprehensive property data",
			"GET /api/property/scores":          "Get property scores (ESG, Profit, Opportunity)",
			"GET /api/property/recommendations": "Get smart recommendations",
			"GET /api/property/analysis":        "Get full analysis (data + scores + recommendations)",
		},
		"query_parameters": map[string]string{
			"postcode":    "Dutch postcode (e.g., 3541ED)",
			"houseNumber": "House number (e.g., 53)",
		},
	})
}
