package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
)

// AirQualityData represents comprehensive air quality measurements
type AirQualityData struct {
	StationID    string           `json:"stationId"`
	StationName  string           `json:"stationName"`
	Measurements []AirMeasurement `json:"measurements"`
	AQI          int              `json:"aqi"`      // Air Quality Index 0-500
	Category     string           `json:"category"` // Good, Moderate, Unhealthy, etc.
	LastUpdated  string           `json:"lastUpdated"`
}

// AirMeasurement represents a single pollutant measurement
type AirMeasurement struct {
	Parameter string  `json:"parameter"` // NO2, PM10, PM2.5, O3, etc.
	Value     float64 `json:"value"`     // µg/m³
	Unit      string  `json:"unit"`
}

// FetchAirQualityData retrieves real-time air quality data
// Documentation: https://api-docs.luchtmeetnet.nl
func (c *ApiClient) FetchAirQualityData(cfg *config.Config, lat, lon float64) (*AirQualityData, error) {
	logutil.Debugf("[APIClient] FetchAirQualityData: url=%s, lat=%.6f, lon=%.6f", cfg.LuchtmeetnetApiURL, lat, lon)

	// Return empty data if not configured
	if cfg.LuchtmeetnetApiURL == "" {
		return &AirQualityData{
			StationID:    "",
			StationName:  "",
			Measurements: []AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}

	// Find nearest station
	stationURL := fmt.Sprintf("%s/stations?lat=%f&lon=%f&limit=1", cfg.LuchtmeetnetApiURL, lat, lon)
	logutil.Debugf("[APIClient] FetchAirQualityData: stationURL=%s", stationURL)
	req, err := http.NewRequest("GET", stationURL, nil)
	if err != nil {
		return &AirQualityData{
			StationID:    "",
			StationName:  "",
			Measurements: []AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return &AirQualityData{
			StationID:    "",
			StationName:  "",
			Measurements: []AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}
	defer resp.Body.Close()

	logutil.Debugf("[APIClient] FetchAirQualityData: response status=%d", resp.StatusCode)

	if resp.StatusCode != 200 {
		return &AirQualityData{
			StationID:    "",
			StationName:  "",
			Measurements: []AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}

	var stations struct {
		Data []struct {
			Number   string `json:"number"`
			Location string `json:"location"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return &AirQualityData{
			StationID:    "",
			StationName:  "",
			Measurements: []AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}

	if len(stations.Data) == 0 {
		logutil.Debugf("[APIClient] FetchAirQualityData: measure request error: %v", err)
		return &AirQualityData{
			StationID:    "",
			StationName:  "",
			Measurements: []AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}

	stationID := stations.Data[0].Number
	stationName := stations.Data[0].Location

	// Get measurements for this station
	measureURL := fmt.Sprintf("%s/stations/%s/measurements?order_by=timestamp_measured&order_direction=desc&limit=25", cfg.LuchtmeetnetApiURL, stationID)
	logutil.Debugf("[APIClient] FetchAirQualityData: measureURL=%s", measureURL)
	req2, err := http.NewRequest("GET", measureURL, nil)
	if err != nil {
		return &AirQualityData{
			StationID:    stationID,
			StationName:  stationName,
			Measurements: []AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}
	req2.Header.Set("Accept", "application/json")

	resp2, err := c.HTTP.Do(req2)
	if err != nil {
		return &AirQualityData{
			StationID:    stationID,
			StationName:  stationName,
			Measurements: []AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}
	defer resp2.Body.Close()

	var measurements struct {
		Data []struct {
			Formula           string  `json:"formula"`
			Value             float64 `json:"value"`
			TimestampMeasured string  `json:"timestamp_measured"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp2.Body).Decode(&measurements); err != nil {
		return &AirQualityData{
			StationID:    stationID,
			StationName:  stationName,
			Measurements: []AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}

	airMeasurements := make([]AirMeasurement, 0, len(measurements.Data))
	lastUpdated := ""

	// Calculate simplified AQI based on PM2.5 (if available)
	aqi := 50 // Default "Good"
	category := "Good"

	for _, m := range measurements.Data {
		if lastUpdated == "" {
			lastUpdated = m.TimestampMeasured
		}

		unit := "µg/m³"
		parameter := m.Formula
		switch strings.ToUpper(m.Formula) {
		case "NO2", "NO", "O3", "PM10", "PM25":
			unit = "µg/m³"
		case "CO":
			unit = "µg/m³"
		}

		airMeasurements = append(airMeasurements, AirMeasurement{
			Parameter: parameter,
			Value:     m.Value,
			Unit:      unit,
		})

		if strings.EqualFold(m.Formula, "PM25") && m.Value > 0 {
			if m.Value <= 12 {
				aqi = int((m.Value / 12.0) * 50)
				category = "Good"
			} else if m.Value <= 35.4 {
				aqi = 51 + int(((m.Value-12.1)/(35.4-12.1))*49)
				category = "Moderate"
			} else {
				aqi = 101
				category = "Unhealthy for Sensitive Groups"
			}
		}
	}

	return &AirQualityData{
		StationID:    stationID,
		StationName:  stationName,
		Measurements: airMeasurements,
		AQI:          aqi,
		Category:     category,
		LastUpdated:  lastUpdated,
	}, nil
}

// NoisePollutionData represents noise levels from various sources
type NoisePollutionData struct {
	TotalNoise    float64       `json:"totalNoise"`    // dB(A)
	RoadNoise     float64       `json:"roadNoise"`     // dB(A)
	RailNoise     float64       `json:"railNoise"`     // dB(A)
	IndustryNoise float64       `json:"industryNoise"` // dB(A)
	AircraftNoise float64       `json:"aircraftNoise"` // dB(A)
	NoiseCategory string        `json:"noiseCategory"` // Quiet, Moderate, Loud, Very Loud
	ExceedsLimit  bool          `json:"exceedsLimit"`  // Above 55 dB(A) limit
	Sources       []NoiseSource `json:"sources"`
}

// NoiseSource represents a specific noise contributor
type NoiseSource struct {
	Type       string  `json:"type"`
	Distance   float64 `json:"distance"`   // meters
	NoiseLevel float64 `json:"noiseLevel"` // dB(A)
}

// FetchNoisePollutionData retrieves noise pollution data for livability scoring
// Documentation: Government noise API
func (c *ApiClient) FetchNoisePollutionData(cfg *config.Config, lat, lon float64) (*NoisePollutionData, error) {
	// Return default data if not configured
	if cfg.NoisePollutionApiURL == "" {
		return &NoisePollutionData{
			TotalNoise:    0,
			RoadNoise:     0,
			RailNoise:     0,
			IndustryNoise: 0,
			AircraftNoise: 0,
			NoiseCategory: "Unknown",
			ExceedsLimit:  false,
			Sources:       []NoiseSource{},
		}, nil
	}

	url := fmt.Sprintf("%s/noise?lat=%f&lon=%f", cfg.NoisePollutionApiURL, lat, lon)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &NoisePollutionData{
			TotalNoise:    0,
			RoadNoise:     0,
			RailNoise:     0,
			IndustryNoise: 0,
			AircraftNoise: 0,
			NoiseCategory: "Unknown",
			ExceedsLimit:  false,
			Sources:       []NoiseSource{},
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return &NoisePollutionData{
			TotalNoise:    0,
			RoadNoise:     0,
			RailNoise:     0,
			IndustryNoise: 0,
			AircraftNoise: 0,
			NoiseCategory: "Unknown",
			ExceedsLimit:  false,
			Sources:       []NoiseSource{},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// No noise data - assume quiet area
		return &NoisePollutionData{
			TotalNoise:    45.0,
			NoiseCategory: "Quiet",
			ExceedsLimit:  false,
			Sources:       []NoiseSource{},
		}, nil
	}

	if resp.StatusCode != 200 {
		return &NoisePollutionData{
			TotalNoise:    0,
			RoadNoise:     0,
			RailNoise:     0,
			IndustryNoise: 0,
			AircraftNoise: 0,
			NoiseCategory: "Unknown",
			ExceedsLimit:  false,
			Sources:       []NoiseSource{},
		}, nil
	}

	var result NoisePollutionData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &NoisePollutionData{
			TotalNoise:    0,
			RoadNoise:     0,
			RailNoise:     0,
			IndustryNoise: 0,
			AircraftNoise: 0,
			NoiseCategory: "Unknown",
			ExceedsLimit:  false,
			Sources:       []NoiseSource{},
		}, nil
	}

	// Categorize noise level
	if result.TotalNoise < 50 {
		result.NoiseCategory = "Quiet"
	} else if result.TotalNoise < 55 {
		result.NoiseCategory = "Moderate"
	} else if result.TotalNoise < 65 {
		result.NoiseCategory = "Loud"
	} else {
		result.NoiseCategory = "Very Loud"
	}

	result.ExceedsLimit = result.TotalNoise > 55.0

	return &result, nil
}
