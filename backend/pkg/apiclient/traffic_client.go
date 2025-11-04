package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
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
	if cfg.NDWTrafficApiURL == "" {
		// Return empty data when API is not configured (requires key)
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

// FetchOpenOVData retrieves public transport data for accessibility scoring
// Documentation: https://openov.nl
func (c *ApiClient) FetchOpenOVData(cfg *config.Config, lat, lon float64) (*OpenOVTransportData, error) {
	// Return empty data if not configured
	if cfg.OpenOVApiURL == "" {
		return &OpenOVTransportData{
			NearestStops: []PublicTransportStop{},
			Connections:  []Connection{},
		}, nil
	}

	url := fmt.Sprintf("%s/stops?lat=%f&lon=%f&radius=1000", cfg.OpenOVApiURL, lat, lon)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &OpenOVTransportData{
			NearestStops: []PublicTransportStop{},
			Connections:  []Connection{},
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return &OpenOVTransportData{
			NearestStops: []PublicTransportStop{},
			Connections:  []Connection{},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &OpenOVTransportData{
			NearestStops: []PublicTransportStop{},
			Connections:  []Connection{},
		}, nil
	}

	var result OpenOVTransportData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &OpenOVTransportData{
			NearestStops: []PublicTransportStop{},
			Connections:  []Connection{},
		}, nil
	}

	return &result, nil
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
	if cfg.ParkingApiURL == "" {
		return nil, fmt.Errorf("ParkingApiURL not configured")
	}

	url := fmt.Sprintf("%s/parking?lat=%f&lon=%f&radius=%d", cfg.ParkingApiURL, lat, lon, radius)
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
		// No parking data available
		return &ParkingData{
			TotalSpaces:     0,
			AvailableSpaces: 0,
			ParkingZones:    []ParkingZone{},
		}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("parking API returned status %d", resp.StatusCode)
	}

	var result ParkingData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode parking response: %w", err)
	}

	return &result, nil
}
