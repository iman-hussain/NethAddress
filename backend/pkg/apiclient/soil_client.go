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
		return nil, fmt.Errorf("WURSoilApiURL not configured")
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
		return nil, fmt.Errorf("WUR soil API returned status %d", resp.StatusCode)
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
		return nil, fmt.Errorf("SkyGeoSubsidenceApiURL not configured")
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
		return nil, fmt.Errorf("subsidence API returned status %d", resp.StatusCode)
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
		return nil, fmt.Errorf("BROSoilMapApiURL not configured")
	}

	url := fmt.Sprintf("%s/bro/soil-map?lat=%f&lon=%f", cfg.BROSoilMapApiURL, lat, lon)
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
		return nil, fmt.Errorf("BRO soil map API returned status %d", resp.StatusCode)
	}

	var result BROSoilMapData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode BRO soil map response: %w", err)
	}

	return &result, nil
}
