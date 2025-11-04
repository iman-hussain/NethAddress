package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

// WURSoilData represents soil physical properties
type WURSoilData struct {
	SoilType      string  `json:"soilType"`
	Composition   string  `json:"composition"`
	Permeability  float64 `json:"permeability"`
	OrganicMatter float64 `json:"organicMatter"`
	PH            float64 `json:"ph"`
	Suitability   string  `json:"suitability"`
}

// FetchWURSoilData retrieves soil properties for land quality assessment
// Documentation: https://www.soilphysics.wur.nl
func (c *ApiClient) FetchWURSoilData(cfg *config.Config, lat, lon float64) (*WURSoilData, error) {
	if cfg.WURSoilApiURL == "" {
		// Return default data when API is not configured (requires agreement)
		return &WURSoilData{
			SoilType:      "Unknown",
			Composition:   "Unknown",
			Permeability:  0,
			OrganicMatter: 0,
			PH:            0,
			Suitability:   "Unknown",
		}, nil
	}

	url := fmt.Sprintf("%s/soil?lat=%f&lon=%f", cfg.WURSoilApiURL, lat, lon)
	req, err := http.NewRequest("GET", url, nil)
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
		return &WURSoilData{
			SoilType:      "Unknown",
			Composition:   "Unknown",
			Permeability:  0,
			OrganicMatter: 0,
			PH:            0,
			Suitability:   "Unknown",
		}, nil
	}

	var result WURSoilData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode WUR soil response: %w", err)
	}

	return &result, nil
}

// SubsidenceData represents land subsidence data
type SubsidenceData struct {
	SubsidenceRate  float64 `json:"subsidenceRate"`  // mm/year
	TotalSubsidence float64 `json:"totalSubsidence"` // mm since baseline
	StabilityRating string  `json:"stabilityRating"` // Low, Medium, High risk
	MeasurementDate string  `json:"measurementDate"`
	GroundMovement  float64 `json:"groundMovement"`
}

// FetchSubsidenceData retrieves land subsidence and stability data
// Documentation: https://bodemdalingskaart.nl
func (c *ApiClient) FetchSubsidenceData(cfg *config.Config, lat, lon float64) (*SubsidenceData, error) {
	if cfg.SkyGeoSubsidenceApiURL == "" {
		// Return default data when API is not configured (paid service)
		return &SubsidenceData{
			SubsidenceRate:  0,
			TotalSubsidence: 0,
			StabilityRating: "Unknown",
			MeasurementDate: "",
			GroundMovement:  0,
		}, nil
	}

	url := fmt.Sprintf("%s/subsidence?lat=%f&lon=%f", cfg.SkyGeoSubsidenceApiURL, lat, lon)
	req, err := http.NewRequest("GET", url, nil)
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
		return &SubsidenceData{
			SubsidenceRate:  0,
			TotalSubsidence: 0,
			StabilityRating: "Unknown",
			MeasurementDate: "",
			GroundMovement:  0,
		}, nil
	}

	var result SubsidenceData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode subsidence response: %w", err)
	}

	return &result, nil
}

// SoilQualityData represents soil contamination and quality
type SoilQualityData struct {
	ContaminationLevel string   `json:"contaminationLevel"` // Clean, Light, Moderate, Severe
	Contaminants       []string `json:"contaminants"`
	QualityZone        string   `json:"qualityZone"`
	RestrictedUse      bool     `json:"restrictedUse"`
	LastTested         string   `json:"lastTested"`
}

// FetchSoilQualityData retrieves soil contamination data
// Documentation: https://api.store (government soil quality API)
func (c *ApiClient) FetchSoilQualityData(cfg *config.Config, lat, lon float64) (*SoilQualityData, error) {
	if cfg.SoilQualityApiURL == "" {
		return nil, fmt.Errorf("SoilQualityApiURL not configured")
	}

	url := fmt.Sprintf("%s/soil-quality?lat=%f&lon=%f", cfg.SoilQualityApiURL, lat, lon)
	req, err := http.NewRequest("GET", url, nil)
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
		return &SoilQualityData{
			ContaminationLevel: "Unknown",
			Contaminants:       []string{},
			RestrictedUse:      false,
		}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("soil quality API returned status %d", resp.StatusCode)
	}

	var result SoilQualityData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode soil quality response: %w", err)
	}

	return &result, nil
}

// BROSoilMapData represents soil types and foundation quality
type BROSoilMapData struct {
	SoilType          string  `json:"soilType"`
	PeatComposition   float64 `json:"peatComposition"` // percentage
	Profile           string  `json:"profile"`
	FoundationQuality string  `json:"foundationQuality"` // Excellent, Good, Fair, Poor
	GroundwaterDepth  float64 `json:"groundwaterDepth"`  // meters
}

// FetchBROSoilMapData retrieves BRO soil map data for foundation quality
// Documentation: https://www.dinoloket.nl/en/bro-soil-map
func (c *ApiClient) FetchBROSoilMapData(cfg *config.Config, lat, lon float64) (*BROSoilMapData, error) {
	if cfg.BROSoilMapApiURL == "" {
		// Return default data when API is not configured
		return &BROSoilMapData{
			SoilType:          "Unknown",
			PeatComposition:   0,
			Profile:           "Unknown",
			FoundationQuality: "Unknown",
			GroundwaterDepth:  0,
		}, nil
	}

	// Note: This would need WFS query implementation for actual PDOK BRO service
	// For now, return default data gracefully
	return &BROSoilMapData{
		SoilType:          "Unknown",
		PeatComposition:   0,
		Profile:           "Unknown",
		FoundationQuality: "Unknown",
		GroundwaterDepth:  0,
	}, nil
}
