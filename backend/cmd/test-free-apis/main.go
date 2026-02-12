package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type TestResult struct {
	Name           string
	URL            string
	Success        bool
	Message        string
	Details        string
	ResponseSample string
}

type Coordinates struct {
	Lat float64
	Lon float64
}

func main() {
	// Try multiple .env locations
	godotenv.Load(".env")
	godotenv.Load("../.env")
	godotenv.Load("../../.env")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	fmt.Println("🔍 NethAddress Free API Connectivity Test")
	fmt.Println("========================================")
	fmt.Println()

	testPostcode := "1012LG"
	testHouseNum := "1"

	fmt.Printf("Test address: %s %s\n\n", testPostcode, testHouseNum)

	results := []TestResult{}

	// Step 1: Get coordinates from BAG (required for most other APIs)
	bagResult, coords := testBAG(ctx, testPostcode, testHouseNum)
	results = append(results, bagResult)
	printResult(bagResult)

	if !bagResult.Success {
		fmt.Println("\n❌ Cannot proceed without coordinates from BAG. Stopping.")
		os.Exit(1)
	}

	fmt.Printf("\n✅ Coordinates obtained: %.6f, %.6f\n\n", coords.Lat, coords.Lon)

	// Step 2: Test APIs that need coordinates
	results = append(results, testOpenMeteoWeather(ctx, coords))
	printResult(results[len(results)-1])

	results = append(results, testOpenMeteoSolar(ctx, coords))
	printResult(results[len(results)-1])

	results = append(results, testLuchtmeetnet(ctx, coords))
	printResult(results[len(results)-1])

	// Step 3: Test REST APIs that don't need coordinates
	results = append(results, testOpenOV(ctx))
	printResult(results[len(results)-1])

	results = append(results, testAmsterdamParking(ctx))
	printResult(results[len(results)-1])

	results = append(results, testAmsterdamEducation(ctx))
	printResult(results[len(results)-1])

	results = append(results, testAmsterdamFacilities(ctx))
	printResult(results[len(results)-1])

	results = append(results, testAmsterdamMonumenten(ctx))
	printResult(results[len(results)-1])

	// Step 4: Test WFS services (need coordinates for spatial queries)
	results = append(results, testCBSPopulation(ctx, coords))
	printResult(results[len(results)-1])

	results = append(results, testCBSSquareStats(ctx, coords))
	printResult(results[len(results)-1])

	results = append(results, testBROSoil(ctx, coords))
	printResult(results[len(results)-1])

	results = append(results, testFloodRisk(ctx, coords))
	printResult(results[len(results)-1])

	results = append(results, testGreenSpaces(ctx, coords))
	printResult(results[len(results)-1])

	results = append(results, testAHN(ctx, coords))
	printResult(results[len(results)-1])

	results = append(results, testLandUse(ctx, coords))
	printResult(results[len(results)-1])

	// Summary
	fmt.Println("\n========================================")
	fmt.Println("📊 SUMMARY")
	fmt.Println("========================================")

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	fmt.Printf("Total APIs tested: %d\n", len(results))
	fmt.Printf("✅ Working: %d\n", successCount)
	fmt.Printf("❌ Failed: %d\n\n", len(results)-successCount)

	if successCount < len(results) {
		fmt.Println("Failed APIs:")
		for _, r := range results {
			if !r.Success {
				fmt.Printf("  - %s: %s\n", r.Name, r.Message)
			}
		}
	}
}

func printResult(r TestResult) {
	status := "❌"
	if r.Success {
		status = "✅"
	}
	fmt.Printf("%s %s\n", status, r.Name)
	fmt.Printf("   URL: %s\n", truncate(r.URL, 80))
	fmt.Printf("   %s\n", r.Message)
	if r.Details != "" {
		fmt.Printf("   Details: %s\n", r.Details)
	}
	if r.ResponseSample != "" && r.Success {
		fmt.Printf("   Sample: %s\n", truncate(r.ResponseSample, 100))
	}
	fmt.Println()
}

func testBAG(ctx context.Context, postcode, houseNum string) (TestResult, Coordinates) {
	endpoint := os.Getenv("BAG_API_URL")
	if endpoint == "" {
		return TestResult{
			Name:    "BAG Locatieserver",
			Success: false,
			Message: "BAG_API_URL not set",
		}, Coordinates{}
	}

	query := url.Values{}
	query.Set("q", fmt.Sprintf("%s %s", postcode, houseNum))
	query.Set("rows", "1")

	fullURL := endpoint + "?" + query.Encode()

	body, status, err := doGet(ctx, fullURL)
	if err != nil {
		return TestResult{
			Name:    "BAG Locatieserver",
			URL:     fullURL,
			Success: false,
			Message: fmt.Sprintf("Request failed: %v", err),
		}, Coordinates{}
	}

	if status != 200 {
		return TestResult{
			Name:    "BAG Locatieserver",
			URL:     fullURL,
			Success: false,
			Message: fmt.Sprintf("HTTP %d", status),
		}, Coordinates{}
	}

	var bagResp struct {
		Response struct {
			NumFound int `json:"numFound"`
			Docs     []struct {
				Weergavenaam string `json:"weergavenaam"`
				CentroidLL   string `json:"centroide_ll"`
			} `json:"docs"`
		} `json:"response"`
	}

	if err := json.Unmarshal(body, &bagResp); err != nil {
		return TestResult{
			Name:    "BAG Locatieserver",
			URL:     fullURL,
			Success: false,
			Message: fmt.Sprintf("JSON parse error: %v", err),
		}, Coordinates{}
	}

	if bagResp.Response.NumFound == 0 {
		return TestResult{
			Name:    "BAG Locatieserver",
			URL:     fullURL,
			Success: false,
			Message: "No address found",
		}, Coordinates{}
	}

	doc := bagResp.Response.Docs[0]
	coords, err := parseWKTPoint(doc.CentroidLL)
	if err != nil {
		return TestResult{
			Name:    "BAG Locatieserver",
			URL:     fullURL,
			Success: false,
			Message: fmt.Sprintf("Coordinate parse error: %v", err),
		}, Coordinates{}
	}

	return TestResult{
		Name:           "BAG Locatieserver",
		URL:            fullURL,
		Success:        true,
		Message:        "Address resolved",
		Details:        fmt.Sprintf("Found: %s", doc.Weergavenaam),
		ResponseSample: doc.Weergavenaam,
	}, coords
}

func testOpenMeteoWeather(ctx context.Context, coords Coordinates) TestResult {
	endpoint := os.Getenv("KNMI_WEATHER_API_URL")
	if endpoint == "" {
		return TestResult{Name: "Open-Meteo Weather", Success: false, Message: "KNMI_WEATHER_API_URL not set"}
	}

	query := url.Values{}
	query.Set("latitude", fmt.Sprintf("%.5f", coords.Lat))
	query.Set("longitude", fmt.Sprintf("%.5f", coords.Lon))
	query.Set("current_weather", "true")

	fullURL := endpoint + "?" + query.Encode()

	body, status, err := doGet(ctx, fullURL)
	if err != nil {
		return TestResult{Name: "Open-Meteo Weather", URL: fullURL, Success: false, Message: err.Error()}
	}

	if status != 200 {
		return TestResult{Name: "Open-Meteo Weather", URL: fullURL, Success: false, Message: fmt.Sprintf("HTTP %d", status)}
	}

	var weatherResp struct {
		CurrentWeather struct {
			Temperature float64 `json:"temperature"`
			Windspeed   float64 `json:"windspeed"`
		} `json:"current_weather"`
	}

	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return TestResult{Name: "Open-Meteo Weather", URL: fullURL, Success: false, Message: err.Error()}
	}

	return TestResult{
		Name:    "Open-Meteo Weather",
		URL:     fullURL,
		Success: true,
		Message: "Weather data received",
		Details: fmt.Sprintf("Temp: %.1f°C, Wind: %.1f m/s", weatherResp.CurrentWeather.Temperature, weatherResp.CurrentWeather.Windspeed),
	}
}

func testOpenMeteoSolar(ctx context.Context, coords Coordinates) TestResult {
	endpoint := os.Getenv("KNMI_SOLAR_API_URL")
	if endpoint == "" {
		return TestResult{Name: "Open-Meteo Solar", Success: false, Message: "KNMI_SOLAR_API_URL not set"}
	}

	query := url.Values{}
	query.Set("latitude", fmt.Sprintf("%.5f", coords.Lat))
	query.Set("longitude", fmt.Sprintf("%.5f", coords.Lon))
	query.Set("hourly", "shortwave_radiation")
	query.Set("forecast_days", "1")

	fullURL := endpoint + "?" + query.Encode()

	body, status, err := doGet(ctx, fullURL)
	if err != nil {
		return TestResult{Name: "Open-Meteo Solar", URL: fullURL, Success: false, Message: err.Error()}
	}

	if status != 200 {
		return TestResult{Name: "Open-Meteo Solar", URL: fullURL, Success: false, Message: fmt.Sprintf("HTTP %d", status)}
	}

	var solarResp struct {
		Hourly struct {
			Radiation []float64 `json:"shortwave_radiation"`
		} `json:"hourly"`
	}

	if err := json.Unmarshal(body, &solarResp); err != nil {
		return TestResult{Name: "Open-Meteo Solar", URL: fullURL, Success: false, Message: err.Error()}
	}

	if len(solarResp.Hourly.Radiation) == 0 {
		return TestResult{Name: "Open-Meteo Solar", URL: fullURL, Success: false, Message: "No radiation data"}
	}

	return TestResult{
		Name:    "Open-Meteo Solar",
		URL:     fullURL,
		Success: true,
		Message: "Solar forecast received",
		Details: fmt.Sprintf("%d hourly values, first: %.1f W/m²", len(solarResp.Hourly.Radiation), solarResp.Hourly.Radiation[0]),
	}
}

func testLuchtmeetnet(ctx context.Context, coords Coordinates) TestResult {
	endpoint := os.Getenv("LUCHTMEETNET_API_URL")
	if endpoint == "" {
		return TestResult{Name: "Luchtmeetnet Air Quality", Success: false, Message: "LUCHTMEETNET_API_URL not set"}
	}

	fullURL := strings.TrimSuffix(endpoint, "/") + "/stations"

	body, status, err := doGet(ctx, fullURL)
	if err != nil {
		return TestResult{Name: "Luchtmeetnet Air Quality", URL: fullURL, Success: false, Message: err.Error()}
	}

	if status != 200 {
		return TestResult{Name: "Luchtmeetnet Air Quality", URL: fullURL, Success: false, Message: fmt.Sprintf("HTTP %d", status)}
	}

	var stationsResp struct {
		Data []struct {
			Number    string  `json:"number"`
			Location  string  `json:"location"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &stationsResp); err != nil {
		return TestResult{Name: "Luchtmeetnet Air Quality", URL: fullURL, Success: false, Message: err.Error()}
	}

	if len(stationsResp.Data) == 0 {
		return TestResult{Name: "Luchtmeetnet Air Quality", URL: fullURL, Success: false, Message: "No stations found"}
	}

	// Find nearest station
	nearest := stationsResp.Data[0]
	minDist := haversine(coords.Lat, coords.Lon, nearest.Latitude, nearest.Longitude)

	for _, station := range stationsResp.Data[1:] {
		dist := haversine(coords.Lat, coords.Lon, station.Latitude, station.Longitude)
		if dist < minDist {
			minDist = dist
			nearest = station
		}
	}

	return TestResult{
		Name:    "Luchtmeetnet Air Quality",
		URL:     fullURL,
		Success: true,
		Message: "Air quality stations available",
		Details: fmt.Sprintf("Nearest: %s (%.2f km away), %d total stations", nearest.Location, minDist, len(stationsResp.Data)),
	}
}

func testOpenOV(ctx context.Context) TestResult {
	endpoint := os.Getenv("OPENOV_API_URL")
	if endpoint == "" {
		return TestResult{Name: "openOV Public Transport", Success: false, Message: "OPENOV_API_URL not set"}
	}

	// Test just the base endpoint to see if it responds
	fullURL := strings.TrimSuffix(endpoint, "/")

	body, status, err := doGet(ctx, fullURL)
	if err != nil {
		return TestResult{Name: "openOV Public Transport", URL: fullURL, Success: false, Message: err.Error()}
	}

	if status != 200 {
		return TestResult{Name: "openOV Public Transport", URL: fullURL, Success: false, Message: fmt.Sprintf("HTTP %d", status)}
	}

	// Parse as generic JSON to count stop areas
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return TestResult{Name: "openOV Public Transport", URL: fullURL, Success: false, Message: err.Error()}
	}

	return TestResult{
		Name:    "openOV Public Transport",
		URL:     fullURL,
		Success: true,
		Message: "Transit data available",
		Details: fmt.Sprintf("%d stop areas returned", len(data)),
	}
}

func testAmsterdamParking(ctx context.Context) TestResult {
	endpoint := os.Getenv("PARKING_API_URL")
	if endpoint == "" {
		return TestResult{Name: "Amsterdam Parking", Success: false, Message: "PARKING_API_URL not set"}
	}

	body, status, err := doGet(ctx, endpoint)
	if err != nil {
		return TestResult{Name: "Amsterdam Parking", URL: endpoint, Success: false, Message: err.Error()}
	}

	if status != 200 {
		return TestResult{Name: "Amsterdam Parking", URL: endpoint, Success: false, Message: fmt.Sprintf("HTTP %d", status)}
	}

	// Try to parse count or results array
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return TestResult{Name: "Amsterdam Parking", URL: endpoint, Success: false, Message: err.Error()}
	}

	count := 0
	if results, ok := resp["results"].([]interface{}); ok {
		count = len(results)
	} else if c, ok := resp["count"].(float64); ok {
		count = int(c)
	}

	return TestResult{
		Name:    "Amsterdam Parking",
		URL:     endpoint,
		Success: true,
		Message: "Parking data available",
		Details: fmt.Sprintf("%d parking locations", count),
	}
}

func testAmsterdamEducation(ctx context.Context) TestResult {
	endpoint := os.Getenv("EDUCATION_API_URL")
	if endpoint == "" {
		return TestResult{Name: "Amsterdam Education", Success: false, Message: "EDUCATION_API_URL not set"}
	}

	body, status, err := doGet(ctx, endpoint)
	if err != nil {
		return TestResult{Name: "Amsterdam Education", URL: endpoint, Success: false, Message: err.Error()}
	}

	if status != 200 {
		return TestResult{Name: "Amsterdam Education", URL: endpoint, Success: false, Message: fmt.Sprintf("HTTP %d", status)}
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return TestResult{Name: "Amsterdam Education", URL: endpoint, Success: false, Message: err.Error()}
	}

	count := 0
	if results, ok := resp["results"].([]interface{}); ok {
		count = len(results)
	} else if c, ok := resp["count"].(float64); ok {
		count = int(c)
	}

	return TestResult{
		Name:    "Amsterdam Education",
		URL:     endpoint,
		Success: true,
		Message: "Education facilities available",
		Details: fmt.Sprintf("%d schools/facilities", count),
	}
}

func testAmsterdamFacilities(ctx context.Context) TestResult {
	endpoint := os.Getenv("FACILITIES_API_URL")
	if endpoint == "" {
		return TestResult{Name: "Amsterdam Facilities", Success: false, Message: "FACILITIES_API_URL not set"}
	}

	body, status, err := doGet(ctx, endpoint)
	if err != nil {
		return TestResult{Name: "Amsterdam Facilities", URL: endpoint, Success: false, Message: err.Error()}
	}

	if status != 200 {
		return TestResult{Name: "Amsterdam Facilities", URL: endpoint, Success: false, Message: fmt.Sprintf("HTTP %d", status)}
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return TestResult{Name: "Amsterdam Facilities", URL: endpoint, Success: false, Message: err.Error()}
	}

	count := 0
	if results, ok := resp["results"].([]interface{}); ok {
		count = len(results)
	} else if c, ok := resp["count"].(float64); ok {
		count = int(c)
	}

	return TestResult{
		Name:    "Amsterdam Facilities",
		URL:     endpoint,
		Success: true,
		Message: "Facilities/amenities available",
		Details: fmt.Sprintf("%d facilities", count),
	}
}

func testAmsterdamMonumenten(ctx context.Context) TestResult {
	endpoint := os.Getenv("MONUMENTEN_API_URL")
	if endpoint == "" {
		return TestResult{Name: "Amsterdam Monumenten", Success: false, Message: "MONUMENTEN_API_URL not set"}
	}

	body, status, err := doGet(ctx, endpoint)
	if err != nil {
		return TestResult{Name: "Amsterdam Monumenten", URL: endpoint, Success: false, Message: err.Error()}
	}

	if status != 200 {
		return TestResult{Name: "Amsterdam Monumenten", URL: endpoint, Success: false, Message: fmt.Sprintf("HTTP %d", status)}
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return TestResult{Name: "Amsterdam Monumenten", URL: endpoint, Success: false, Message: err.Error()}
	}

	count := 0
	if results, ok := resp["results"].([]interface{}); ok {
		count = len(results)
	} else if c, ok := resp["count"].(float64); ok {
		count = int(c)
	}

	return TestResult{
		Name:    "Amsterdam Monumenten",
		URL:     endpoint,
		Success: true,
		Message: "Monument data available",
		Details: fmt.Sprintf("%d monuments", count),
	}
}

func testCBSPopulation(ctx context.Context, coords Coordinates) TestResult {
	return testWFS(ctx, "CBS_POPULATION_API_URL", "CBS Population Grid", coords)
}

func testCBSSquareStats(ctx context.Context, coords Coordinates) TestResult {
	return testWFS(ctx, "CBS_SQUARE_STATS_API_URL", "CBS Square Statistics", coords)
}

func testBROSoil(ctx context.Context, coords Coordinates) TestResult {
	return testWFS(ctx, "BRO_SOIL_MAP_API_URL", "BRO Soil Map", coords)
}

func testFloodRisk(ctx context.Context, coords Coordinates) TestResult {
	return testWFS(ctx, "FLOOD_RISK_API_URL", "Flood Risk (Rijkswaterstaat)", coords)
}

func testGreenSpaces(ctx context.Context, coords Coordinates) TestResult {
	return testWFS(ctx, "GREEN_SPACES_API_URL", "Green Spaces", coords)
}

func testAHN(ctx context.Context, coords Coordinates) TestResult {
	return testWFS(ctx, "AHN_HEIGHT_MODEL_API_URL", "AHN Height Model", coords)
}

func testLandUse(ctx context.Context, coords Coordinates) TestResult {
	return testWFS(ctx, "LAND_USE_API_URL", "Land Use & Zoning", coords)
}

func testWFS(ctx context.Context, envVar, name string, coords Coordinates) TestResult {
	endpoint := os.Getenv(envVar)
	if endpoint == "" {
		return TestResult{Name: name, Success: false, Message: fmt.Sprintf("%s not set", envVar)}
	}

	// Strip query params for clean base URL
	baseURL := endpoint
	if idx := strings.Index(baseURL, "?"); idx != -1 {
		baseURL = baseURL[:idx]
	}

	// Get capabilities to find available feature types
	capURL := baseURL + "?service=WFS&request=GetCapabilities&version=2.0.0"
	capBody, status, err := doGet(ctx, capURL)
	if err != nil {
		return TestResult{Name: name, URL: capURL, Success: false, Message: fmt.Sprintf("GetCapabilities failed: %v", err)}
	}

	if status != 200 {
		return TestResult{Name: name, URL: capURL, Success: false, Message: fmt.Sprintf("GetCapabilities HTTP %d", status)}
	}

	featureType, err := extractFirstFeatureType(capBody)
	if err != nil {
		return TestResult{Name: name, URL: capURL, Success: false, Message: fmt.Sprintf("Parse capabilities: %v", err)}
	}

	// Try GetFeature with bbox around our point
	bbox := fmt.Sprintf("%.6f,%.6f,%.6f,%.6f",
		coords.Lon-0.01, coords.Lat-0.01, coords.Lon+0.01, coords.Lat+0.01)

	featureParams := url.Values{}
	featureParams.Set("service", "WFS")
	featureParams.Set("request", "GetFeature")
	featureParams.Set("version", "2.0.0")
	featureParams.Set("typeNames", featureType)
	featureParams.Set("bbox", bbox)
	featureParams.Set("srsName", "EPSG:4326")
	featureParams.Set("count", "1")
	featureParams.Set("outputFormat", "application/json")

	featureURL := baseURL + "?" + featureParams.Encode()
	featureBody, featureStatus, err := doGet(ctx, featureURL)
	if err != nil {
		return TestResult{Name: name, URL: featureURL, Success: false, Message: fmt.Sprintf("GetFeature failed: %v", err)}
	}

	if featureStatus != 200 {
		return TestResult{Name: name, URL: featureURL, Success: false, Message: fmt.Sprintf("GetFeature HTTP %d", featureStatus)}
	}

	return TestResult{
		Name:    name,
		URL:     featureURL,
		Success: true,
		Message: "WFS service operational",
		Details: fmt.Sprintf("FeatureType: %s, Response: %d bytes", featureType, len(featureBody)),
	}
}

func extractFirstFeatureType(xmlData []byte) (string, error) {
	type FeatureTypeList struct {
		FeatureTypes []struct {
			Name string `xml:"Name"`
		} `xml:"FeatureType"`
	}

	var capabilities struct {
		FeatureTypeList FeatureTypeList `xml:"FeatureTypeList"`
	}

	if err := xml.Unmarshal(xmlData, &capabilities); err != nil {
		return "", err
	}

	if len(capabilities.FeatureTypeList.FeatureTypes) == 0 {
		return "", fmt.Errorf("no feature types found in capabilities")
	}

	return capabilities.FeatureTypeList.FeatureTypes[0].Name, nil
}

func doGet(ctx context.Context, url string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("User-Agent", "NethAddress-API-Test/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024)) // 2MB limit
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}

func parseWKTPoint(wkt string) (Coordinates, error) {
	// POINT(lon lat)
	wkt = strings.TrimSpace(wkt)
	if !strings.HasPrefix(wkt, "POINT") {
		return Coordinates{}, fmt.Errorf("not a POINT geometry: %s", wkt)
	}

	start := strings.Index(wkt, "(")
	end := strings.Index(wkt, ")")
	if start == -1 || end == -1 {
		return Coordinates{}, fmt.Errorf("invalid POINT syntax: %s", wkt)
	}

	coords := strings.Fields(wkt[start+1 : end])
	if len(coords) != 2 {
		return Coordinates{}, fmt.Errorf("POINT should have 2 coordinates: %s", wkt)
	}

	var lon, lat float64
	fmt.Sscanf(coords[0], "%f", &lon)
	fmt.Sscanf(coords[1], "%f", &lat)

	return Coordinates{Lat: lat, Lon: lon}, nil
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func toRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
