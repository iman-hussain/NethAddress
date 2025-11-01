package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

// PDOKPlatformData represents comprehensive geodata from PDOK
type PDOKPlatformData struct {
	CadastralData  *CadastralInfo  `json:"cadastralData"`
	AddressData    *AddressInfo    `json:"addressData"`
	TopographyData *TopographyInfo `json:"topographyData"`
	BoundariesData *BoundariesInfo `json:"boundariesData"`
}

// CadastralInfo represents cadastral parcel information
type CadastralInfo struct {
	ParcelID     string  `json:"parcelId"`
	Municipality string  `json:"municipality"`
	Section      string  `json:"section"`
	ParcelNumber string  `json:"parcelNumber"`
	Area         float64 `json:"area"` // m²
	LandUse      string  `json:"landUse"`
}

// AddressInfo represents comprehensive address data
type AddressInfo struct {
	BAGID        string  `json:"bagId"`
	FullAddress  string  `json:"fullAddress"`
	Municipality string  `json:"municipality"`
	Province     string  `json:"province"`
	PostalCode   string  `json:"postalCode"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

// TopographyInfo represents topographic data
type TopographyInfo struct {
	LandType        string   `json:"landType"`
	TerrainFeatures []string `json:"terrainFeatures"`
	WaterBodies     []string `json:"waterBodies"`
}

// BoundariesInfo represents administrative boundaries
type BoundariesInfo struct {
	Municipality     string `json:"municipality"`
	Province         string `json:"province"`
	Neighborhood     string `json:"neighborhood"`
	NeighborhoodCode string `json:"neighborhoodCode"`
	District         string `json:"district"`
}

// FetchPDOKPlatformData retrieves comprehensive PDOK platform data
// Documentation: https://api.pdok.nl
func (c *ApiClient) FetchPDOKPlatformData(cfg *config.Config, lat, lon float64) (*PDOKPlatformData, error) {
	if cfg.PDOKApiURL == "" {
		return nil, fmt.Errorf("PDOKApiURL not configured")
	}

	url := fmt.Sprintf("%s/comprehensive?lat=%f&lon=%f", cfg.PDOKApiURL, lat, lon)
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
		return nil, fmt.Errorf("PDOK platform API returned status %d", resp.StatusCode)
	}

	var result PDOKPlatformData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode PDOK platform response: %w", err)
	}

	return &result, nil
}

// StratopoEnvironmentData represents comprehensive environmental assessment
type StratopoEnvironmentData struct {
	EnvironmentScore   float64                `json:"environmentScore"` // 0-100
	TotalVariables     int                    `json:"totalVariables"`
	PollutionIndex     float64                `json:"pollutionIndex"`
	UrbanizationLevel  string                 `json:"urbanizationLevel"` // Rural, Suburban, Urban, Metropolitan
	EnvironmentFactors map[string]interface{} `json:"environmentFactors"`
	ESGRating          string                 `json:"esgRating"` // A+ to E
	Recommendations    []string               `json:"recommendations"`
}

// FetchStratopoEnvironmentData retrieves 700+ environmental variables
// Documentation: https://stratopo.nl/en/environment-api/
func (c *ApiClient) FetchStratopoEnvironmentData(cfg *config.Config, lat, lon float64) (*StratopoEnvironmentData, error) {
	if cfg.StratopoApiURL == "" {
		return nil, fmt.Errorf("StratopoApiURL not configured")
	}

	url := fmt.Sprintf("%s/environment?lat=%f&lon=%f", cfg.StratopoApiURL, lat, lon)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if cfg.StratopoApiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.StratopoApiKey))
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("stratopo API returned status %d", resp.StatusCode)
	}

	var result StratopoEnvironmentData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode stratopo response: %w", err)
	}

	return &result, nil
}

// LandUseData represents land use and zoning information
type LandUseData struct {
	PrimaryUse      string            `json:"primaryUse"` // Residential, Commercial, Industrial, etc.
	ZoningCode      string            `json:"zoningCode"`
	ZoningDetails   string            `json:"zoningDetails"`
	Restrictions    []string          `json:"restrictions"`
	AllowedUses     []string          `json:"allowedUses"`
	BuildingRights  *BuildingRights   `json:"buildingRights"`
	ProtectedStatus string            `json:"protectedStatus"` // None, Monument, Conservation Area
	FuturePlans     []DevelopmentPlan `json:"futurePlans"`
}

// BuildingRights represents development rights
type BuildingRights struct {
	MaxHeight      float64 `json:"maxHeight"`      // meters
	MaxBuildArea   float64 `json:"maxBuildArea"`   // m²
	FloorAreaRatio float64 `json:"floorAreaRatio"` // FSI
	GroundCoverage float64 `json:"groundCoverage"` // percentage
	CanSubdivide   bool    `json:"canSubdivide"`
	CanExpand      bool    `json:"canExpand"`
}

// DevelopmentPlan represents future development
type DevelopmentPlan struct {
	PlanName     string `json:"planName"`
	Type         string `json:"type"`
	Status       string `json:"status"` // Proposed, Approved, In Progress
	ExpectedDate string `json:"expectedDate"`
	Impact       string `json:"impact"` // Positive, Neutral, Negative
}

// FetchLandUseData retrieves land use, zoning, and planning data
// Documentation: https://www.nationaalgeoregister.nl (CBS/PDOK)
func (c *ApiClient) FetchLandUseData(cfg *config.Config, lat, lon float64) (*LandUseData, error) {
	if cfg.LandUseApiURL == "" {
		return nil, fmt.Errorf("LandUseApiURL not configured")
	}

	url := fmt.Sprintf("%s/land-use?lat=%f&lon=%f", cfg.LandUseApiURL, lat, lon)
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
		return nil, fmt.Errorf("land use API returned status %d", resp.StatusCode)
	}

	var result LandUseData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode land use response: %w", err)
	}

	return &result, nil
}
