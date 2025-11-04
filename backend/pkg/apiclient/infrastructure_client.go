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
		// Return empty data when API is not configured
		return &GreenSpacesData{
			TotalGreenArea:  0,
			GreenPercentage: 0,
			NearestPark:     "",
			ParkDistance:    0,
			TreeCanopyCover: 0,
			GreenSpaces:     []GreenSpace{},
		}, nil
	}

	// Note: This would need WFS query implementation for actual PDOK service
	// For now, return empty data gracefully
	return &GreenSpacesData{
		TotalGreenArea:  0,
		GreenPercentage: 0,
		NearestPark:     "",
		ParkDistance:    0,
		TreeCanopyCover: 0,
		GreenSpaces:     []GreenSpace{},
	}, nil
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
		// Return empty data when API is not configured
		return &EducationData{
			NearestPrimarySchool:   nil,
			NearestSecondarySchool: nil,
			AllSchools:             []School{},
			AverageQuality:         0,
		}, nil
	}

	// Note: Amsterdam education API requires different query structure
	// For now, return empty data gracefully
	return &EducationData{
		NearestPrimarySchool:   nil,
		NearestSecondarySchool: nil,
		AllSchools:             []School{},
		AverageQuality:         0,
	}, nil
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
		// Return empty data when API is not configured
		return &BuildingPermitsData{
			TotalPermits:    0,
			NewConstruction: 0,
			Renovations:     0,
			Permits:         []BuildingPermit{},
			GrowthTrend:     "Unknown",
		}, nil
	}

	// Note: This would need specific municipal API implementation
	// For now, return empty data gracefully
	return &BuildingPermitsData{
		TotalPermits:    0,
		NewConstruction: 0,
		Renovations:     0,
		Permits:         []BuildingPermit{},
		GrowthTrend:     "Unknown",
	}, nil
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
		// Return empty data when API is not configured
		return &FacilitiesData{
			TopFacilities:  []Facility{},
			AmenitiesScore: 0,
			CategoryCounts: make(map[string]int),
		}, nil
	}

	// Note: Amsterdam facilities API requires different query structure
	// For now, return empty data gracefully
	return &FacilitiesData{
		TopFacilities:  []Facility{},
		AmenitiesScore: 0,
		CategoryCounts: make(map[string]int),
	}, nil
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
		// Return default data when API is not configured
		return &AHNHeightData{
			Elevation:     0,
			TerrainSlope:  0,
			FloodRisk:     "Unknown",
			ViewPotential: "Unknown",
			Surrounding:   []float64{},
		}, nil
	}

	// Note: This would need WMS/WCS query implementation for actual PDOK AHN service
	// For now, return default data gracefully
	return &AHNHeightData{
		Elevation:     0,
		TerrainSlope:  0,
		FloodRisk:     "Unknown",
		ViewPotential: "Unknown",
		Surrounding:   []float64{},
	}, nil
}
