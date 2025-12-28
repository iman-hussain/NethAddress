package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// FetchAirQualityData retrieves real-time air quality data
// Documentation: https://api-docs.luchtmeetnet.nl
func (c *ApiClient) FetchAirQualityData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.AirQualityData, error) {
	logutil.Debugf("[APIClient] FetchAirQualityData: url=%s, lat=%.6f, lon=%.6f", cfg.LuchtmeetnetApiURL, lat, lon)

	// Return empty data if not configured
	if cfg.LuchtmeetnetApiURL == "" {
		return &models.AirQualityData{
			StationID:    "",
			StationName:  "",
			Measurements: []models.AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}

	// Find nearest station
	stationURL := fmt.Sprintf("%s/stations?lat=%f&lon=%f&limit=1", cfg.LuchtmeetnetApiURL, lat, lon)
	logutil.Debugf("[APIClient] FetchAirQualityData: stationURL=%s", stationURL)
	req, err := http.NewRequestWithContext(ctx, "GET", stationURL, nil)
	if err != nil {
		return &models.AirQualityData{
			StationID:    "",
			StationName:  "",
			Measurements: []models.AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return &models.AirQualityData{
			StationID:    "",
			StationName:  "",
			Measurements: []models.AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}
	defer resp.Body.Close()

	logutil.Debugf("[APIClient] FetchAirQualityData: response status=%d", resp.StatusCode)

	if resp.StatusCode != 200 {
		return &models.AirQualityData{
			StationID:    "",
			StationName:  "",
			Measurements: []models.AirMeasurement{},
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
		return &models.AirQualityData{
			StationID:    "",
			StationName:  "",
			Measurements: []models.AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}

	if len(stations.Data) == 0 {
		logutil.Debugf("[APIClient] FetchAirQualityData: measure request error: %v", err)
		return &models.AirQualityData{
			StationID:    "",
			StationName:  "",
			Measurements: []models.AirMeasurement{},
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
	req2, err := http.NewRequestWithContext(ctx, "GET", measureURL, nil)
	if err != nil {
		return &models.AirQualityData{
			StationID:    stationID,
			StationName:  stationName,
			Measurements: []models.AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}
	req2.Header.Set("Accept", "application/json")

	resp2, err := c.HTTP.Do(req2)
	if err != nil {
		return &models.AirQualityData{
			StationID:    stationID,
			StationName:  stationName,
			Measurements: []models.AirMeasurement{},
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
		return &models.AirQualityData{
			StationID:    stationID,
			StationName:  stationName,
			Measurements: []models.AirMeasurement{},
			AQI:          0,
			Category:     "Unknown",
			LastUpdated:  "",
		}, nil
	}

	airMeasurements := make([]models.AirMeasurement, 0, len(measurements.Data))
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

		airMeasurements = append(airMeasurements, models.AirMeasurement{
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

	return &models.AirQualityData{
		StationID:    stationID,
		StationName:  stationName,
		Measurements: airMeasurements,
		AQI:          aqi,
		Category:     category,
		LastUpdated:  lastUpdated,
	}, nil
}

// FetchNoisePollutionData retrieves noise pollution data for livability scoring
// Documentation: Government noise API
func (c *ApiClient) FetchNoisePollutionData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.NoisePollutionData, error) {
	// Return default data if not configured
	if cfg.NoisePollutionApiURL == "" {
		return &models.NoisePollutionData{
			TotalNoise:    0,
			RoadNoise:     0,
			RailNoise:     0,
			IndustryNoise: 0,
			AircraftNoise: 0,
			NoiseCategory: "Unknown",
			ExceedsLimit:  false,
			Sources:       []models.NoiseSource{},
		}, nil
	}

	url := fmt.Sprintf("%s/noise?lat=%f&lon=%f", cfg.NoisePollutionApiURL, lat, lon)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &models.NoisePollutionData{
			TotalNoise:    0,
			RoadNoise:     0,
			RailNoise:     0,
			IndustryNoise: 0,
			AircraftNoise: 0,
			NoiseCategory: "Unknown",
			ExceedsLimit:  false,
			Sources:       []models.NoiseSource{},
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return &models.NoisePollutionData{
			TotalNoise:    0,
			RoadNoise:     0,
			RailNoise:     0,
			IndustryNoise: 0,
			AircraftNoise: 0,
			NoiseCategory: "Unknown",
			ExceedsLimit:  false,
			Sources:       []models.NoiseSource{},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// No noise data - assume quiet area
		return &models.NoisePollutionData{
			TotalNoise:    45.0,
			NoiseCategory: "Quiet",
			ExceedsLimit:  false,
			Sources:       []models.NoiseSource{},
		}, nil
	}

	if resp.StatusCode != 200 {
		return &models.NoisePollutionData{
			TotalNoise:    0,
			RoadNoise:     0,
			RailNoise:     0,
			IndustryNoise: 0,
			AircraftNoise: 0,
			NoiseCategory: "Unknown",
			ExceedsLimit:  false,
			Sources:       []models.NoiseSource{},
		}, nil
	}

	var result models.NoisePollutionData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &models.NoisePollutionData{
			TotalNoise:    0,
			RoadNoise:     0,
			RailNoise:     0,
			IndustryNoise: 0,
			AircraftNoise: 0,
			NoiseCategory: "Unknown",
			ExceedsLimit:  false,
			Sources:       []models.NoiseSource{},
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
