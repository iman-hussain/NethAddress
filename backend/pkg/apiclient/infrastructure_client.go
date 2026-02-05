package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// Default PDOK API endpoints for infrastructure data (free, no auth required)
const (
	defaultBGTApiURL     = "https://api.pdok.nl/lv/bgt/ogc/v1"
	defaultNatura2000URL = "https://api.pdok.nl/rvo/natura2000/ogc/v1"
)

// extractPolygonCentroid extracts an approximate centroid from GeoJSON polygon coordinates
// Returns the centroid lat/lon and calculated area
func extractPolygonCentroid(coordsRaw json.RawMessage, geomType string) (float64, float64, float64) {
	// Parse coordinates based on geometry type
	// Polygon: [[[lon, lat], [lon, lat], ...]]
	// MultiPolygon: [[[[lon, lat], [lon, lat], ...]]]
	var coords [][][]float64

	if geomType == "MultiPolygon" {
		var multiCoords [][][][]float64
		if err := json.Unmarshal(coordsRaw, &multiCoords); err != nil || len(multiCoords) == 0 {
			return 0, 0, 0
		}
		// Use the first polygon of the multipolygon
		if len(multiCoords[0]) > 0 {
			coords = multiCoords[0]
		}
	} else {
		if err := json.Unmarshal(coordsRaw, &coords); err != nil || len(coords) == 0 {
			return 0, 0, 0
		}
	}

	if len(coords) == 0 || len(coords[0]) == 0 {
		return 0, 0, 0
	}

	// Calculate centroid and area using the outer ring (first array)
	ring := coords[0]
	var sumLat, sumLon float64
	var area float64
	n := len(ring)

	for i, point := range ring {
		if len(point) < 2 {
			continue
		}
		lon, lat := point[0], point[1]
		sumLon += lon
		sumLat += lat

		// Shoelace formula for area calculation
		if i < n-1 && len(ring[i+1]) >= 2 {
			nextLon, nextLat := ring[i+1][0], ring[i+1][1]
			area += (lon * nextLat) - (nextLon * lat)
		}
	}

	if n > 0 {
		centroidLon := sumLon / float64(n)
		centroidLat := sumLat / float64(n)
		// Area in square degrees - convert roughly to m² (very approximate)
		areaM2 := math.Abs(area/2) * 111000 * 111000
		return centroidLat, centroidLon, areaM2
	}
	return 0, 0, 0
}

// FetchGreenSpacesData retrieves parks and green areas using PDOK BGT API with retry logic
// Documentation: https://api.pdok.nl/lv/bgt/ogc/v1
func (c *ApiClient) FetchGreenSpacesData(ctx context.Context, cfg *config.Config, lat, lon float64, radius int) (*models.GreenSpacesData, error) {
	// Always use PDOK BGT API default (free, no auth) - ignore config overrides which may have bad URLs
	baseURL := defaultBGTApiURL

	// Create bounding box based on radius (convert meters to degrees approximately)
	delta := float64(radius) / 111000.0 // ~111km per degree
	bbox := fmt.Sprintf("%.6f,%.6f,%.6f,%.6f", lon-delta, lat-delta, lon+delta, lat+delta)

	// Query BGT begroeidterreindeel (vegetated terrain)
	url := fmt.Sprintf("%s/collections/begroeidterreindeel/items?bbox=%s&f=json&limit=50", baseURL, bbox)
	logutil.Debugf("[GreenSpaces] Request URL: %s", url)

	// Attempt up to 3 times with 10-second initial delay between retries
	var finalResp *http.Response
	err := c.retryWithBackoff(ctx, "GreenSpaces", 3, 10*time.Second, func(retryCtx context.Context) error {
		req, err := http.NewRequestWithContext(retryCtx, "GET", url, nil)
		if err != nil {
			logutil.Debugf("[GreenSpaces] Request error: %v", err)
			return err
		}
		req.Header.Set("Accept", "application/json")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			logutil.Debugf("[GreenSpaces] HTTP error: %v", err)
			return err
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			logutil.Debugf("[GreenSpaces] Non-200 status: %d", resp.StatusCode)
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		finalResp = resp
		return nil
	})

	if err != nil {
		logutil.Debugf("[GreenSpaces] Failed after retries: %v", err)
		return emptyGreenSpacesData(), nil
	}

	if finalResp == nil {
		return emptyGreenSpacesData(), nil
	}
	defer finalResp.Body.Close()

	var apiResp models.BgtGreenResponse
	if err := json.NewDecoder(finalResp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[GreenSpaces] Decode error: %v", err)
		return emptyGreenSpacesData(), nil
	}

	logutil.Debugf("[GreenSpaces] Found %d green areas", len(apiResp.Features))

	// Calculate total green area and categorise
	greenSpaces := make([]models.GreenSpace, 0)
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

		// Extract centroid and area from polygon geometry
		centroidLat, centroidLon, area := extractPolygonCentroid(
			feature.Geometry.Coordinates,
			feature.Geometry.Type,
		)

		// Use default if extraction failed
		if area == 0 {
			area = 500.0
		}

		// Calculate actual distance from property to centroid using Haversine
		distance := float64(radius) / 2 // Default fallback
		if centroidLat != 0 && centroidLon != 0 {
			distance = haversineDistance(lat, lon, centroidLat, centroidLon)
		}

		if distance < nearestDistance && (greenType == "Park" || greenType == "Garden") {
			nearestDistance = distance
			nearestPark = name
		}

		totalArea += area

		greenSpaces = append(greenSpaces, models.GreenSpace{
			Name:     name,
			Type:     greenType,
			Area:     area,
			Distance: distance,
			Lat:      centroidLat,
			Lon:      centroidLon,
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
	natura2000Parks := c.fetchNatura2000Areas(ctx, lat, lon, radius)
	greenSpaces = append(greenSpaces, natura2000Parks...)

	if nearestPark == "" && len(natura2000Parks) > 0 {
		nearestPark = natura2000Parks[0].Name
		nearestDistance = natura2000Parks[0].Distance
	}

	result := &models.GreenSpacesData{
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

func (c *ApiClient) fetchNatura2000Areas(ctx context.Context, lat, lon float64, radius int) []models.GreenSpace {
	delta := float64(radius) / 111000.0 * 5 // Larger search area for nature reserves
	bbox := fmt.Sprintf("%.6f,%.6f,%.6f,%.6f", lon-delta, lat-delta, lon+delta, lat+delta)

	url := fmt.Sprintf("%s/collections/natura2000/items?bbox=%s&f=json&limit=5", defaultNatura2000URL, bbox)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	var apiResp models.Natura2000Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil
	}

	parks := make([]models.GreenSpace, 0, len(apiResp.Features))
	for _, f := range apiResp.Features {
		parks = append(parks, models.GreenSpace{
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

func emptyGreenSpacesData() *models.GreenSpacesData {
	return &models.GreenSpacesData{
		TotalGreenArea:  0,
		GreenPercentage: 0,
		NearestPark:     "",
		ParkDistance:    0,
		TreeCanopyCover: 0,
		GreenSpaces:     []models.GreenSpace{},
	}
}

// FetchEducationData retrieves school locations using OSM Overpass API with retry logic
// Documentation: https://wiki.openstreetmap.org/wiki/Overpass_API
func (c *ApiClient) FetchEducationData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.EducationData, error) {
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

	// Attempt up to 3 times with 10-second initial delay between retries
	var finalResp *http.Response
	err := c.retryWithBackoff(ctx, "Education", 3, 10*time.Second, func(retryCtx context.Context) error {
		// Send query as POST body (not query string)
		reqBody := strings.NewReader("data=" + query)
		req, err := http.NewRequestWithContext(retryCtx, "POST", overpassURL, reqBody)
		if err != nil {
			logutil.Debugf("[Education] Request error: %v", err)
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			logutil.Debugf("[Education] HTTP error: %v", err)
			return err
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			logutil.Debugf("[Education] Non-200 status: %d", resp.StatusCode)
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		finalResp = resp
		return nil
	})

	if err != nil {
		logutil.Debugf("[Education] Failed after retries: %v", err)
		return emptyEducationData(), nil
	}

	if finalResp == nil {
		return emptyEducationData(), nil
	}
	defer finalResp.Body.Close()

	var apiResp models.OverpassResponse
	if err := json.NewDecoder(finalResp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[Education] Decode error: %v", err)
		return emptyEducationData(), nil
	}

	logutil.Debugf("[Education] Found %d schools", len(apiResp.Elements))

	allSchools := make([]models.School, 0, len(apiResp.Elements))
	var nearestPrimary *models.School
	var nearestSecondary *models.School
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

		school := models.School{
			Name:         elem.Tags.Name,
			Type:         schoolType,
			Distance:     distance,
			QualityScore: 7.0, // Default quality score (would need separate data source)
			Address:      address,
			Denomination: denomination,
			Lat:          elem.Lat,
			Lon:          elem.Lon,
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

	result := &models.EducationData{
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

func emptyEducationData() *models.EducationData {
	return &models.EducationData{
		NearestPrimarySchool:   nil,
		NearestSecondarySchool: nil,
		AllSchools:             []models.School{},
		AverageQuality:         0,
	}
}

// FetchBuildingPermitsData retrieves recent building activity with retry logic
// Documentation: https://api.store (CBS Building Permits)
func (c *ApiClient) FetchBuildingPermitsData(ctx context.Context, cfg *config.Config, lat, lon float64, radius int) (*models.BuildingPermitsData, error) {
	// Return empty data if not configured
	if cfg.BuildingPermitsApiURL == "" {
		return &models.BuildingPermitsData{
			TotalPermits:    0,
			NewConstruction: 0,
			Renovations:     0,
			Permits:         []models.BuildingPermit{},
			GrowthTrend:     "Unknown",
		}, nil
	}

	url := fmt.Sprintf("%s/permits?lat=%f&lon=%f&radius=%d&years=2", cfg.BuildingPermitsApiURL, lat, lon, radius)

	// Attempt up to 3 times with 10-second initial delay between retries
	var finalResp *http.Response
	err := c.retryWithBackoff(ctx, "BuildingPermits", 3, 10*time.Second, func(retryCtx context.Context) error {
		req, err := http.NewRequestWithContext(retryCtx, "GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/json")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		finalResp = resp
		return nil
	})

	if err != nil {
		logutil.Debugf("[BuildingPermits] Failed after retries: %v", err)
		return &models.BuildingPermitsData{
			TotalPermits:    0,
			NewConstruction: 0,
			Renovations:     0,
			Permits:         []models.BuildingPermit{},
			GrowthTrend:     "Unknown",
		}, nil
	}

	if finalResp == nil {
		return &models.BuildingPermitsData{
			TotalPermits:    0,
			NewConstruction: 0,
			Renovations:     0,
			Permits:         []models.BuildingPermit{},
			GrowthTrend:     "Unknown",
		}, nil
	}
	defer finalResp.Body.Close()

	var result models.BuildingPermitsData
	if err := json.NewDecoder(finalResp.Body).Decode(&result); err != nil {
		return &models.BuildingPermitsData{
			TotalPermits:    0,
			NewConstruction: 0,
			Renovations:     0,
			Permits:         []models.BuildingPermit{},
			GrowthTrend:     "Unknown",
		}, nil
	}

	return &result, nil
}

// FetchFacilitiesData retrieves nearby amenities using OSM Overpass API with retry logic
// Documentation: https://wiki.openstreetmap.org/wiki/Overpass_API
func (c *ApiClient) FetchFacilitiesData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.FacilitiesData, error) {
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

	// Attempt up to 3 times with 10-second initial delay between retries
	var finalResp *http.Response
	err := c.retryWithBackoff(ctx, "Facilities", 3, 10*time.Second, func(retryCtx context.Context) error {
		// Send query as POST body (not query string)
		reqBody := strings.NewReader("data=" + query)
		req, err := http.NewRequestWithContext(retryCtx, "POST", overpassURL, reqBody)
		if err != nil {
			logutil.Debugf("[Facilities] Request error: %v", err)
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			logutil.Debugf("[Facilities] HTTP error: %v", err)
			return err
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			logutil.Debugf("[Facilities] Non-200 status: %d", resp.StatusCode)
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		finalResp = resp
		return nil
	})

	if err != nil {
		logutil.Debugf("[Facilities] Failed after retries: %v", err)
		return emptyFacilitiesData(), nil
	}

	if finalResp == nil {
		return emptyFacilitiesData(), nil
	}
	defer finalResp.Body.Close()

	var apiResp models.OverpassFacilitiesResponse
	if err := json.NewDecoder(finalResp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[Facilities] Decode error: %v", err)
		return emptyFacilitiesData(), nil
	}

	logutil.Debugf("[Facilities] Found %d amenities", len(apiResp.Elements))

	facilities := make([]models.Facility, 0, len(apiResp.Elements))
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

		facility := models.Facility{
			Name:      name,
			Category:  category,
			Type:      facilityType,
			Distance:  distance,
			WalkTime:  walkTime,
			DriveTime: driveTime,
			Rating:    4.0, // Default - would need external API
			Address:   "",  // Address not parsed for facilities in this simple version
			Lat:       elem.Lat,
			Lon:       elem.Lon,
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

	result := &models.FacilitiesData{
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

func sortFacilitiesByDistance(facilities []models.Facility) {
	for i := 0; i < len(facilities); i++ {
		for j := i + 1; j < len(facilities); j++ {
			if facilities[j].Distance < facilities[i].Distance {
				facilities[i], facilities[j] = facilities[j], facilities[i]
			}
		}
	}
}

func calculateAmenitiesScore(counts map[string]int, facilities []models.Facility) float64 {
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

func emptyFacilitiesData() *models.FacilitiesData {
	return &models.FacilitiesData{
		TopFacilities:  []models.Facility{},
		AmenitiesScore: 0,
		CategoryCounts: make(map[string]int),
	}
}

// FetchAHNHeightData retrieves elevation data using Open-Elevation API
// Documentation: https://open-elevation.com
func (c *ApiClient) FetchAHNHeightData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.AHNHeightData, error) {
	// Always use Open-Elevation API (free, no auth required)
	// Ignore config override which may have wrong WFS URLs
	openElevationURL := "https://api.open-elevation.com/api/v1/lookup"

	// Query elevation for the point and surrounding points
	url := fmt.Sprintf("%s?locations=%.6f,%.6f", openElevationURL, lat, lon)
	logutil.Debugf("[AHN] Request URL: %s", url)

	var apiResp models.OpenElevationResponse
	if err := c.GetJSON(ctx, "AHN", url, nil, &apiResp); err != nil {
		// Fallback to estimate on API failure
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

	result := &models.AHNHeightData{
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
func estimateElevationForAmsterdam(lat, lon float64) *models.AHNHeightData {
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

	return &models.AHNHeightData{
		Elevation:     elevation,
		TerrainSlope:  0.5,
		FloodRisk:     floodRisk,
		ViewPotential: "Fair",
		Surrounding:   []float64{elevation - 0.5, elevation + 0.5},
	}
}
