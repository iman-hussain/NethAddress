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

// FetchNDWTrafficData retrieves real-time traffic data for accessibility scoring
// Documentation: https://opendata.ndw.nu
func (c *ApiClient) FetchNDWTrafficData(ctx context.Context, cfg *config.Config, lat, lon float64, radius int) ([]models.NDWTrafficData, error) {
	// NDW requires registration - return empty data if not configured
	if cfg.NDWTrafficApiURL == "" {
		logutil.Debugf("[NDW Traffic] No API URL configured, returning empty data")
		return []models.NDWTrafficData{}, nil
	}

	// Query traffic data within radius (meters) of location
	url := fmt.Sprintf("%s/traffic?lat=%f&lon=%f&radius=%d", cfg.NDWTrafficApiURL, lat, lon, radius)
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
		return []models.NDWTrafficData{}, nil
	}

	var result struct {
		Data []models.NDWTrafficData `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode NDW traffic response: %w", err)
	}

	return result.Data, nil
}

// FetchOpenOVData retrieves public transport data using OSM Overpass API
// Documentation: https://wiki.openstreetmap.org/wiki/Overpass_API
func (c *ApiClient) FetchOpenOVData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.OpenOVTransportData, error) {
	// Primary endpoint gets special treatment with longer retry
	primaryEndpoint := "https://overpass-api.de/api/interpreter"
	// Fallback endpoints tried every 3 seconds
	fallbackEndpoints := []string{
		"https://overpass.kumi.systems/api/interpreter",
		"https://maps.mail.ru/osm/tools/overpass/api/interpreter",
	}

	// Query for public transport stops within 1km radius
	radius := 1000
	query := fmt.Sprintf(`[out:json][timeout:15];
(
  node["highway"="bus_stop"](around:%d,%.6f,%.6f);
  node["railway"="tram_stop"](around:%d,%.6f,%.6f);
  node["railway"="station"](around:%d,%.6f,%.6f);
  node["railway"="halt"](around:%d,%.6f,%.6f);
  node["public_transport"="stop_position"](around:%d,%.6f,%.6f);
  node["public_transport"="platform"](around:%d,%.6f,%.6f);
  node["amenity"="bus_station"](around:%d,%.6f,%.6f);
);
out body qt 50;`, radius, lat, lon, radius, lat, lon, radius, lat, lon, radius, lat, lon, radius, lat, lon, radius, lat, lon, radius, lat, lon)

	logutil.Debugf("[OpenOV] Querying Overpass API for PT stops near %.6f, %.6f", lat, lon)

	// Helper to execute query against an endpoint
	queryEndpoint := func(url string) (*models.OverpassTransportResponse, error) {
		reqBody := strings.NewReader("data=" + query)
		req, err := http.NewRequestWithContext(ctx, "POST", url, reqBody)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "AddressIQ/1.0 (https://github.com/iman-hussain/AddressIQ)")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			return nil, fmt.Errorf("rate limited")
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		var result models.OverpassTransportResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode error: %w", err)
		}
		return &result, nil
	}

	var apiResp *models.OverpassTransportResponse
	var lastErr error

	// Step 1: Try primary endpoint first
	logutil.Debugf("[OpenOV] Attempt 1: querying primary endpoint %s", primaryEndpoint)
	apiResp, lastErr = queryEndpoint(primaryEndpoint)
	if lastErr == nil && len(apiResp.Elements) > 0 {
		logutil.Debugf("[OpenOV] Primary endpoint succeeded on first attempt, found %d elements", len(apiResp.Elements))
		return c.processTransportStops(lat, lon, apiResp)
	}
	if lastErr != nil {
		logutil.Debugf("[OpenOV] Primary attempt 1 failed: %v", lastErr)
	}

	// Step 2: Wait 10 seconds, try primary again
	select {
	case <-ctx.Done():
		logutil.Debugf("[OpenOV] Context cancelled during 10s wait")
		return emptyTransportData(), nil
	case <-time.After(10 * time.Second):
	}

	logutil.Debugf("[OpenOV] Attempt 2: retrying primary endpoint after 10s wait")
	apiResp, lastErr = queryEndpoint(primaryEndpoint)
	if lastErr == nil && len(apiResp.Elements) > 0 {
		logutil.Debugf("[OpenOV] Primary endpoint succeeded on second attempt, found %d elements", len(apiResp.Elements))
		return c.processTransportStops(lat, lon, apiResp)
	}
	if lastErr != nil {
		logutil.Debugf("[OpenOV] Primary attempt 2 failed: %v", lastErr)
	}

	// Step 3: Try fallback endpoints every 3 seconds
	for _, fallbackURL := range fallbackEndpoints {
		select {
		case <-ctx.Done():
			logutil.Debugf("[OpenOV] Context cancelled before fallback attempt")
			return emptyTransportData(), nil
		case <-time.After(3 * time.Second):
		}

		logutil.Debugf("[OpenOV] Trying fallback endpoint: %s", fallbackURL)
		apiResp, lastErr = queryEndpoint(fallbackURL)
		if lastErr == nil && len(apiResp.Elements) > 0 {
			logutil.Debugf("[OpenOV] Fallback %s succeeded, found %d elements", fallbackURL, len(apiResp.Elements))
			return c.processTransportStops(lat, lon, apiResp)
		}
		if lastErr != nil {
			logutil.Debugf("[OpenOV] Fallback %s failed: %v", fallbackURL, lastErr)
		}
	}

	logutil.Debugf("[OpenOV] All endpoints failed, last error: %v", lastErr)
	return emptyTransportData(), nil
}

// processTransportStops converts Overpass response to OpenOVTransportData
func (c *ApiClient) processTransportStops(lat, lon float64, apiResp *models.OverpassTransportResponse) (*models.OpenOVTransportData, error) {
	logutil.Debugf("[OpenOV] Processing %d PT stops", len(apiResp.Elements))

	stops := make([]models.PublicTransportStop, 0, len(apiResp.Elements))
	for _, elem := range apiResp.Elements {
		distance := haversineDistanceTraffic(lat, lon, elem.Lat, elem.Lon)

		stopType := determineStopType(elem.Tags.Highway, elem.Tags.Railway, elem.Tags.PublicTransport)
		name := elem.Tags.Name
		if name == "" {
			name = fmt.Sprintf("%s stop", stopType)
		}

		stop := models.PublicTransportStop{
			StopID:   fmt.Sprintf("%d", elem.ID),
			Name:     name,
			Type:     stopType,
			Distance: distance,
			Lines:    []string{}, // Would need real-time API for line info
		}
		stop.Coordinates.Lat = elem.Lat
		stop.Coordinates.Lon = elem.Lon

		stops = append(stops, stop)
	}

	// Sort by distance and limit to nearest 10
	sortStopsByDistance(stops)
	if len(stops) > 10 {
		stops = stops[:10]
	}

	result := &models.OpenOVTransportData{
		NearestStops: stops,
		Connections:  []models.Connection{}, // Would need real-time API
	}

	logutil.Debugf("[OpenOV] Result: %d stops found", len(stops))
	return result, nil
}

func determineStopType(highway, railway, publicTransport string) string {
	if railway == "station" {
		return "Train"
	}
	if railway == "tram_stop" {
		return "Tram"
	}
	if railway == "halt" {
		return "Train"
	}
	if highway == "bus_stop" {
		return "Bus"
	}
	if publicTransport == "stop_position" {
		return "Bus" // Default for generic stops
	}
	return "Bus"
}

func sortStopsByDistance(stops []models.PublicTransportStop) {
	// Simple bubble sort for small arrays
	for i := 0; i < len(stops); i++ {
		for j := i + 1; j < len(stops); j++ {
			if stops[j].Distance < stops[i].Distance {
				stops[i], stops[j] = stops[j], stops[i]
			}
		}
	}
}

// haversineDistanceTraffic calculates distance in meters
func haversineDistanceTraffic(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000

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

func emptyTransportData() *models.OpenOVTransportData {
	return &models.OpenOVTransportData{
		NearestStops: []models.PublicTransportStop{},
		Connections:  []models.Connection{},
	}
}

// FetchParkingData retrieves parking availability for convenience scoring
// Documentation: Municipal API (varies by city)
func (c *ApiClient) FetchParkingData(ctx context.Context, cfg *config.Config, lat, lon float64, radius int) (*models.ParkingData, error) {
	// Return empty data if not configured
	if cfg.ParkingApiURL == "" {
		return &models.ParkingData{
			TotalSpaces:     0,
			AvailableSpaces: 0,
			ParkingZones:    []models.ParkingZone{},
		}, nil
	}

	url := fmt.Sprintf("%s/parking?lat=%f&lon=%f&radius=%d", cfg.ParkingApiURL, lat, lon, radius)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &models.ParkingData{
			TotalSpaces:     0,
			AvailableSpaces: 0,
			ParkingZones:    []models.ParkingZone{},
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return &models.ParkingData{
			TotalSpaces:     0,
			AvailableSpaces: 0,
			ParkingZones:    []models.ParkingZone{},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Return empty data for any non-200 status (including 404)
		return &models.ParkingData{
			TotalSpaces:     0,
			AvailableSpaces: 0,
			ParkingZones:    []models.ParkingZone{},
		}, nil
	}

	var result models.ParkingData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &models.ParkingData{
			TotalSpaces:     0,
			AvailableSpaces: 0,
			ParkingZones:    []models.ParkingZone{},
		}, nil
	}

	return &result, nil
}
