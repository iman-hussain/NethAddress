package main

import (
	"net/http"
	"os"
	"time"

	"github.com/iman-hussain/nethaddress/backend/pkg/aggregator"
	"github.com/iman-hussain/nethaddress/backend/pkg/apiclient"
	"github.com/iman-hussain/nethaddress/backend/pkg/cache"
	"github.com/iman-hussain/nethaddress/backend/pkg/config"
	"github.com/iman-hussain/nethaddress/backend/pkg/handlers"
	"github.com/iman-hussain/nethaddress/backend/pkg/logutil"
	"github.com/iman-hussain/nethaddress/backend/pkg/routes"
	"github.com/iman-hussain/nethaddress/backend/pkg/scoring"
	"github.com/rs/cors"
)

// Build-time variables (injected by ldflags during build)
var (
	BuildCommit         = "unknown"
	BuildDate           = "unknown"
	FrontendBuildCommit = "unknown"
	FrontendBuildDate   = "unknown"
)

func main() {
	logutil.Info("Starting NethAddress backend...")
	logutil.Infof("Build metadata: commit=%s, date=%s", BuildCommit, BuildDate)

	cfg, err := config.LoadConfig()
	if err != nil {
		logutil.Fatalf("FATAL: could not load config: %v", err)
	}
	logutil.Info("Configuration loaded successfully")

	// Initialize Redis client for caching
	var cacheService *cache.CacheService
	if cfg.RedisURL != "" {
		var err error
		cacheService, err = cache.NewCacheService(cfg.RedisURL)
		if err != nil {
			logutil.Warnf("Warning: could not connect to Redis: %v, caching disabled", err)
			cacheService = nil
		} else {
			logutil.Info("Redis caching enabled")
		}
	}

	// Fallback: create a no-op aggregator if cache is not available
	if cacheService == nil {
		logutil.Warn("Redis not configured, using direct API calls (no caching)")
	}

	// Initialize API client
	apiClient := apiclient.NewApiClient(&http.Client{
		Timeout: 30 * time.Second,
	}, cfg)

	// Initialize aggregator
	propertyAggregator := aggregator.NewPropertyAggregator(apiClient, cacheService, cfg)

	// Initialize scoring engine
	scoringEngine := scoring.NewEnhancedScoringEngine()

	// Initialize handlers
	propertyHandler := handlers.NewPropertyHandler(propertyAggregator, scoringEngine, cfg)
	searchHandler := handlers.NewSearchHandler(apiClient, cfg)

	// Set build info for routes
	routes.SetBuildInfo(BuildCommit, BuildDate)

	// Set frontend build info from environment variables (for production deployment)
	frontendCommit := os.Getenv("FRONTEND_BUILD_COMMIT")
	if frontendCommit == "" {
		frontendCommit = BuildCommit
	}
	frontendDate := os.Getenv("FRONTEND_BUILD_DATE")
	if frontendDate == "" {
		frontendDate = BuildDate
	}
	routes.SetFrontendBuildInfo(frontendCommit, frontendDate)

	// Initialize router
	router := routes.NewRouter(propertyHandler, searchHandler, cacheService)

	// Setup routes
	mux := http.NewServeMux()
	router.SetupRoutes(mux)

	// Register SSE endpoint directly (or add to router)
	// Since router.SetupRoutes might not expose everything, let's check routes.go or just add here.
	// Adding here is safest if we have access to handler instance.
	// But `propertyHandler` and `searchHandler` are local variables.
	mux.HandleFunc("/api/search/stream", searchHandler.HandleSearchStream)

	// CORS middleware
	allowedOrigins := []string{"http://localhost:3000"}
	if cfg.FrontendOrigin != "" {
		allowedOrigins = append(allowedOrigins, cfg.FrontendOrigin)
	}
	logutil.Infof("CORS allowed origins: %v", allowedOrigins)
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Admin-Secret"},
		AllowCredentials: false,
		Debug:            false,
	})
	handler := c.Handler(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logutil.Infof("Server will listen on port %s", port)

	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	logutil.Infof("NethAddress API server starting on port %s", port)
	logutil.Info("Endpoints available:")
	logutil.Info("   GET  /                                  - API information")
	logutil.Info("   GET  /healthz                           - Health check")
	logutil.Info("   GET  /search                            - Legacy search")
	logutil.Info("   GET  /api/search/stream                 - Real-time search stream (SSE)")
	logutil.Info("   GET  /api/property                      - Full property data")
	logutil.Info("   GET  /api/property/scores               - Property scores")
	logutil.Info("   GET  /api/property/recommendations      - Recommendations")
	logutil.Info("   GET  /api/property/analysis             - Complete analysis")

	logutil.Infof("Server ready, listening on 0.0.0.0:%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logutil.Fatalf("Server failed: %v", err)
	}
}
