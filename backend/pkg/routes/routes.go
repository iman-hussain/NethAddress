package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/iman-hussain/nethaddress/backend/pkg/cache"
	"github.com/iman-hussain/nethaddress/backend/pkg/handlers"
)

// Build-time variables (set by main.go during initialization)
var (
	BuildCommit         = "unknown"
	BuildDate           = "unknown"
	FrontendBuildCommit = "unknown"
	FrontendBuildDate   = "unknown"
)

// SetBuildInfo sets the build information variables
func SetBuildInfo(commit, date string) {
	BuildCommit = commit
	BuildDate = date
}

// SetFrontendBuildInfo sets the frontend build information variables
func SetFrontendBuildInfo(commit, date string) {
	FrontendBuildCommit = commit
	FrontendBuildDate = date
}

// Router holds all HTTP handlers
type Router struct {
	propertyHandler *handlers.PropertyHandler
	searchHandler   *handlers.SearchHandler
	cacheService    *cache.CacheService
}

// NewRouter creates a new router with all handlers
func NewRouter(
	propertyHandler *handlers.PropertyHandler,
	searchHandler *handlers.SearchHandler,
	cacheService *cache.CacheService,
) *Router {
	return &Router{
		propertyHandler: propertyHandler,
		searchHandler:   searchHandler,
		cacheService:    cacheService,
	}
}

// SetupRoutes configures all HTTP routes
func (router *Router) SetupRoutes(mux *http.ServeMux) {
	// Health check
	mux.HandleFunc("/healthz", handleHealthCheck)
	mux.HandleFunc("/build-info", handleBuildInfo)
	mux.HandleFunc("/", handleRoot)

	// Admin endpoints
	mux.HandleFunc("/admin/cache/flush", router.handleCacheFlush)

	// Legacy endpoint (backward compatibility)
	mux.HandleFunc("/search", router.searchHandler.HandleSearch)

	// New comprehensive API endpoints - longest paths first for proper matching
	mux.HandleFunc("/api/property/analysis", router.propertyHandler.HandleGetFullAnalysis)
	mux.HandleFunc("/api/property/scores", router.propertyHandler.HandleGetPropertyScores)
	mux.HandleFunc("/api/property/recommendations", router.propertyHandler.HandleGetRecommendations)
	mux.HandleFunc("/api/property/solar", router.propertyHandler.HandleCheckSolarEligibility)
	mux.HandleFunc("/api/property", router.propertyHandler.HandleGetPropertyData)
}

// Health check endpoint
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Only accept exact path match
	if r.URL.Path != "/healthz" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "nethaddress-backend",
	})
}

// Cache flush endpoint (admin)
func (router *Router) handleCacheFlush(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin/cache/flush" {
		http.NotFound(w, r)
		return
	}

	// Only allow POST requests
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "method not allowed, use POST",
		})
		return
	}
	// Check for strict admin secret
	adminSecret := os.Getenv("ADMIN_SECRET")
	authHeader := r.Header.Get("X-Admin-Secret")

	// If ADMIN_SECRET is not set in env, fail secure (deny all)
	if adminSecret == "" {
		log.Println("⚠️  Admin flush attempted but ADMIN_SECRET not configured")
		http.Error(w, "Endpoint configuration error", http.StatusForbidden)
		return
	}

	if authHeader != adminSecret {
		log.Printf("⚠️  Unauthorized cache flush attempt from %s", r.RemoteAddr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if router.cacheService == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "cache service not available",
		})
		return
	}

	if err := router.cacheService.FlushAll(); err != nil {
		log.Printf("Cache flush error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "failed to flush cache",
		})
		return
	}

	log.Println("✅ Cache flushed successfully via admin endpoint")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "cache flushed successfully",
	})
}

// Build info endpoint
func handleBuildInfo(w http.ResponseWriter, r *http.Request) {
	// Only accept exact path match
	if r.URL.Path != "/build-info" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"backend": map[string]string{
			"commit": BuildCommit,
			"date":   BuildDate,
		},
		"frontend": map[string]string{
			"commit": FrontendBuildCommit,
			"date":   FrontendBuildDate,
		},
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
		"service": "NethAddress API",
		"version": "2.0",
		"endpoints": map[string]string{
			"GET /healthz":                      "Health check",
			"GET /build-info":                   "Build information",
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
