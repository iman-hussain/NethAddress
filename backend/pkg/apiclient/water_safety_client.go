package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// Default PDOK flood risk API endpoint (free, no auth required)
const defaultFloodRiskApiURL = "https://api.pdok.nl/rws/overstromingen-risicogebied/ogc/v1"

// FetchFloodRiskData retrieves flood risk assessment using PDOK Rijkswaterstaat API
// Documentation: https://api.pdok.nl/rws/overstromingen-risicogebied/ogc/v1
func (c *ApiClient) FetchFloodRiskData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.FloodRiskData, error) {
	// Always use PDOK API default (free, no auth) - ignore config overrides which may have bad URLs
	baseURL := defaultFloodRiskApiURL

	// Create bounding box around the point
	delta := 0.005 // ~500m
	bbox := fmt.Sprintf("%.6f,%.6f,%.6f,%.6f", lon-delta, lat-delta, lon+delta, lat+delta)

	url := fmt.Sprintf("%s/collections/risk_zone/items?bbox=%s&f=json&limit=5", baseURL, bbox)
	logutil.Debugf("[FloodRisk] Request URL: %s", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logutil.Debugf("[FloodRisk] Request error: %v", err)
		return defaultFloodRiskData("Unknown"), nil
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[FloodRisk] HTTP error: %v", err)
		return defaultFloodRiskData("Unknown"), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logutil.Debugf("[FloodRisk] Non-200 status: %d", resp.StatusCode)
		// If API returns error, assume low risk (most of Netherlands is protected)
		return defaultFloodRiskData("Low"), nil
	}

	var apiResp models.FloodRiskResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[FloodRisk] Decode error: %v", err)
		return defaultFloodRiskData("Low"), nil
	}

	logutil.Debugf("[FloodRisk] Found %d risk areas", len(apiResp.Features))

	if len(apiResp.Features) == 0 {
		// No flood risk areas found - location is likely safe
		return &models.FloodRiskData{
			RiskLevel:        "Low",
			FloodProbability: 0.01,
			WaterDepth:       0,
			FloodZone:        "Protected",
			DikeQuality:      "Good",
		}, nil
	}

	// Analyse the risk areas found
	feature := apiResp.Features[0]
	riskLevel := "Medium"
	probability := 0.1

	// Parse risk level from qualitative_value field (INSPIRE format)
	qualValue := strings.ToLower(feature.Properties.QualitativeValue)
	description := strings.ToLower(feature.Properties.Description)

	switch {
	case strings.Contains(qualValue, "potential significant"):
		// "Area of Potential Significant Flood Risk"
		riskLevel = "Medium"
		probability = 0.1
	case strings.Contains(qualValue, "high") || strings.Contains(qualValue, "significant"):
		riskLevel = "High"
		probability = 1.0
	case strings.Contains(qualValue, "low") || strings.Contains(qualValue, "minor"):
		riskLevel = "Low"
		probability = 0.01
	}

	// Refine based on description if available
	if strings.Contains(description, "beschermd") {
		// "beschermd" = protected, so reduce risk
		if riskLevel == "High" {
			riskLevel = "Medium"
			probability = 0.1
		}
	}

	floodZone := feature.Properties.Description
	if floodZone == "" {
		floodZone = feature.Properties.QualitativeValue
	}

	result := &models.FloodRiskData{
		RiskLevel:        riskLevel,
		FloodProbability: probability,
		WaterDepth:       0, // Not provided in this API
		FloodZone:        floodZone,
		DikeQuality:      "Good", // Netherlands has good dikes generally
	}

	logutil.Debugf("[FloodRisk] Result: level=%s, zone=%s", riskLevel, floodZone)
	return result, nil
}

func defaultFloodRiskData(level string) *models.FloodRiskData {
	return &models.FloodRiskData{
		RiskLevel:        level,
		FloodProbability: 0,
		WaterDepth:       0,
		NearestDike:      0,
		DikeQuality:      "Unknown",
		FloodZone:        "",
	}
}

// FetchWaterQualityData retrieves water quality and management data
// Documentation: https://www.dutchwatersector.com (Digital Delta)
func (c *ApiClient) FetchWaterQualityData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.WaterQualityData, error) {
	if cfg.DigitalDeltaApiURL == "" {
		return nil, fmt.Errorf("DigitalDeltaApiURL not configured")
	}

	url := fmt.Sprintf("%s/water-quality?lat=%f&lon=%f", cfg.DigitalDeltaApiURL, lat, lon)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// No nearby water body
		return &models.WaterQualityData{
			WaterQuality: "N/A",
			Distance:     9999,
			Parameters:   make(map[string]float64),
		}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("water quality API returned status %d", resp.StatusCode)
	}

	var result models.WaterQualityData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode water quality response: %w", err)
	}

	return &result, nil
}

// FetchSafetyData retrieves safety perception and crime statistics
// Documentation: https://api.store (CBS Safety Experience)
func (c *ApiClient) FetchSafetyData(ctx context.Context, cfg *config.Config, neighborhoodCode string) (*models.SafetyData, error) {
	if cfg.SafetyExperienceApiURL == "" {
		return nil, fmt.Errorf("SafetyExperienceApiURL not configured")
	}

	url := fmt.Sprintf("%s/safety?neighborhood=%s", cfg.SafetyExperienceApiURL, neighborhoodCode)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// No safety data available
		return &models.SafetyData{
			SafetyScore:      70.0, // Neutral default
			SafetyPerception: "Moderate",
			CrimeTypes:       make(map[string]int),
		}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("safety API returned status %d", resp.StatusCode)
	}

	var result models.SafetyData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode safety response: %w", err)
	}

	// Categorize safety perception
	if result.SafetyScore >= 80 {
		result.SafetyPerception = "Very Safe"
	} else if result.SafetyScore >= 60 {
		result.SafetyPerception = "Safe"
	} else if result.SafetyScore >= 40 {
		result.SafetyPerception = "Moderate"
	} else {
		result.SafetyPerception = "Unsafe"
	}

	return &result, nil
}

// FetchSchipholFlightData retrieves flight path and noise data
// Documentation: https://developer.schiphol.nl
func (c *ApiClient) FetchSchipholFlightData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.SchipholFlightData, error) {
	if cfg.SchipholApiURL == "" {
		return nil, fmt.Errorf("SchipholApiURL not configured")
	}

	url := fmt.Sprintf("%s/noise-impact?lat=%f&lon=%f", cfg.SchipholApiURL, lat, lon)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Schiphol requires both API key and App ID
	if cfg.SchipholApiKey != "" {
		req.Header.Set("ResourceVersion", "v4")
		req.Header.Set("app_id", cfg.SchipholAppID)
		req.Header.Set("app_key", cfg.SchipholApiKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// Location not affected by Schiphol noise
		return &models.SchipholFlightData{
			DailyFlights: 0,
			NoiseLevel:   0,
			FlightPaths:  []models.FlightPath{},
			NightFlights: 0,
			NoiseContour: "None",
		}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("schiphol API returned status %d", resp.StatusCode)
	}

	var result models.SchipholFlightData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode schiphol response: %w", err)
	}

	return &result, nil
}
