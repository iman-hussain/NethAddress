package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// FetchWURSoilData retrieves soil properties for land quality assessment
// Documentation: https://www.soilphysics.wur.nl
func (c *ApiClient) FetchWURSoilData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.WURSoilData, error) {
	if cfg.WURSoilApiURL == "" {
		// Return default data when API is not configured (requires agreement)
		return &models.WURSoilData{
			SoilType:      "Unknown",
			Composition:   "Unknown",
			Permeability:  0,
			OrganicMatter: 0,
			PH:            0,
			Suitability:   "Unknown",
		}, nil
	}

	url := fmt.Sprintf("%s/soil?lat=%f&lon=%f", cfg.WURSoilApiURL, lat, lon)
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

	if resp.StatusCode != 200 {
		return &models.WURSoilData{
			SoilType:      "Unknown",
			Composition:   "Unknown",
			Permeability:  0,
			OrganicMatter: 0,
			PH:            0,
			Suitability:   "Unknown",
		}, nil
	}

	var result models.WURSoilData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode WUR soil response: %w", err)
	}

	return &result, nil
}

// FetchSubsidenceData retrieves land subsidence and stability data
// Documentation: https://bodemdalingskaart.nl
func (c *ApiClient) FetchSubsidenceData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.SubsidenceData, error) {
	if cfg.SkyGeoSubsidenceApiURL == "" {
		// Return default data when API is not configured (paid service)
		return &models.SubsidenceData{
			SubsidenceRate:  0,
			TotalSubsidence: 0,
			StabilityRating: "Unknown",
			MeasurementDate: "",
			GroundMovement:  0,
		}, nil
	}

	url := fmt.Sprintf("%s/subsidence?lat=%f&lon=%f", cfg.SkyGeoSubsidenceApiURL, lat, lon)
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

	if resp.StatusCode != 200 {
		return &models.SubsidenceData{
			SubsidenceRate:  0,
			TotalSubsidence: 0,
			StabilityRating: "Unknown",
			MeasurementDate: "",
			GroundMovement:  0,
		}, nil
	}

	var result models.SubsidenceData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode subsidence response: %w", err)
	}

	return &result, nil
}

// FetchSoilQualityData retrieves soil contamination data
// Documentation: https://api.store (government soil quality API)
func (c *ApiClient) FetchSoilQualityData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.SoilQualityData, error) {
	if cfg.SoilQualityApiURL == "" {
		return nil, fmt.Errorf("SoilQualityApiURL not configured")
	}

	url := fmt.Sprintf("%s/soil-quality?lat=%f&lon=%f", cfg.SoilQualityApiURL, lat, lon)
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
		// No contamination data available - assume clean
		return &models.SoilQualityData{
			ContaminationLevel: "Unknown",
			Contaminants:       []string{},
			RestrictedUse:      false,
		}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("soil quality API returned status %d", resp.StatusCode)
	}

	var result models.SoilQualityData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode soil quality response: %w", err)
	}

	return &result, nil
}

// FetchBROSoilMapData retrieves BRO soil map data for foundation quality
// Documentation: https://www.dinoloket.nl/en/bro-soil-map
func (c *ApiClient) FetchBROSoilMapData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.BROSoilMapData, error) {
	// Return default data if not configured
	if cfg.BROSoilMapApiURL == "" {
		return &models.BROSoilMapData{
			SoilType:          "Unknown",
			PeatComposition:   0,
			Profile:           "Unknown",
			FoundationQuality: "Unknown",
			GroundwaterDepth:  0,
		}, nil
	}

	url := fmt.Sprintf("%s/bro/soil-map?lat=%f&lon=%f", cfg.BROSoilMapApiURL, lat, lon)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &models.BROSoilMapData{
			SoilType:          "Unknown",
			PeatComposition:   0,
			Profile:           "Unknown",
			FoundationQuality: "Unknown",
			GroundwaterDepth:  0,
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return &models.BROSoilMapData{
			SoilType:          "Unknown",
			PeatComposition:   0,
			Profile:           "Unknown",
			FoundationQuality: "Unknown",
			GroundwaterDepth:  0,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &models.BROSoilMapData{
			SoilType:          "Unknown",
			PeatComposition:   0,
			Profile:           "Unknown",
			FoundationQuality: "Unknown",
			GroundwaterDepth:  0,
		}, nil
	}

	var result models.BROSoilMapData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &models.BROSoilMapData{
			SoilType:          "Unknown",
			PeatComposition:   0,
			Profile:           "Unknown",
			FoundationQuality: "Unknown",
			GroundwaterDepth:  0,
		}, nil
	}

	return &result, nil
}
