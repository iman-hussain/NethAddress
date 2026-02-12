package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
	"github.com/iman-hussain/nethaddress/backend/pkg/logutil"
	"github.com/iman-hussain/nethaddress/backend/pkg/models"
)

// emptyAirQualityData returns a default AirQualityData struct for soft failures.
func emptyAirQualityData() *models.AirQualityData {
	return &models.AirQualityData{
		StationID:    "",
		StationName:  "",
		Measurements: []models.AirMeasurement{},
		AQI:          0,
		Category:     "Unknown",
		LastUpdated:  "",
	}
}

// FetchAirQualityData retrieves real-time air quality data
// Documentation: https://api-docs.luchtmeetnet.nl
// Note: This function makes 2 sequential API calls (find station, then get measurements)
// and is intentionally not refactored to use GetJSON due to intermediate state handling.
func (c *ApiClient) FetchAirQualityData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.AirQualityData, error) {
	logutil.Debugf("[APIClient] FetchAirQualityData: url=%s, lat=%.6f, lon=%.6f", cfg.LuchtmeetnetApiURL, lat, lon)

	// Return empty data if not configured
	if cfg.LuchtmeetnetApiURL == "" {
		return emptyAirQualityData(), nil
	}

	// Find nearest station
	stationURL := fmt.Sprintf("%s/stations?lat=%f&lon=%f&limit=1", cfg.LuchtmeetnetApiURL, lat, lon)
	logutil.Debugf("[APIClient] FetchAirQualityData: stationURL=%s", stationURL)
	req, err := http.NewRequestWithContext(ctx, "GET", stationURL, nil)
	if err != nil {
		return emptyAirQualityData(), nil
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return emptyAirQualityData(), nil
	}
	defer resp.Body.Close()

	logutil.Debugf("[APIClient] FetchAirQualityData: response status=%d", resp.StatusCode)

	if resp.StatusCode != 200 {
		return emptyAirQualityData(), nil
	}

	var stations struct {
		Data []struct {
			Number   string `json:"number"`
			Location string `json:"location"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return emptyAirQualityData(), nil
	}

	if len(stations.Data) == 0 {
		logutil.Debugf("[APIClient] FetchAirQualityData: no stations found")
		return emptyAirQualityData(), nil
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

// emptyNoisePollutionData returns a default NoisePollutionData struct for soft failures.
func emptyNoisePollutionData() *models.NoisePollutionData {
	return &models.NoisePollutionData{
		TotalNoise:    0,
		RoadNoise:     0,
		RailNoise:     0,
		IndustryNoise: 0,
		AircraftNoise: 0,
		NoiseCategory: "Unknown",
		ExceedsLimit:  false,
		Sources:       []models.NoiseSource{},
	}
}

// FetchNoisePollutionData retrieves noise pollution data for livability scoring
// Documentation: Government noise API
func (c *ApiClient) FetchNoisePollutionData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.NoisePollutionData, error) {
	// Return default data if not configured
	if cfg.NoisePollutionApiURL == "" {
		return emptyNoisePollutionData(), nil
	}

	url := fmt.Sprintf("%s/noise?lat=%f&lon=%f", cfg.NoisePollutionApiURL, lat, lon)

	var result models.NoisePollutionData
	if err := c.GetJSON(ctx, "NoisePollution", url, nil, &result); err != nil {
		// Return default data for failures (soft failure)
		return emptyNoisePollutionData(), nil
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
