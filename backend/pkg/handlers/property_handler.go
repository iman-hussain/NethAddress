package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/aggregator"
	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/scoring"
)

// PropertyHandler handles property-related HTTP requests
type PropertyHandler struct {
	aggregator    *aggregator.PropertyAggregator
	scoringEngine *scoring.EnhancedScoringEngine
	config        *config.Config
}

// NewPropertyHandler creates a new property handler
func NewPropertyHandler(
	agg *aggregator.PropertyAggregator,
	scoringEngine *scoring.EnhancedScoringEngine,
	cfg *config.Config,
) *PropertyHandler {
	return &PropertyHandler{
		aggregator:    agg,
		scoringEngine: scoringEngine,
		config:        cfg,
	}
}

// PropertyDataResponse represents the comprehensive property data response
type PropertyDataResponse struct {
	Property *aggregator.ComprehensivePropertyData `json:"property"`
	Error    string                                `json:"error,omitempty"`
}

// PropertyScoresResponse represents the property scores response
type PropertyScoresResponse struct {
	Postcode    string                  `json:"postcode"`
	HouseNumber string                  `json:"houseNumber"`
	Scores      *scoring.PropertyScores `json:"scores"`
	Error       string                  `json:"error,omitempty"`
}

// RecommendationsResponse represents just the recommendations
type RecommendationsResponse struct {
	Postcode        string   `json:"postcode"`
	HouseNumber     string   `json:"houseNumber"`
	Recommendations []string `json:"recommendations"`
	Error           string   `json:"error,omitempty"`
}

// HandleGetPropertyData returns comprehensive aggregated property data
// GET /api/property/:postcode/:houseNumber
func (h *PropertyHandler) HandleGetPropertyData(w http.ResponseWriter, r *http.Request) {
	postcode := r.URL.Query().Get("postcode")
	houseNumber := r.URL.Query().Get("houseNumber")

	if postcode == "" || houseNumber == "" {
		respondWithError(w, http.StatusBadRequest, "missing postcode or houseNumber query parameters")
		return
	}

	log.Printf("Fetching property data for %s %s", postcode, houseNumber)

	data, err := h.aggregator.AggregatePropertyData(postcode, houseNumber)
	if err != nil {
		log.Printf("Error aggregating property data: %v", err)
		respondWithError(w, http.StatusInternalServerError, "failed to aggregate property data")
		return
	}

	respondWithJSON(w, http.StatusOK, PropertyDataResponse{
		Property: data,
	})
}

// HandleGetPropertyScores returns comprehensive property scores
// GET /api/property/:postcode/:houseNumber/scores
func (h *PropertyHandler) HandleGetPropertyScores(w http.ResponseWriter, r *http.Request) {
	postcode := r.URL.Query().Get("postcode")
	houseNumber := r.URL.Query().Get("houseNumber")

	if postcode == "" || houseNumber == "" {
		respondWithError(w, http.StatusBadRequest, "missing postcode or houseNumber query parameters")
		return
	}

	log.Printf("Calculating scores for %s %s", postcode, houseNumber)

	// Get aggregated data first
	data, err := h.aggregator.AggregatePropertyData(postcode, houseNumber)
	if err != nil {
		log.Printf("Error aggregating property data for scoring: %v", err)
		respondWithError(w, http.StatusInternalServerError, "failed to aggregate property data")
		return
	}

	// Calculate scores
	scores := h.scoringEngine.CalculateComprehensiveScores(data)

	respondWithJSON(w, http.StatusOK, PropertyScoresResponse{
		Postcode:    postcode,
		HouseNumber: houseNumber,
		Scores:      scores,
	})
}

// HandleGetRecommendations returns smart recommendations for a property
// GET /api/property/:postcode/:houseNumber/recommendations
func (h *PropertyHandler) HandleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	postcode := r.URL.Query().Get("postcode")
	houseNumber := r.URL.Query().Get("houseNumber")

	if postcode == "" || houseNumber == "" {
		respondWithError(w, http.StatusBadRequest, "missing postcode or houseNumber query parameters")
		return
	}

	log.Printf("Generating recommendations for %s %s", postcode, houseNumber)

	// Get aggregated data
	data, err := h.aggregator.AggregatePropertyData(postcode, houseNumber)
	if err != nil {
		log.Printf("Error aggregating property data for recommendations: %v", err)
		respondWithError(w, http.StatusInternalServerError, "failed to aggregate property data")
		return
	}

	// Calculate scores (includes recommendations)
	scores := h.scoringEngine.CalculateComprehensiveScores(data)

	respondWithJSON(w, http.StatusOK, RecommendationsResponse{
		Postcode:        postcode,
		HouseNumber:     houseNumber,
		Recommendations: scores.Recommendations,
	})
}

// HandleGetFullAnalysis returns everything - data, scores, and recommendations
// GET /api/property/:postcode/:houseNumber/analysis
func (h *PropertyHandler) HandleGetFullAnalysis(w http.ResponseWriter, r *http.Request) {
	postcode := r.URL.Query().Get("postcode")
	houseNumber := r.URL.Query().Get("houseNumber")

	if postcode == "" || houseNumber == "" {
		respondWithError(w, http.StatusBadRequest, "missing postcode or houseNumber query parameters")
		return
	}

	log.Printf("Performing full analysis for %s %s", postcode, houseNumber)

	// Get aggregated data
	data, err := h.aggregator.AggregatePropertyData(postcode, houseNumber)
	if err != nil {
		log.Printf("Error aggregating property data for analysis: %v", err)
		respondWithError(w, http.StatusInternalServerError, "failed to aggregate property data")
		return
	}

	// Calculate scores
	scores := h.scoringEngine.CalculateComprehensiveScores(data)

	// Combine everything
	response := map[string]interface{}{
		"postcode":    postcode,
		"houseNumber": houseNumber,
		"property":    data,
		"scores":      scores,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Utility functions

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, map[string]string{"error": message})
}
