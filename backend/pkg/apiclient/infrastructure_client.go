package apiclient

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
)

// Default PDOK API endpoints for infrastructure data (free, no auth required)
const (
	defaultBGTApiURL     = "https://api.pdok.nl/lv/bgt/ogc/v1"
	defaultNatura2000URL = "https://api.pdok.nl/rvo/natura2000/ogc/v1"
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

// bgtGreenResponse represents PDOK BGT begroeidterreindeel API response
type bgtGreenResponse struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Properties struct {
			FysiekVoorkomen string `json:"fysiekVoorkomen"` // e.g., "groenvoorziening", "bos"
			Naam            string `json:"naam"`
			OpenbareRuimte  string `json:"openbareRuimte"`
		} `json:"properties"`
		Geometry struct {
			Type        string          `json:"type"`
			Coordinates json.RawMessage `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
	NumberReturned int `json:"numberReturned"`
}

// natura2000Response represents PDOK Natura2000 API response
type natura2000Response struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Properties struct {
			Naam        string  `json:"naam"`
			Oppervlakte float64 `json:"oppervlakte"`
			Status      string  `json:"status"`
		} `json:"properties"`
	} `json:"features"`
	NumberReturned int `json:"numberReturned"`
}

// FetchGreenSpacesData retrieves parks and green areas using PDOK BGT API
// Documentation: https://api.pdok.nl/lv/bgt/ogc/v1
func (c *ApiClient) FetchGreenSpacesData(cfg *config.Config, lat, lon float64, radius int) (*GreenSpacesData, error) {
	// Always use PDOK BGT API default (free, no auth) - ignore config overrides which may have bad URLs
	baseURL := defaultBGTApiURL

	// Create bounding box based on radius (convert meters to degrees approximately)
	delta := float64(radius) / 111000.0 // ~111km per degree
	bbox := fmt.Sprintf("%.6f,%.6f,%.6f,%.6f", lon-delta, lat-delta, lon+delta, lat+delta)

	// Query BGT begroeidterreindeel (vegetated terrain)
	url := fmt.Sprintf("%s/collections/begroeidterreindeel/items?bbox=%s&f=json&limit=50", baseURL, bbox)
	logutil.Debugf("[GreenSpaces] Request URL: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logutil.Debugf("[GreenSpaces] Request error: %v", err)
		return emptyGreenSpacesData(), nil
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[GreenSpaces] HTTP error: %v", err)
		return emptyGreenSpacesData(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logutil.Debugf("[GreenSpaces] Non-200 status: %d", resp.StatusCode)
		return emptyGreenSpacesData(), nil
	}

	var apiResp bgtGreenResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[GreenSpaces] Decode error: %v", err)
		return emptyGreenSpacesData(), nil
	}

	logutil.Debugf("[GreenSpaces] Found %d green areas", len(apiResp.Features))

	// Calculate total green area and categorise
	greenSpaces := make([]GreenSpace, 0)
	totalArea := 0.0
	nearestPark := ""
	nearestDistance := math.MaxFloat64

	for _, feature := range apiResp.Features {
		greenType := mapBGTTypeToGreenType(feature.Properties.FysiekVoorkomen)
		name := feature.Properties.Naam
		if name == "" {
			name = feature.Properties.OpenbareRuimte
		}
		if name == "" {
			name = greenType
		}

		// Estimate area (simplified - would need proper geometry calculation)
		area := 500.0 // Default estimate

		// Estimate distance from center (simplified)
		distance := float64(radius) / 2 // Placeholder

		if distance < nearestDistance && (greenType == "Park" || greenType == "Garden") {
			nearestDistance = distance
			nearestPark = name
		}

		totalArea += area

		greenSpaces = append(greenSpaces, GreenSpace{
			Name:     name,
			Type:     greenType,
			Area:     area,
			Distance: distance,
		})
	}

	// Calculate green percentage of search area
	searchArea := math.Pi * float64(radius) * float64(radius)
	greenPercentage := 0.0
	if searchArea > 0 {
		greenPercentage = (totalArea / searchArea) * 100
		if greenPercentage > 100 {
			greenPercentage = 100
		}
	}

	// Also check for Natura2000 protected areas nearby
	natura2000Parks := c.fetchNatura2000Areas(lat, lon, radius)
	greenSpaces = append(greenSpaces, natura2000Parks...)

	if nearestPark == "" && len(natura2000Parks) > 0 {
		nearestPark = natura2000Parks[0].Name
		nearestDistance = natura2000Parks[0].Distance
	}

	result := &GreenSpacesData{
		TotalGreenArea:  totalArea,
		GreenPercentage: greenPercentage,
		NearestPark:     nearestPark,
		ParkDistance:    nearestDistance,
		TreeCanopyCover: greenPercentage * 0.3, // Estimate tree cover as 30% of green
		GreenSpaces:     greenSpaces,
	}

	logutil.Debugf("[GreenSpaces] Result: area=%.0f, percentage=%.1f%%, nearest=%s", totalArea, greenPercentage, nearestPark)
	return result, nil
}

func (c *ApiClient) fetchNatura2000Areas(lat, lon float64, radius int) []GreenSpace {
	delta := float64(radius) / 111000.0 * 5 // Larger search area for nature reserves
	bbox := fmt.Sprintf("%.6f,%.6f,%.6f,%.6f", lon-delta, lat-delta, lon+delta, lat+delta)

	url := fmt.Sprintf("%s/collections/natura2000/items?bbox=%s&f=json&limit=5", defaultNatura2000URL, bbox)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	var apiResp natura2000Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil
	}

	parks := make([]GreenSpace, 0, len(apiResp.Features))
	for _, f := range apiResp.Features {
		parks = append(parks, GreenSpace{
			Name:     f.Properties.Naam,
			Type:     "Nature Reserve",
			Area:     f.Properties.Oppervlakte,
			Distance: float64(radius) * 3, // These are typically further away
		})
	}

	return parks
}

func mapBGTTypeToGreenType(bgtType string) string {
	switch bgtType {
	case "groenvoorziening":
		return "Garden"
	case "bos", "loofbos", "naaldbos", "gemengd bos":
		return "Forest"
	case "grasland agrarisch", "grasland overig":
		return "Grassland"
	case "heide":
		return "Heath"
	case "moeras":
		return "Wetland"
	case "fruitteelt":
		return "Orchard"
	default:
		return "Green Area"
	}
}

func emptyGreenSpacesData() *GreenSpacesData {
	return &GreenSpacesData{
		TotalGreenArea:  0,
		GreenPercentage: 0,
		NearestPark:     "",
		ParkDistance:    0,
		TreeCanopyCover: 0,
		GreenSpaces:     []GreenSpace{},
	}
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

// overpassResponse represents OSM Overpass API response
type overpassResponse struct {
	Elements []struct {
		Type string  `json:"type"`
		ID   int64   `json:"id"`
		Lat  float64 `json:"lat"`
		Lon  float64 `json:"lon"`
		Tags struct {
			Name        string `json:"name"`
			Amenity     string `json:"amenity"`
			ISCEDLevel  string `json:"isced:level"`
			Operator    string `json:"operator"`
			Religion    string `json:"religion"`
			AddrStreet  string `json:"addr:street"`
			AddrHouseNo string `json:"addr:housenumber"`
			AddrCity    string `json:"addr:city"`
		} `json:"tags"`
	} `json:"elements"`
}

// FetchEducationData retrieves school locations using OSM Overpass API
// Documentation: https://wiki.openstreetmap.org/wiki/Overpass_API
func (c *ApiClient) FetchEducationData(cfg *config.Config, lat, lon float64) (*EducationData, error) {
	// Use Overpass API to find schools (free, no auth required)
	overpassURL := "https://overpass-api.de/api/interpreter"

	// Query for schools within ~2km radius
	radius := 2000
	query := fmt.Sprintf(`[out:json][timeout:10];
(
  node["amenity"="school"](around:%d,%.6f,%.6f);
  way["amenity"="school"](around:%d,%.6f,%.6f);
);
out center body qt 20;`, radius, lat, lon, radius, lat, lon)

	logutil.Debugf("[Education] Querying Overpass API for schools near %.6f, %.6f", lat, lon)

	// Send query as POST body (not query string)
	reqBody := strings.NewReader("data=" + query)
	req, err := http.NewRequest("POST", overpassURL, reqBody)
	if err != nil {
		logutil.Debugf("[Education] Request error: %v", err)
		return emptyEducationData(), nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[Education] HTTP error: %v", err)
		return emptyEducationData(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logutil.Debugf("[Education] Non-200 status: %d", resp.StatusCode)
		return emptyEducationData(), nil
	}

	var apiResp overpassResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[Education] Decode error: %v", err)
		return emptyEducationData(), nil
	}

	logutil.Debugf("[Education] Found %d schools", len(apiResp.Elements))

	allSchools := make([]School, 0, len(apiResp.Elements))
	var nearestPrimary *School
	var nearestSecondary *School
	minPrimaryDist := math.MaxFloat64
	minSecondaryDist := math.MaxFloat64

	for _, elem := range apiResp.Elements {
		// Calculate distance using Haversine formula (simplified)
		distance := haversineDistance(lat, lon, elem.Lat, elem.Lon)

		schoolType := determineSchoolType(elem.Tags.ISCEDLevel, elem.Tags.Name)
		denomination := elem.Tags.Religion
		if denomination == "" {
			denomination = "Public"
		}

		address := ""
		if elem.Tags.AddrStreet != "" {
			address = fmt.Sprintf("%s %s, %s", elem.Tags.AddrStreet, elem.Tags.AddrHouseNo, elem.Tags.AddrCity)
		}

		school := School{
			Name:         elem.Tags.Name,
			Type:         schoolType,
			Distance:     distance,
			QualityScore: 7.0, // Default quality score (would need separate data source)
			Address:      address,
			Denomination: denomination,
		}

		allSchools = append(allSchools, school)

		// Track nearest primary and secondary schools
		if schoolType == "Primary" && distance < minPrimaryDist {
			minPrimaryDist = distance
			schoolCopy := school
			nearestPrimary = &schoolCopy
		} else if schoolType == "Secondary" && distance < minSecondaryDist {
			minSecondaryDist = distance
			schoolCopy := school
			nearestSecondary = &schoolCopy
		}
	}

	// Calculate average quality
	avgQuality := 7.0 // Default
	if len(allSchools) > 0 {
		total := 0.0
		for _, s := range allSchools {
			total += s.QualityScore
		}
		avgQuality = total / float64(len(allSchools))
	}

	result := &EducationData{
		NearestPrimarySchool:   nearestPrimary,
		NearestSecondarySchool: nearestSecondary,
		AllSchools:             allSchools,
		AverageQuality:         avgQuality,
	}

	logutil.Debugf("[Education] Result: %d schools found, nearest primary: %v", len(allSchools), nearestPrimary != nil)
	return result, nil
}

func determineSchoolType(iscedLevel, name string) string {
	// ISCED levels: 0=pre-primary, 1=primary, 2=lower secondary, 3=upper secondary
	switch iscedLevel {
	case "0":
		return "Pre-Primary"
	case "1":
		return "Primary"
	case "2", "3":
		return "Secondary"
	}

	// Fallback: infer from name
	nameLower := strings.ToLower(name)
	if strings.Contains(nameLower, "basisschool") || strings.Contains(nameLower, "primary") {
		return "Primary"
	}
	if strings.Contains(nameLower, "vmbo") || strings.Contains(nameLower, "havo") ||
		strings.Contains(nameLower, "vwo") || strings.Contains(nameLower, "college") ||
		strings.Contains(nameLower, "lyceum") || strings.Contains(nameLower, "secondary") {
		return "Secondary"
	}

	return "Primary" // Default assumption
}

// haversineDistance calculates the distance in meters between two lat/lon points
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func emptyEducationData() *EducationData {
	return &EducationData{
		NearestPrimarySchool:   nil,
		NearestSecondarySchool: nil,
		AllSchools:             []School{},
		AverageQuality:         0,
	}
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
	// Return empty data if not configured
	if cfg.BuildingPermitsApiURL == "" {
		return &BuildingPermitsData{
			TotalPermits:    0,
			NewConstruction: 0,
			Renovations:     0,
			Permits:         []BuildingPermit{},
			GrowthTrend:     "Unknown",
		}, nil
	}

	url := fmt.Sprintf("%s/permits?lat=%f&lon=%f&radius=%d&years=2", cfg.BuildingPermitsApiURL, lat, lon, radius)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &BuildingPermitsData{
			TotalPermits:    0,
			NewConstruction: 0,
			Renovations:     0,
			Permits:         []BuildingPermit{},
			GrowthTrend:     "Unknown",
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return &BuildingPermitsData{
			TotalPermits:    0,
			NewConstruction: 0,
			Renovations:     0,
			Permits:         []BuildingPermit{},
			GrowthTrend:     "Unknown",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &BuildingPermitsData{
			TotalPermits:    0,
			NewConstruction: 0,
			Renovations:     0,
			Permits:         []BuildingPermit{},
			GrowthTrend:     "Unknown",
		}, nil
	}

	var result BuildingPermitsData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &BuildingPermitsData{
			TotalPermits:    0,
			NewConstruction: 0,
			Renovations:     0,
			Permits:         []BuildingPermit{},
			GrowthTrend:     "Unknown",
		}, nil
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

// overpassFacilitiesResponse for OSM amenities query
type overpassFacilitiesResponse struct {
	Elements []struct {
		Type string  `json:"type"`
		ID   int64   `json:"id"`
		Lat  float64 `json:"lat"`
		Lon  float64 `json:"lon"`
		Tags struct {
			Name       string `json:"name"`
			Amenity    string `json:"amenity"`
			Shop       string `json:"shop"`
			Leisure    string `json:"leisure"`
			Healthcare string `json:"healthcare"`
		} `json:"tags"`
	} `json:"elements"`
}

// FetchFacilitiesData retrieves nearby amenities using OSM Overpass API
// Documentation: https://wiki.openstreetmap.org/wiki/Overpass_API
func (c *ApiClient) FetchFacilitiesData(cfg *config.Config, lat, lon float64) (*FacilitiesData, error) {
	overpassURL := "https://overpass-api.de/api/interpreter"

	// Query for common amenities within 1.5km
	radius := 1500
	query := fmt.Sprintf(`[out:json][timeout:15];
(
  node["shop"="supermarket"](around:%d,%.6f,%.6f);
  node["amenity"="pharmacy"](around:%d,%.6f,%.6f);
  node["amenity"="doctors"](around:%d,%.6f,%.6f);
  node["amenity"="hospital"](around:%d,%.6f,%.6f);
  node["amenity"="restaurant"](around:%d,%.6f,%.6f);
  node["amenity"="cafe"](around:%d,%.6f,%.6f);
  node["leisure"="fitness_centre"](around:%d,%.6f,%.6f);
  node["amenity"="bank"](around:%d,%.6f,%.6f);
  node["amenity"="post_office"](around:%d,%.6f,%.6f);
);
out body qt 50;`, radius, lat, lon, radius, lat, lon, radius, lat, lon, radius, lat, lon,
		radius, lat, lon, radius, lat, lon, radius, lat, lon, radius, lat, lon, radius, lat, lon)

	logutil.Debugf("[Facilities] Querying Overpass API for amenities near %.6f, %.6f", lat, lon)

	// Send query as POST body (not query string)
	reqBody := strings.NewReader("data=" + query)
	req, err := http.NewRequest("POST", overpassURL, reqBody)
	if err != nil {
		logutil.Debugf("[Facilities] Request error: %v", err)
		return emptyFacilitiesData(), nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[Facilities] HTTP error: %v", err)
		return emptyFacilitiesData(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logutil.Debugf("[Facilities] Non-200 status: %d", resp.StatusCode)
		return emptyFacilitiesData(), nil
	}

	var apiResp overpassFacilitiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[Facilities] Decode error: %v", err)
		return emptyFacilitiesData(), nil
	}

	logutil.Debugf("[Facilities] Found %d amenities", len(apiResp.Elements))

	facilities := make([]Facility, 0, len(apiResp.Elements))
	categoryCounts := make(map[string]int)

	for _, elem := range apiResp.Elements {
		distance := haversineDistance(lat, lon, elem.Lat, elem.Lon)

		category, facilityType := categorizeFacility(elem.Tags.Amenity, elem.Tags.Shop, elem.Tags.Leisure, elem.Tags.Healthcare)
		name := elem.Tags.Name
		if name == "" {
			name = facilityType
		}

		// Estimate walk time (80m/min average walking speed)
		walkTime := int(distance / 80)
		driveTime := int(distance / 500) // ~30km/h in city

		facility := Facility{
			Name:      name,
			Category:  category,
			Type:      facilityType,
			Distance:  distance,
			WalkTime:  walkTime,
			DriveTime: driveTime,
			Rating:    4.0, // Default - would need external API
		}

		facilities = append(facilities, facility)
		categoryCounts[category]++
	}

	// Sort by distance and limit
	sortFacilitiesByDistance(facilities)
	if len(facilities) > 20 {
		facilities = facilities[:20]
	}

	// Calculate amenities score (0-100)
	amenitiesScore := calculateAmenitiesScore(categoryCounts, facilities)

	result := &FacilitiesData{
		TopFacilities:  facilities,
		AmenitiesScore: amenitiesScore,
		CategoryCounts: categoryCounts,
	}

	logutil.Debugf("[Facilities] Result: %d facilities, score=%.1f", len(facilities), amenitiesScore)
	return result, nil
}

func categorizeFacility(amenity, shop, leisure, healthcare string) (category, facilityType string) {
	if shop == "supermarket" {
		return "Retail", "Supermarket"
	}
	if amenity == "pharmacy" || healthcare == "pharmacy" {
		return "Healthcare", "Pharmacy"
	}
	if amenity == "doctors" || healthcare == "doctor" {
		return "Healthcare", "Doctor"
	}
	if amenity == "hospital" {
		return "Healthcare", "Hospital"
	}
	if amenity == "restaurant" {
		return "Dining", "Restaurant"
	}
	if amenity == "cafe" {
		return "Dining", "Cafe"
	}
	if leisure == "fitness_centre" {
		return "Leisure", "Gym"
	}
	if amenity == "bank" {
		return "Services", "Bank"
	}
	if amenity == "post_office" {
		return "Services", "Post Office"
	}
	return "Other", amenity
}

func sortFacilitiesByDistance(facilities []Facility) {
	for i := 0; i < len(facilities); i++ {
		for j := i + 1; j < len(facilities); j++ {
			if facilities[j].Distance < facilities[i].Distance {
				facilities[i], facilities[j] = facilities[j], facilities[i]
			}
		}
	}
}

func calculateAmenitiesScore(counts map[string]int, facilities []Facility) float64 {
	// Score based on variety and proximity
	score := 0.0

	// Points for category diversity (max 40 points)
	categoryScore := float64(len(counts)) * 8
	if categoryScore > 40 {
		categoryScore = 40
	}
	score += categoryScore

	// Points for total facilities (max 30 points)
	facilityScore := float64(len(facilities)) * 2
	if facilityScore > 30 {
		facilityScore = 30
	}
	score += facilityScore

	// Points for proximity (max 30 points)
	if len(facilities) > 0 {
		avgDistance := 0.0
		for _, f := range facilities {
			avgDistance += f.Distance
		}
		avgDistance /= float64(len(facilities))

		// Closer is better: 0m = 30 points, 1500m = 0 points
		proximityScore := 30 * (1 - avgDistance/1500)
		if proximityScore < 0 {
			proximityScore = 0
		}
		score += proximityScore
	}

	return score
}

func emptyFacilitiesData() *FacilitiesData {
	return &FacilitiesData{
		TopFacilities:  []Facility{},
		AmenitiesScore: 0,
		CategoryCounts: make(map[string]int),
	}
}

// AHNHeightData represents elevation and terrain data
type AHNHeightData struct {
	Elevation     float64   `json:"elevation"`     // meters above NAP
	TerrainSlope  float64   `json:"terrainSlope"`  // degrees
	FloodRisk     string    `json:"floodRisk"`     // Low, Medium, High based on elevation
	ViewPotential string    `json:"viewPotential"` // Poor, Fair, Good, Excellent
	Surrounding   []float64 `json:"surrounding"`   // Elevations of nearby points
}

// openElevationResponse represents Open-Elevation API response
type openElevationResponse struct {
	Results []struct {
		Elevation float64 `json:"elevation"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

// FetchAHNHeightData retrieves elevation data using Open-Elevation API
// Documentation: https://open-elevation.com
func (c *ApiClient) FetchAHNHeightData(cfg *config.Config, lat, lon float64) (*AHNHeightData, error) {
	// Always use Open-Elevation API (free, no auth required)
	// Ignore config override which may have wrong WFS URLs
	openElevationURL := "https://api.open-elevation.com/api/v1/lookup"

	// Query elevation for the point and surrounding points
	url := fmt.Sprintf("%s?locations=%.6f,%.6f", openElevationURL, lat, lon)
	logutil.Debugf("[AHN] Request URL: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logutil.Debugf("[AHN] Request error: %v", err)
		return estimateElevationForAmsterdam(lat, lon), nil
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[AHN] HTTP error: %v", err)
		return estimateElevationForAmsterdam(lat, lon), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logutil.Debugf("[AHN] Non-200 status: %d", resp.StatusCode)
		return estimateElevationForAmsterdam(lat, lon), nil
	}

	var apiResp openElevationResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[AHN] Decode error: %v", err)
		return estimateElevationForAmsterdam(lat, lon), nil
	}

	if len(apiResp.Results) == 0 {
		return estimateElevationForAmsterdam(lat, lon), nil
	}

	elevation := apiResp.Results[0].Elevation
	logutil.Debugf("[AHN] Elevation: %.2f meters", elevation)

	// Assess flood risk based on elevation (NAP = Amsterdam Ordnance Datum)
	// Most of Amsterdam is below sea level (-2m to 2m)
	floodRisk := "Low"
	if elevation < -2.0 {
		floodRisk = "High"
	} else if elevation < 1.0 {
		floodRisk = "Medium"
	}

	// Assess view potential
	viewPotential := "Fair"
	if elevation > 5.0 {
		viewPotential = "Excellent"
	} else if elevation > 2.0 {
		viewPotential = "Good"
	} else if elevation < -1.0 {
		viewPotential = "Poor"
	}

	result := &AHNHeightData{
		Elevation:     elevation,
		TerrainSlope:  0.5, // Netherlands is very flat
		FloodRisk:     floodRisk,
		ViewPotential: viewPotential,
		Surrounding:   []float64{elevation - 0.5, elevation + 0.5, elevation, elevation - 0.3},
	}

	logutil.Debugf("[AHN] Result: elevation=%.2f, risk=%s", elevation, floodRisk)
	return result, nil
}

// estimateElevationForAmsterdam provides a reasonable estimate for Amsterdam area
// Most of Amsterdam is around -2m to +2m NAP
func estimateElevationForAmsterdam(lat, lon float64) *AHNHeightData {
	// Amsterdam center is approximately 52.37N, 4.89E
	// Elevation varies from about -5m (polders) to +2m (city center)

	// Simple estimation based on distance from city center
	centerLat := 52.37
	centerLon := 4.89

	// Central Amsterdam is around sea level, outskirts are lower
	distance := math.Sqrt(math.Pow(lat-centerLat, 2) + math.Pow(lon-centerLon, 2))

	// Estimate elevation: central = 0m, further out = -2m typical
	elevation := -1.0 - (distance * 10)
	if elevation < -5 {
		elevation = -5
	}
	if elevation > 2 {
		elevation = 2
	}

	floodRisk := "Medium"
	if elevation < -2.0 {
		floodRisk = "High"
	} else if elevation > 0 {
		floodRisk = "Low"
	}

	return &AHNHeightData{
		Elevation:     elevation,
		TerrainSlope:  0.5,
		FloodRisk:     floodRisk,
		ViewPotential: "Fair",
		Surrounding:   []float64{elevation - 0.5, elevation + 0.5},
	}
}
