package apiclient

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
)

// NDWTrafficData represents real-time traffic data
type NDWTrafficData struct {
	LocationID      string  `json:"locationId"`
	Intensity       int     `json:"intensity"`       // vehicles/hour
	AverageSpeed    float64 `json:"averageSpeed"`    // km/h
	CongestionLevel string  `json:"congestionLevel"` // Free, Light, Moderate, Heavy, Jammed
	LastUpdated     string  `json:"lastUpdated"`
	Coordinates     struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coordinates"`
}

// FetchNDWTrafficData retrieves real-time traffic data for accessibility scoring
// Documentation: https://opendata.ndw.nu
func (c *ApiClient) FetchNDWTrafficData(cfg *config.Config, lat, lon float64, radius int) ([]NDWTrafficData, error) {
	// NDW requires registration - return empty data if not configured
	if cfg.NDWTrafficApiURL == "" {
		logutil.Debugf("[NDW Traffic] No API URL configured, returning empty data")
		return []NDWTrafficData{}, nil
	}

	// Query traffic data within radius (meters) of location
	url := fmt.Sprintf("%s/traffic?lat=%f&lon=%f&radius=%d", cfg.NDWTrafficApiURL, lat, lon, radius)
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
		return []NDWTrafficData{}, nil
	}

	var result struct {
		Data []NDWTrafficData `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode NDW traffic response: %w", err)
	}

	return result.Data, nil
}

// OpenOVTransportData represents public transport accessibility
type OpenOVTransportData struct {
	NearestStops []PublicTransportStop `json:"nearestStops"`
	Connections  []Connection          `json:"connections"`
}

// PublicTransportStop represents a PT stop
type PublicTransportStop struct {
	StopID      string  `json:"stopId"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`     // Bus, Tram, Metro, Train
	Distance    float64 `json:"distance"` // meters
	Coordinates struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coordinates"`
	Lines []string `json:"lines"`
}

// Connection represents a PT connection
type Connection struct {
	Line      string `json:"line"`
	Direction string `json:"direction"`
	Departure string `json:"departure"`
	Delay     int    `json:"delay"` // minutes
	Platform  string `json:"platform"`
}

// overpassTransportResponse for OSM public transport query
type overpassTransportResponse struct {
	Elements []struct {
		Type string  `json:"type"`
		ID   int64   `json:"id"`
		Lat  float64 `json:"lat"`
		Lon  float64 `json:"lon"`
		Tags struct {
			Name        string `json:"name"`
			Highway     string `json:"highway"`     // bus_stop
			Railway     string `json:"railway"`     // station, tram_stop, halt
			PublicTransport string `json:"public_transport"` // stop_position, platform
			Network     string `json:"network"`
			Operator    string `json:"operator"`
			Ref         string `json:"ref"`
		} `json:"tags"`
	} `json:"elements"`
}

// FetchOpenOVData retrieves public transport data using OSM Overpass API
// Documentation: https://wiki.openstreetmap.org/wiki/Overpass_API
func (c *ApiClient) FetchOpenOVData(cfg *config.Config, lat, lon float64) (*OpenOVTransportData, error) {
	// Use Overpass API to find PT stops (free, no auth required)
	overpassURL := "https://overpass-api.de/api/interpreter"

	// Query for public transport stops within 1km radius
	radius := 1000
	query := fmt.Sprintf(`[out:json][timeout:10];
(
  node["highway"="bus_stop"](around:%d,%.6f,%.6f);
  node["railway"="tram_stop"](around:%d,%.6f,%.6f);
  node["railway"="station"](around:%d,%.6f,%.6f);
  node["railway"="halt"](around:%d,%.6f,%.6f);
  node["public_transport"="stop_position"](around:%d,%.6f,%.6f);
);
out body qt 30;`, radius, lat, lon, radius, lat, lon, radius, lat, lon, radius, lat, lon, radius, lat, lon)

	logutil.Debugf("[OpenOV] Querying Overpass API for PT stops near %.6f, %.6f", lat, lon)

	req, err := http.NewRequest("POST", overpassURL, nil)
	if err != nil {
		logutil.Debugf("[OpenOV] Request error: %v", err)
		return emptyTransportData(), nil
	}
	
	q := req.URL.Query()
	q.Add("data", query)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[OpenOV] HTTP error: %v", err)
		return emptyTransportData(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logutil.Debugf("[OpenOV] Non-200 status: %d", resp.StatusCode)
		return emptyTransportData(), nil
	}

	var apiResp overpassTransportResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[OpenOV] Decode error: %v", err)
		return emptyTransportData(), nil
	}

	logutil.Debugf("[OpenOV] Found %d PT stops", len(apiResp.Elements))

	stops := make([]PublicTransportStop, 0, len(apiResp.Elements))
	for _, elem := range apiResp.Elements {
		distance := haversineDistanceTraffic(lat, lon, elem.Lat, elem.Lon)
		
		stopType := determineStopType(elem.Tags.Highway, elem.Tags.Railway, elem.Tags.PublicTransport)
		name := elem.Tags.Name
		if name == "" {
			name = fmt.Sprintf("%s stop", stopType)
		}

		stop := PublicTransportStop{
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

	result := &OpenOVTransportData{
		NearestStops: stops,
		Connections:  []Connection{}, // Would need real-time API
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

func sortStopsByDistance(stops []PublicTransportStop) {
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

func emptyTransportData() *OpenOVTransportData {
	return &OpenOVTransportData{
		NearestStops: []PublicTransportStop{},
		Connections:  []Connection{},
	}
}

// ParkingData represents parking availability data
type ParkingData struct {
	TotalSpaces     int           `json:"totalSpaces"`
	AvailableSpaces int           `json:"availableSpaces"`
	OccupancyRate   float64       `json:"occupancyRate"` // percentage
	ParkingZones    []ParkingZone `json:"parkingZones"`
	LastUpdated     string        `json:"lastUpdated"`
}

// ParkingZone represents a parking area
type ParkingZone struct {
	ZoneID      string  `json:"zoneId"`
	Name        string  `json:"name"`
	Type        string  `json:"type"` // Street, Garage, Private
	Capacity    int     `json:"capacity"`
	Available   int     `json:"available"`
	Distance    float64 `json:"distance"`   // meters
	HourlyRate  float64 `json:"hourlyRate"` // EUR
	Coordinates struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coordinates"`
}

// FetchParkingData retrieves parking availability for convenience scoring
// Documentation: Municipal API (varies by city)
func (c *ApiClient) FetchParkingData(cfg *config.Config, lat, lon float64, radius int) (*ParkingData, error) {
	// Return empty data if not configured
	if cfg.ParkingApiURL == "" {
		return &ParkingData{
			TotalSpaces:     0,
			AvailableSpaces: 0,
			ParkingZones:    []ParkingZone{},
		}, nil
	}

	url := fmt.Sprintf("%s/parking?lat=%f&lon=%f&radius=%d", cfg.ParkingApiURL, lat, lon, radius)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &ParkingData{
			TotalSpaces:     0,
			AvailableSpaces: 0,
			ParkingZones:    []ParkingZone{},
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return &ParkingData{
			TotalSpaces:     0,
			AvailableSpaces: 0,
			ParkingZones:    []ParkingZone{},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Return empty data for any non-200 status (including 404)
		return &ParkingData{
			TotalSpaces:     0,
			AvailableSpaces: 0,
			ParkingZones:    []ParkingZone{},
		}, nil
	}

	var result ParkingData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &ParkingData{
			TotalSpaces:     0,
			AvailableSpaces: 0,
			ParkingZones:    []ParkingZone{},
		}, nil
	}

	return &result, nil
}
