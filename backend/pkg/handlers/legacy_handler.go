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
	Name   string      `json:"name"`
	Status string      `json:"status"` // "success", "error", "not_configured"
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// ComprehensiveSearchResponse represents complete property data with all API results
type ComprehensiveSearchResponse struct {
	Address     string      `json:"address"`
	Coordinates [2]float64  `json:"coordinates"`
	GeoJSON     string      `json:"geojson"`
	APIResults  []APIResult `json:"apiResults"`
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

	log.Printf("Found address: %s with %d API results", bagData.Address, len(apiResults))

	// Return HTML response for HTMX - simple structure without IDs that HTMX might hijack
	html := fmt.Sprintf(`
<div data-target="header">
    <div class="box">
        <h5 class="title is-5">%s</h5>
        <p class="is-size-7"><strong>Coordinates:</strong> %.6f, %.6f</p>
        <p class="is-size-7"><strong>Postcode:</strong> %s | <strong>House Number:</strong> %s</p>
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

// buildAPIResults creates an array of API results with status and error info
func (h *LegacySearchHandler) buildAPIResults(data *aggregator.ComprehensivePropertyData) []APIResult {
	results := []APIResult{
		{Name: "BAG Address", Status: "success", Data: map[string]interface{}{"address": data.Address, "coordinates": data.Coordinates}},
	}

	// Property & Land Data
	if data.KadasterInfo != nil {
		results = append(results, APIResult{Name: "Kadaster Object Info", Status: "success", Data: data.KadasterInfo})
	} else {
		results = append(results, APIResult{Name: "Kadaster Object Info", Status: "not_configured", Error: "API key not configured"})
	}

	if data.WOZData != nil {
		results = append(results, APIResult{Name: "Altum WOZ", Status: "success", Data: data.WOZData})
	} else {
		results = append(results, APIResult{Name: "Altum WOZ", Status: "not_configured", Error: "API key not configured"})
	}

	if data.MarketValuation != nil {
		results = append(results, APIResult{Name: "Matrixian Property Value+", Status: "success", Data: data.MarketValuation})
	} else {
		results = append(results, APIResult{Name: "Matrixian Property Value+", Status: "not_configured", Error: "API key not configured"})
	}

	if data.TransactionHistory != nil {
		results = append(results, APIResult{Name: "Altum Transactions", Status: "success", Data: data.TransactionHistory})
	} else {
		results = append(results, APIResult{Name: "Altum Transactions", Status: "not_configured", Error: "API key not configured"})
	}

	// Weather & Climate
	if data.Weather != nil {
		results = append(results, APIResult{Name: "KNMI Weather", Status: "success", Data: data.Weather})
	} else {
		results = append(results, APIResult{Name: "KNMI Weather", Status: "error", Error: getErrorMessage(data, "KNMI Weather", "Failed to fetch weather data")})
	}

	if data.SolarPotential != nil {
		results = append(results, APIResult{Name: "KNMI Solar", Status: "success", Data: data.SolarPotential})
	} else {
		results = append(results, APIResult{Name: "KNMI Solar", Status: "error", Error: getErrorMessage(data, "KNMI Solar", "Failed to fetch solar data")})
	}

	// Environmental Quality
	if data.AirQuality != nil {
		results = append(results, APIResult{Name: "Luchtmeetnet Air Quality", Status: "success", Data: data.AirQuality})
	} else {
		results = append(results, APIResult{Name: "Luchtmeetnet Air Quality", Status: "error", Error: getErrorMessage(data, "Luchtmeetnet Air Quality", "Failed to fetch air quality data")})
	}

	if data.NoisePollution != nil {
		results = append(results, APIResult{Name: "Noise Pollution", Status: "success", Data: data.NoisePollution})
	} else {
		results = append(results, APIResult{Name: "Noise Pollution", Status: "not_configured", Error: "API not configured"})
	}

	// Demographics
	if data.Population != nil {
		results = append(results, APIResult{Name: "CBS Population", Status: "success", Data: data.Population})
	} else {
		results = append(results, APIResult{Name: "CBS Population", Status: "error", Error: "Failed to fetch population data"})
	}

	if data.SquareStats != nil {
		results = append(results, APIResult{Name: "CBS Square Statistics", Status: "success", Data: data.SquareStats})
	} else {
		results = append(results, APIResult{Name: "CBS Square Statistics", Status: "error", Error: "Failed to fetch square stats"})
	}

	// Soil & Geology
	if data.SoilData != nil {
		results = append(results, APIResult{Name: "WUR Soil Physicals", Status: "success", Data: data.SoilData})
	} else {
		results = append(results, APIResult{Name: "WUR Soil Physicals", Status: "not_configured", Error: "API agreement required"})
	}

	if data.Subsidence != nil {
		results = append(results, APIResult{Name: "SkyGeo Subsidence", Status: "success", Data: data.Subsidence})
	} else {
		results = append(results, APIResult{Name: "SkyGeo Subsidence", Status: "not_configured", Error: "API key not configured"})
	}

	if data.SoilQuality != nil {
		results = append(results, APIResult{Name: "Soil Quality", Status: "success", Data: data.SoilQuality})
	} else {
		results = append(results, APIResult{Name: "Soil Quality", Status: "not_configured", Error: "API key required"})
	}

	if data.BROSoilMap != nil {
		results = append(results, APIResult{Name: "BRO Soil Map", Status: "success", Data: data.BROSoilMap})
	} else {
		results = append(results, APIResult{Name: "BRO Soil Map", Status: "error", Error: "Failed to fetch BRO data"})
	}

	// Energy & Sustainability
	if data.EnergyClimate != nil {
		results = append(results, APIResult{Name: "Altum Energy & Climate", Status: "success", Data: data.EnergyClimate})
	} else {
		results = append(results, APIResult{Name: "Altum Energy & Climate", Status: "not_configured", Error: "API key not configured"})
	}

	if data.Sustainability != nil {
		results = append(results, APIResult{Name: "Altum Sustainability", Status: "success", Data: data.Sustainability})
	} else {
		results = append(results, APIResult{Name: "Altum Sustainability", Status: "not_configured", Error: "API key not configured"})
	}

	// Traffic & Mobility
	if data.TrafficData != nil && len(data.TrafficData) > 0 {
		results = append(results, APIResult{Name: "NDW Traffic", Status: "success", Data: data.TrafficData})
	} else {
		results = append(results, APIResult{Name: "NDW Traffic", Status: "error", Error: "No traffic data available"})
	}

	if data.PublicTransport != nil {
		results = append(results, APIResult{Name: "openOV Public Transport", Status: "success", Data: data.PublicTransport})
	} else {
		results = append(results, APIResult{Name: "openOV Public Transport", Status: "error", Error: "Failed to fetch transport data"})
	}

	if data.ParkingData != nil {
		results = append(results, APIResult{Name: "Parking Availability", Status: "success", Data: data.ParkingData})
	} else {
		results = append(results, APIResult{Name: "Parking Availability", Status: "not_configured", Error: "API varies by municipality"})
	}

	// Water & Safety
	if data.FloodRisk != nil {
		results = append(results, APIResult{Name: "Flood Risk", Status: "success", Data: data.FloodRisk})
	} else {
		results = append(results, APIResult{Name: "Flood Risk", Status: "error", Error: "Failed to fetch flood risk"})
	}

	if data.WaterQuality != nil {
		results = append(results, APIResult{Name: "Digital Delta Water Quality", Status: "success", Data: data.WaterQuality})
	} else {
		results = append(results, APIResult{Name: "Digital Delta Water Quality", Status: "not_configured", Error: "Water authority account required"})
	}

	if data.Safety != nil {
		results = append(results, APIResult{Name: "CBS Safety Experience", Status: "success", Data: data.Safety})
	} else {
		results = append(results, APIResult{Name: "CBS Safety Experience", Status: "not_configured", Error: "API key required"})
	}

	if data.SchipholFlights != nil {
		results = append(results, APIResult{Name: "Schiphol Flight Noise", Status: "success", Data: data.SchipholFlights})
	} else {
		results = append(results, APIResult{Name: "Schiphol Flight Noise", Status: "not_configured", Error: "API key not configured"})
	}

	// Infrastructure & Facilities
	if data.GreenSpaces != nil {
		results = append(results, APIResult{Name: "Green Spaces", Status: "success", Data: data.GreenSpaces})
	} else {
		results = append(results, APIResult{Name: "Green Spaces", Status: "error", Error: "Failed to fetch green spaces"})
	}

	if data.Education != nil {
		results = append(results, APIResult{Name: "Education Facilities", Status: "success", Data: data.Education})
	} else {
		results = append(results, APIResult{Name: "Education Facilities", Status: "error", Error: "Failed to fetch education data"})
	}

	if data.BuildingPermits != nil {
		results = append(results, APIResult{Name: "Building Permits", Status: "success", Data: data.BuildingPermits})
	} else {
		results = append(results, APIResult{Name: "Building Permits", Status: "not_configured", Error: "API varies by region"})
	}

	if data.Facilities != nil {
		results = append(results, APIResult{Name: "Facilities & Amenities", Status: "success", Data: data.Facilities})
	} else {
		results = append(results, APIResult{Name: "Facilities & Amenities", Status: "error", Error: "Failed to fetch facilities"})
	}

	if data.Elevation != nil {
		results = append(results, APIResult{Name: "AHN Height Model", Status: "success", Data: data.Elevation})
	} else {
		results = append(results, APIResult{Name: "AHN Height Model", Status: "error", Error: "Failed to fetch elevation data"})
	}

	// Heritage
	if data.MonumentStatus != nil {
		results = append(results, APIResult{Name: "Monument Status", Status: "success", Data: data.MonumentStatus})
	} else {
		results = append(results, APIResult{Name: "Monument Status", Status: "error", Error: "No monument data (or Amsterdam only)"})
	}

	// Comprehensive Platforms
	if data.PDOKData != nil {
		results = append(results, APIResult{Name: "PDOK Platform", Status: "success", Data: data.PDOKData})
	} else {
		results = append(results, APIResult{Name: "PDOK Platform", Status: "error", Error: "Failed to fetch PDOK data"})
	}

	if data.StratopoEnvironment != nil {
		results = append(results, APIResult{Name: "Stratopo Environment", Status: "success", Data: data.StratopoEnvironment})
	} else {
		results = append(results, APIResult{Name: "Stratopo Environment", Status: "not_configured", Error: "API key not configured"})
	}

	if data.LandUse != nil {
		results = append(results, APIResult{Name: "Land Use & Zoning", Status: "success", Data: data.LandUse})
	} else {
		results = append(results, APIResult{Name: "Land Use & Zoning", Status: "error", Error: "Failed to fetch land use data"})
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
