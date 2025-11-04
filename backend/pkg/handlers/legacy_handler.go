package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/iman-hussain/AddressIQ/backend/pkg/aggregator"
	"github.com/iman-hussain/AddressIQ/backend/pkg/apiclient"
	"github.com/iman-hussain/AddressIQ/backend/pkg/cache"
	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

// LegacySearchHandler handles legacy /search endpoint for backward compatibility
type LegacySearchHandler struct {
	apiClient  *apiclient.ApiClient
	config     *config.Config
	aggregator *aggregator.PropertyAggregator
}

// NewLegacySearchHandler creates a new legacy search handler
func NewLegacySearchHandler(apiClient *apiclient.ApiClient, cfg *config.Config) *LegacySearchHandler {
	// Try to create cache service, but don't fail if Redis is unavailable
	var cacheService *cache.CacheService
	if cfg.RedisURL != "" {
		cs, err := cache.NewCacheService(cfg.RedisURL)
		if err != nil {
			log.Printf("Warning: Failed to initialize cache: %v", err)
		} else {
			cacheService = cs
		}
	}

	agg := aggregator.NewPropertyAggregator(apiClient, cacheService, cfg)
	return &LegacySearchHandler{
		apiClient:  apiClient,
		config:     cfg,
		aggregator: agg,
	}
}

// APIResult represents the result of a single API call (success or error)
type APIResult struct {
	Name     string      `json:"name"`
	Status   string      `json:"status"` // "success", "error", "not_configured"
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
	Category string      `json:"category"` // "free", "freemium", "premium"
}

// APIResultsGrouped represents grouped API results by category
type APIResultsGrouped struct {
	Free     []APIResult `json:"free"`
	Freemium []APIResult `json:"freemium"`
	Premium  []APIResult `json:"premium"`
}

// ComprehensiveSearchResponse represents complete property data with all API results
type ComprehensiveSearchResponse struct {
	Address     string             `json:"address"`
	Coordinates [2]float64         `json:"coordinates"`
	GeoJSON     string             `json:"geojson"`
	APIResults  APIResultsGrouped  `json:"apiResults"`
}

// HandleSearch handles the legacy /search endpoint
// GET /search?address=<postcode+houseNumber>
// POST /search with form fields postcode and houseNumber
func (h *LegacySearchHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	var postcode, houseNumber string

	// Support both GET with ?address= and POST with form fields
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid form data")
			return
		}
		postcode = r.FormValue("postcode")
		houseNumber = r.FormValue("houseNumber")
		if postcode == "" || houseNumber == "" {
			respondWithError(w, http.StatusBadRequest, "missing postcode or houseNumber")
			return
		}
	} else {
		// GET request with address parameter
		addressParam := r.URL.Query().Get("address")
		if addressParam == "" {
			respondWithError(w, http.StatusBadRequest, "missing address parameter")
			return
		}

		// Parse address parameter (expected format: "3541ED 53" or "3541ED+53")
		parts := strings.Fields(strings.ReplaceAll(addressParam, "+", " "))
		if len(parts) < 2 {
			respondWithError(w, http.StatusBadRequest, "invalid address format, expected: postcode houseNumber")
			return
		}

		postcode = parts[0]
		houseNumber = parts[1]
	}

	log.Printf("Comprehensive search for %s %s", postcode, houseNumber)

	// Fetch basic BAG data first
	bagData, err := h.apiClient.FetchBAGData(postcode, houseNumber)
	if err != nil {
		log.Printf("Error fetching BAG data: %v", err)
		respondWithError(w, http.StatusInternalServerError, "failed to fetch property data")
		return
	}

	if bagData == nil || bagData.Address == "" {
		log.Printf("No BAG data found for %s %s", postcode, houseNumber)
		respondWithError(w, http.StatusNotFound, "address not found")
		return
	}

	// Aggregate all API data
	comprehensiveData, err := h.aggregator.AggregatePropertyData(postcode, houseNumber)
	if err != nil {
		log.Printf("Error aggregating property data: %v", err)
		// Continue with basic BAG data only
		comprehensiveData = &aggregator.ComprehensivePropertyData{
			Address:     bagData.Address,
			Coordinates: bagData.Coordinates,
		}
	}

	// Build API results array
	apiResults := h.buildAPIResults(comprehensiveData)

	// Build response
	response := ComprehensiveSearchResponse{
		Address:     bagData.Address,
		Coordinates: bagData.Coordinates,
		GeoJSON:     bagData.GeoJSON,
		APIResults:  apiResults,
	}

	// Serialize to JSON for embedding in HTML
	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		respondWithError(w, http.StatusInternalServerError, "failed to build response")
		return
	}

	log.Printf("Found address: %s with %d API results", bagData.Address, len(apiResults.Free)+len(apiResults.Freemium)+len(apiResults.Premium))

	// Return HTML response for HTMX - simple structure without IDs that HTMX might hijack
	html := fmt.Sprintf(`
<div data-target="header">
    <div class="box">
        <h5 class="title is-5">%s</h5>
        <p class="is-size-6"><strong>Coordinates:</strong> %.6f, %.6f</p>
        <p class="is-size-6"><strong>Postcode:</strong> %s | <strong>House Number:</strong> %s</p>
		<div class="buttons mt-3">
			<button class="button is-success is-small is-fullwidth" onclick="exportCSV()">Export CSV</button>
            <button class="button is-info is-small is-fullwidth" onclick="openSettings()">
                <span class="icon is-small"><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16"><path d="M8 4.754a3.246 3.246 0 1 0 0 6.492 3.246 3.246 0 0 0 0-6.492zM5.754 8a2.246 2.246 0 1 1 4.492 0 2.246 2.246 0 0 1-4.492 0z"/><path d="M9.796 1.343c-.527-1.79-3.065-1.79-3.592 0l-.094.319a.873.873 0 0 1-1.255.52l-.292-.16c-1.64-.892-3.433.902-2.54 2.541l.159.292a.873.873 0 0 1-.52 1.255l-.319.094c-1.79.527-1.79 3.065 0 3.592l.319.094a.873.873 0 0 1 .52 1.255l-.16.292c-.892 1.64.901 3.434 2.541 2.54l.292-.159a.873.873 0 0 1 1.255.52l.094.319c.527 1.79 3.065 1.79 3.592 0l.094-.319a.873.873 0 0 1 1.255-.52l.292.16c1.64.893 3.434-.902 2.54-2.541l-.159-.292a.873.873 0 0 1 .52-1.255l.319-.094c1.79-.527 1.79-3.065 0-3.592l-.319-.094a.873.873 0 0 1-.52-1.255l.16-.292c.893-1.64-.902-3.433-2.541-2.54l-.292.159a.873.873 0 0 1-1.255-.52l-.094-.319z"/></svg></span>
                <span>Settings</span>
            </button>
        </div>
    </div>
</div>
<div data-target="results">
</div>
<div data-geojson='%s' data-response='%s' style="display:none;"></div>`,
		bagData.Address,
		bagData.Coordinates[1], bagData.Coordinates[0],
		postcode, houseNumber,
		bagData.GeoJSON,
		string(responseJSON))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// buildAPIResults creates grouped API results with status and error info
func (h *LegacySearchHandler) buildAPIResults(data *aggregator.ComprehensivePropertyData) APIResultsGrouped {
	results := APIResultsGrouped{
		Free:     []APIResult{},
		Freemium: []APIResult{},
		Premium:  []APIResult{},
	}

	// Helper function to add result to appropriate category
	addResult := func(name, status, error, category string, data interface{}) {
		result := APIResult{
			Name:     name,
			Status:   status,
			Error:    error,
			Category: category,
			Data:     data,
		}
		switch category {
		case "free":
			results.Free = append(results.Free, result)
		case "freemium":
			results.Freemium = append(results.Freemium, result)
		case "premium":
			results.Premium = append(results.Premium, result)
		}
	}

	// BAG Address - FREE
	addResult("BAG Address", "success", "", "free", map[string]interface{}{"address": data.Address, "coordinates": data.Coordinates})

	// Property & Land Data
	if data.KadasterInfo != nil {
		addResult("Kadaster Object Info", "success", "", "freemium", data.KadasterInfo)
	} else {
		addResult("Kadaster Object Info", "not_configured", "API key not configured", "freemium", nil)
	}

	if data.WOZData != nil {
		addResult("Altum WOZ", "success", "", "freemium", data.WOZData)
	} else {
		addResult("Altum WOZ", "not_configured", "API key not configured", "freemium", nil)
	}

	if data.MarketValuation != nil {
		addResult("Matrixian Property Value+", "success", "", "freemium", data.MarketValuation)
	} else {
		addResult("Matrixian Property Value+", "not_configured", "API key not configured", "freemium", nil)
	}

	if data.TransactionHistory != nil {
		addResult("Altum Transactions", "success", "", "freemium", data.TransactionHistory)
	} else {
		addResult("Altum Transactions", "not_configured", "API key not configured", "freemium", nil)
	}

	// Weather & Climate - FREE
	if data.Weather != nil {
		addResult("KNMI Weather", "success", "", "free", data.Weather)
	} else {
		addResult("KNMI Weather", "error", getErrorMessage(data, "KNMI Weather", "Failed to fetch weather data"), "free", nil)
	}

	if data.SolarPotential != nil {
		addResult("KNMI Solar", "success", "", "free", data.SolarPotential)
	} else {
		addResult("KNMI Solar", "error", getErrorMessage(data, "KNMI Solar", "Failed to fetch solar data"), "free", nil)
	}

	// Environmental Quality - FREE
	if data.AirQuality != nil {
		addResult("Luchtmeetnet Air Quality", "success", "", "free", data.AirQuality)
	} else {
		addResult("Luchtmeetnet Air Quality", "error", getErrorMessage(data, "Luchtmeetnet Air Quality", "Failed to fetch air quality data"), "free", nil)
	}

	if data.NoisePollution != nil {
		addResult("Noise Pollution", "success", "", "freemium", data.NoisePollution)
	} else {
		addResult("Noise Pollution", "not_configured", "API not configured", "freemium", nil)
	}

	// Demographics - FREE
	if data.Population != nil {
		addResult("CBS Population", "success", "", "free", data.Population)
	} else {
		addResult("CBS Population", "error", "Failed to fetch population data", "free", nil)
	}

	if data.SquareStats != nil {
		addResult("CBS Square Statistics", "success", "", "free", data.SquareStats)
	} else {
		addResult("CBS Square Statistics", "error", "Failed to fetch square stats", "free", nil)
	}

	// Soil & Geology
	if data.SoilData != nil {
		addResult("WUR Soil Physicals", "success", "", "freemium", data.SoilData)
	} else {
		addResult("WUR Soil Physicals", "not_configured", "API agreement required", "freemium", nil)
	}

	if data.Subsidence != nil {
		addResult("SkyGeo Subsidence", "success", "", "freemium", data.Subsidence)
	} else {
		addResult("SkyGeo Subsidence", "not_configured", "API key not configured", "freemium", nil)
	}

	if data.SoilQuality != nil {
		addResult("Soil Quality", "success", "", "freemium", data.SoilQuality)
	} else {
		addResult("Soil Quality", "not_configured", "API key required", "freemium", nil)
	}

	if data.BROSoilMap != nil {
		addResult("BRO Soil Map", "success", "", "free", data.BROSoilMap)
	} else {
		addResult("BRO Soil Map", "error", "Failed to fetch BRO data", "free", nil)
	}

	// Energy & Sustainability
	if data.EnergyClimate != nil {
		addResult("Altum Energy & Climate", "success", "", "freemium", data.EnergyClimate)
	} else {
		addResult("Altum Energy & Climate", "not_configured", "API key not configured", "freemium", nil)
	}

	if data.Sustainability != nil {
		addResult("Altum Sustainability", "success", "", "freemium", data.Sustainability)
	} else {
		addResult("Altum Sustainability", "not_configured", "API key not configured", "freemium", nil)
	}

	// Traffic & Mobility
	if data.TrafficData != nil && len(data.TrafficData) > 0 {
		addResult("NDW Traffic", "success", "", "free", data.TrafficData)
	} else {
		addResult("NDW Traffic", "error", "No traffic data available", "free", nil)
	}

	if data.PublicTransport != nil {
		addResult("openOV Public Transport", "success", "", "free", data.PublicTransport)
	} else {
		addResult("openOV Public Transport", "error", "Failed to fetch transport data", "free", nil)
	}

	if data.ParkingData != nil {
		addResult("Parking Availability", "success", "", "freemium", data.ParkingData)
	} else {
		addResult("Parking Availability", "not_configured", "API varies by municipality", "freemium", nil)
	}

	// Water & Safety
	if data.FloodRisk != nil {
		addResult("Flood Risk", "success", "", "free", data.FloodRisk)
	} else {
		addResult("Flood Risk", "error", "Failed to fetch flood risk", "free", nil)
	}

	if data.WaterQuality != nil {
		addResult("Digital Delta Water Quality", "success", "", "freemium", data.WaterQuality)
	} else {
		addResult("Digital Delta Water Quality", "not_configured", "Water authority account required", "freemium", nil)
	}

	if data.Safety != nil {
		addResult("CBS Safety Experience", "success", "", "freemium", data.Safety)
	} else {
		addResult("CBS Safety Experience", "not_configured", "API key required", "freemium", nil)
	}

	if data.SchipholFlights != nil {
		addResult("Schiphol Flight Noise", "success", "", "freemium", data.SchipholFlights)
	} else {
		addResult("Schiphol Flight Noise", "not_configured", "API key not configured", "freemium", nil)
	}

	// Infrastructure & Facilities - FREE
	if data.GreenSpaces != nil {
		addResult("Green Spaces", "success", "", "free", data.GreenSpaces)
	} else {
		addResult("Green Spaces", "error", "Failed to fetch green spaces", "free", nil)
	}

	if data.Education != nil {
		addResult("Education Facilities", "success", "", "free", data.Education)
	} else {
		addResult("Education Facilities", "error", "Failed to fetch education data", "free", nil)
	}

	if data.BuildingPermits != nil {
		addResult("Building Permits", "success", "", "freemium", data.BuildingPermits)
	} else {
		addResult("Building Permits", "not_configured", "API varies by region", "freemium", nil)
	}

	if data.Facilities != nil {
		addResult("Facilities & Amenities", "success", "", "free", data.Facilities)
	} else {
		addResult("Facilities & Amenities", "error", "Failed to fetch facilities", "free", nil)
	}

	if data.Elevation != nil {
		addResult("AHN Height Model", "success", "", "free", data.Elevation)
	} else {
		addResult("AHN Height Model", "error", "Failed to fetch elevation data", "free", nil)
	}

	// Heritage
	if data.MonumentStatus != nil {
		addResult("Monument Status", "success", "", "free", data.MonumentStatus)
	} else {
		addResult("Monument Status", "error", "No monument data (or Amsterdam only)", "free", nil)
	}

	// Comprehensive Platforms - FREE
	if data.PDOKData != nil {
		addResult("PDOK Platform", "success", "", "free", data.PDOKData)
	} else {
		addResult("PDOK Platform", "error", "Failed to fetch PDOK data", "free", nil)
	}

	if data.StratopoEnvironment != nil {
		addResult("Stratopo Environment", "success", "", "freemium", data.StratopoEnvironment)
	} else {
		addResult("Stratopo Environment", "not_configured", "API key not configured", "freemium", nil)
	}

	if data.LandUse != nil {
		addResult("Land Use & Zoning", "success", "", "free", data.LandUse)
	} else {
		addResult("Land Use & Zoning", "error", "Failed to fetch land use data", "free", nil)
	}

	return results
}

func getErrorMessage(data *aggregator.ComprehensivePropertyData, source, fallback string) string {
	if data != nil && data.Errors != nil {
		if msg, ok := data.Errors[source]; ok && msg != "" {
			return msg
		}
	}
	return fallback
}
