package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iman-hussain/AddressIQ/backend/pkg/aggregator"
	"github.com/iman-hussain/AddressIQ/backend/pkg/apiclient"
	"github.com/iman-hussain/AddressIQ/backend/pkg/cache"
	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/handlers"
	"github.com/iman-hussain/AddressIQ/backend/pkg/routes"
	"github.com/iman-hussain/AddressIQ/backend/pkg/scoring"
	"github.com/rs/cors"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Warning: could not load config: %v, using defaults", err)
		cfg = &config.Config{}
	}

	// Initialize Redis client for caching
	var cacheService *cache.CacheService
	if cfg.RedisURL != "" {
		var err error
		cacheService, err = cache.NewCacheService(cfg.RedisURL)
		if err != nil {
			log.Printf("Warning: could not connect to Redis: %v, caching disabled", err)
			cacheService = nil
		} else {
			log.Println("✅ Redis caching enabled")
		}
	}

	// Fallback: create a no-op aggregator if cache is not available
	if cacheService == nil {
		log.Println("⚠️  Redis not configured, using direct API calls (no caching)")
	}

	// Initialize API client
	apiClient := apiclient.NewApiClient(&http.Client{
		Timeout: 30 * time.Second,
	})

	// Initialize aggregator
	propertyAggregator := aggregator.NewPropertyAggregator(apiClient, cacheService, cfg)

	// Initialize scoring engine
	scoringEngine := scoring.NewEnhancedScoringEngine()

	// Initialize handlers
	propertyHandler := handlers.NewPropertyHandler(propertyAggregator, scoringEngine, cfg)
	legacySearchHandler := handlers.NewLegacySearchHandler(apiClient, cfg)

	// Initialize router
	router := routes.NewRouter(propertyHandler, legacySearchHandler)

	// Setup routes
	mux := http.NewServeMux()
	router.SetupRoutes(mux)

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			cfg.FRONTEND_ORIGIN,
			"http://localhost:3000",
		},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		Debug:            false,
	})
	handler := c.Handler(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 AddressIQ API server starting on port %s", port)
		log.Printf("📍 Endpoints available:")
		log.Printf("   GET  /                                  - API information")
		log.Printf("   GET  /healthz                           - Health check")
		log.Printf("   GET  /search                            - Legacy search")
		log.Printf("   GET  /api/property                      - Full property data")
		log.Printf("   GET  /api/property/scores               - Property scores")
		log.Printf("   GET  /api/property/recommendations      - Recommendations")
		log.Printf("   GET  /api/property/analysis             - Complete analysis")
		log.Printf("")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if cacheService != nil {
		if err := cacheService.Close(); err != nil {
			log.Printf("Warning: error closing cache: %v", err)
		}
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server shutdown failed: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}
