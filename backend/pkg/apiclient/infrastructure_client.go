package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

// GreenSpacesData represents parks and green areas
type GreenSpacesData struct {
	TotalGreenArea  float64      `json:"totalGreenArea"`  // m² within radius
	GreenPercentage float64      `json:"greenPercentage"` // percentage of area
	NearestPark     string       `json:"nearestPark"`
	ParkDistance    float64      `json:"parkDistance"`    // meters
	TreeCanopyCover float64      `json:"treeCanopyCover"` // percentage
	GreenSpaces     []GreenSpace `json:"greenSpaces"`
}

// GreenSpace represents a park or green area
type GreenSpace struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`       // Park, Forest, Garden, etc.
	Area       float64  `json:"area"`       // m²
	Distance   float64  `json:"distance"`   // meters
	Facilities []string `json:"facilities"` // Playground, Sports, etc.
}

// FetchGreenSpacesData retrieves parks and green areas for environmental quality
// Documentation: https://api.store (PDOK Green Spaces)
func (c *ApiClient) FetchGreenSpacesData(cfg *config.Config, lat, lon float64, radius int) (*GreenSpacesData, error) {
	if cfg.GreenSpacesApiURL == "" {
		return nil, fmt.Errorf("GreenSpacesApiURL not configured")
	}

	url := fmt.Sprintf("%s/green-spaces?lat=%f&lon=%f&radius=%d", cfg.GreenSpacesApiURL, lat, lon, radius)
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
		return nil, fmt.Errorf("green spaces API returned status %d", resp.StatusCode)
	}

	var result GreenSpacesData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode green spaces response: %w", err)
	}

	return &result, nil
}

// EducationData represents schools and education facilities
type EducationData struct {
	NearestPrimarySchool   *School  `json:"nearestPrimarySchool"`
	NearestSecondarySchool *School  `json:"nearestSecondarySchool"`
	AllSchools             []School `json:"allSchools"`
	AverageQuality         float64  `json:"averageQuality"` // 0-10 rating
}

// School represents an educational facility
type School struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`         // Primary, Secondary, Special
	Distance     float64 `json:"distance"`     // meters
	QualityScore float64 `json:"qualityScore"` // 0-10
	Students     int     `json:"students"`
	Address      string  `json:"address"`
	Denomination string  `json:"denomination"` // Public, Catholic, etc.
}

// FetchEducationData retrieves school locations and quality ratings
// Documentation: https://www.ocwincijfers.nl/open-data (CBS Education)
func (c *ApiClient) FetchEducationData(cfg *config.Config, lat, lon float64) (*EducationData, error) {
	if cfg.EducationApiURL == "" {
		return nil, fmt.Errorf("EducationApiURL not configured")
	}

	url := fmt.Sprintf("%s/schools?lat=%f&lon=%f&radius=2000", cfg.EducationApiURL, lat, lon)
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
		return nil, fmt.Errorf("education API returned status %d", resp.StatusCode)
	}

	var result EducationData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode education response: %w", err)
	}

	return &result, nil
}

// BuildingPermitsData represents recent construction activity
type BuildingPermitsData struct {
	TotalPermits    int              `json:"totalPermits"`
	NewConstruction int              `json:"newConstruction"`
	Renovations     int              `json:"renovations"`
	Permits         []BuildingPermit `json:"permits"`
	GrowthTrend     string           `json:"growthTrend"` // Increasing, Stable, Decreasing
}

// BuildingPermit represents a single permit
type BuildingPermit struct {
	PermitID     string  `json:"permitId"`
	Type         string  `json:"type"` // New, Renovation, Extension, Demolition
	Address      string  `json:"address"`
	Distance     float64 `json:"distance"` // meters
	IssueDate    string  `json:"issueDate"`
	ProjectValue float64 `json:"projectValue"` // EUR
	Status       string  `json:"status"`       // Approved, In Progress, Completed
}

// FetchBuildingPermitsData retrieves recent building activity
// Documentation: https://api.store (CBS Building Permits)
func (c *ApiClient) FetchBuildingPermitsData(cfg *config.Config, lat, lon float64, radius int) (*BuildingPermitsData, error) {
	if cfg.BuildingPermitsApiURL == "" {
		return nil, fmt.Errorf("BuildingPermitsApiURL not configured")
	}

	url := fmt.Sprintf("%s/permits?lat=%f&lon=%f&radius=%d&years=2", cfg.BuildingPermitsApiURL, lat, lon, radius)
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
		return nil, fmt.Errorf("building permits API returned status %d", resp.StatusCode)
	}

	var result BuildingPermitsData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode building permits response: %w", err)
	}

	return &result, nil
}

// FacilitiesData represents nearby amenities
type FacilitiesData struct {
	TopFacilities  []Facility     `json:"topFacilities"`
	AmenitiesScore float64        `json:"amenitiesScore"` // 0-100
	CategoryCounts map[string]int `json:"categoryCounts"`
}

// Facility represents a single amenity
type Facility struct {
	Name      string  `json:"name"`
	Category  string  `json:"category"`  // Retail, Healthcare, Leisure, etc.
	Type      string  `json:"type"`      // Supermarket, Hospital, Gym, etc.
	Distance  float64 `json:"distance"`  // meters
	WalkTime  int     `json:"walkTime"`  // minutes
	DriveTime int     `json:"driveTime"` // minutes
	Rating    float64 `json:"rating"`    // 0-5 stars
	Address   string  `json:"address"`
}

// FetchFacilitiesData retrieves nearby retail, healthcare, and amenities
// Documentation: Municipal API (varies by city)
func (c *ApiClient) FetchFacilitiesData(cfg *config.Config, lat, lon float64) (*FacilitiesData, error) {
	if cfg.FacilitiesApiURL == "" {
		return nil, fmt.Errorf("FacilitiesApiURL not configured")
	}

	url := fmt.Sprintf("%s/facilities?lat=%f&lon=%f&radius=1500", cfg.FacilitiesApiURL, lat, lon)
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
		return nil, fmt.Errorf("facilities API returned status %d", resp.StatusCode)
	}

	var result FacilitiesData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode facilities response: %w", err)
	}

	return &result, nil
}

// AHNHeightData represents elevation and terrain data
type AHNHeightData struct {
	Elevation     float64   `json:"elevation"`     // meters above NAP
	TerrainSlope  float64   `json:"terrainSlope"`  // degrees
	FloodRisk     string    `json:"floodRisk"`     // Low, Medium, High based on elevation
	ViewPotential string    `json:"viewPotential"` // Poor, Fair, Good, Excellent
	Surrounding   []float64 `json:"surrounding"`   // Elevations of nearby points
}

// FetchAHNHeightData retrieves elevation and terrain models
// Documentation: https://www.ahn.nl (PDOK/Kadaster)
func (c *ApiClient) FetchAHNHeightData(cfg *config.Config, lat, lon float64) (*AHNHeightData, error) {
	if cfg.AHNHeightModelApiURL == "" {
		return nil, fmt.Errorf("AHNHeightModelApiURL not configured")
	}

	url := fmt.Sprintf("%s/height?lat=%f&lon=%f", cfg.AHNHeightModelApiURL, lat, lon)
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
		return nil, fmt.Errorf("AHN height API returned status %d", resp.StatusCode)
	}

	var result AHNHeightData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode AHN height response: %w", err)
	}

	// Assess flood risk based on elevation
	if result.Elevation < -2.0 {
		result.FloodRisk = "High"
	} else if result.Elevation < 1.0 {
		result.FloodRisk = "Medium"
	} else {
		result.FloodRisk = "Low"
	}

	// Assess view potential based on relative elevation
	if result.Elevation > 5.0 {
		result.ViewPotential = "Excellent"
	} else if result.Elevation > 2.0 {
		result.ViewPotential = "Good"
	} else {
		result.ViewPotential = "Fair"
	}

	return &result, nil
}
