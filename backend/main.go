package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
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

// Build-time variables (injected by ldflags during build)
var (
	BuildCommit         = "unknown"
	BuildDate           = "unknown"
	FrontendBuildCommit = "unknown"
	FrontendBuildDate   = "unknown"
)

func populateBuildMetadata() {
	if BuildCommit == "unknown" || BuildDate == "unknown" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					if BuildCommit == "unknown" {
						BuildCommit = strings.TrimSpace(setting.Value)
					}
				case "vcs.time":
					if BuildDate == "unknown" {
						BuildDate = setting.Value
					}
				}
			}
		}
	}

	if BuildCommit == "unknown" {
		// Try git commands from current directory first
		if commit, err := runGitCommand("rev-parse", "HEAD"); err == nil && commit != "" {
			BuildCommit = commit
		} else {
			// If that fails, try from parent directory (assuming backend is in subdirectory)
			oldDir, _ := os.Getwd()
			os.Chdir("..")
			if commit, err := runGitCommand("rev-parse", "HEAD"); err == nil && commit != "" {
				BuildCommit = commit
			}
			os.Chdir(oldDir) // Restore original directory
		}
	}

	if BuildDate == "unknown" {
		// Try git commands from current directory first
		if date, err := runGitCommand("log", "-1", "--format=%cI"); err == nil && date != "" {
			BuildDate = date
		} else {
			// If that fails, try from parent directory
			oldDir, _ := os.Getwd()
			os.Chdir("..")
			if date, err := runGitCommand("log", "-1", "--format=%cI"); err == nil && date != "" {
				BuildDate = date
			}
			os.Chdir(oldDir) // Restore original directory
		}
	}
}

func runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func main() {
	log.Println("Starting AddressIQ backend...")
	log.Printf("Initial build metadata: commit=%s, date=%s", BuildCommit, BuildDate)
	populateBuildMetadata()
	log.Printf("Final build metadata: commit=%s, date=%s", BuildCommit, BuildDate)

	// Debug: force set build info for testing
	if BuildCommit == "unknown" {
		BuildCommit = "test-commit-123"
	}
	if BuildDate == "unknown" {
		BuildDate = "2025-11-04T15:00:00Z"
	}
	log.Printf("After debug set: commit=%s, date=%s", BuildCommit, BuildDate)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Warning: could not load config: %v, using defaults", err)
		cfg = &config.Config{}
	} else {
		log.Println("✅ Configuration loaded successfully")
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

	// CORS middleware
	allowedOrigins := []string{"http://localhost:3000"}
	if cfg.FrontendOrigin != "" {
		allowedOrigins = append(allowedOrigins, cfg.FrontendOrigin)
	}
	log.Printf("🌐 CORS allowed origins: %v", allowedOrigins)
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
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
	log.Printf("📡 Server will listen on port %s", port)

	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
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

	log.Printf("✅ Server ready, listening on 0.0.0.0:%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
