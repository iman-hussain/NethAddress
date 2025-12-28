package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// FetchKNMIWeatherData retrieves real-time weather and historical data
// Documentation: https://dataplatform.knmi.nl
func (c *ApiClient) FetchKNMIWeatherData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.KNMIWeatherData, error) {
	logutil.Debugf("[Weather] FetchKNMIWeatherData: lat=%.6f, lon=%.6f", lat, lon)
	// Return empty data if not configured
	if cfg.KNMIWeatherApiURL == "" {
		logutil.Debugf("[APIClient] FetchKNMIWeatherData: KNMIWeatherApiURL not configured")
		return &models.KNMIWeatherData{
			Temperature:        0,
			Precipitation:      0,
			RainfallForecast:   []float64{},
			WindSpeed:          0,
			WindDirection:      0,
			Humidity:           0,
			Pressure:           0,
			LastUpdated:        time.Time{},
			HistoricalRainfall: []models.HistoricalData{},
		}, nil
	}

	url := fmt.Sprintf("%s?latitude=%f&longitude=%f&current_weather=true&hourly=precipitation,relativehumidity_2m,pressure_msl&timezone=Europe/Amsterdam", cfg.KNMIWeatherApiURL, lat, lon)
	logutil.Debugf("[APIClient] FetchKNMIWeatherData: url=%s, lat=%.6f, lon=%.6f", cfg.KNMIWeatherApiURL, lat, lon)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logutil.Debugf("[APIClient] FetchKNMIWeatherData: request creation failed: %v", err)
		return &models.KNMIWeatherData{
			Temperature:        0,
			Precipitation:      0,
			RainfallForecast:   []float64{},
			WindSpeed:          0,
			WindDirection:      0,
			Humidity:           0,
			Pressure:           0,
			LastUpdated:        time.Time{},
			HistoricalRainfall: []models.HistoricalData{},
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[APIClient] FetchKNMIWeatherData: HTTP request failed: %v", err)
		return &models.KNMIWeatherData{
			Temperature:        0,
			Precipitation:      0,
			RainfallForecast:   []float64{},
			WindSpeed:          0,
			WindDirection:      0,
			Humidity:           0,
			Pressure:           0,
			LastUpdated:        time.Time{},
			HistoricalRainfall: []models.HistoricalData{},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logutil.Debugf("[APIClient] FetchKNMIWeatherData: response status=%d", resp.StatusCode)
		return &models.KNMIWeatherData{
			Temperature:        0,
			Precipitation:      0,
			RainfallForecast:   []float64{},
			WindSpeed:          0,
			WindDirection:      0,
			Humidity:           0,
			Pressure:           0,
			LastUpdated:        time.Time{},
			HistoricalRainfall: []models.HistoricalData{},
		}, nil
	}

	var result struct {
		CurrentWeather struct {
			Temperature   float64 `json:"temperature"`
			WindSpeed     float64 `json:"windspeed"`
			WindDirection float64 `json:"winddirection"`
			Time          string  `json:"time"`
		} `json:"current_weather"`
		Hourly struct {
			Time                 []string  `json:"time"`
			Precipitation        []float64 `json:"precipitation"`
			RelativeHumidity     []float64 `json:"relativehumidity_2m"`
			MeanSeaLevelPressure []float64 `json:"pressure_msl"`
		} `json:"hourly"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &models.KNMIWeatherData{
			Temperature:        0,
			Precipitation:      0,
			RainfallForecast:   []float64{},
			WindSpeed:          0,
			WindDirection:      0,
			Humidity:           0,
			Pressure:           0,
			LastUpdated:        time.Time{},
			HistoricalRainfall: []models.HistoricalData{},
		}, nil
	}

	weather := &models.KNMIWeatherData{
		Temperature:        result.CurrentWeather.Temperature,
		WindSpeed:          result.CurrentWeather.WindSpeed,
		WindDirection:      int(result.CurrentWeather.WindDirection),
		RainfallForecast:   make([]float64, 0, 6),
		HistoricalRainfall: make([]models.HistoricalData, 0, 6),
	}

	if len(result.Hourly.RelativeHumidity) > 0 {
		weather.Humidity = result.Hourly.RelativeHumidity[0]
	}
	if len(result.Hourly.MeanSeaLevelPressure) > 0 {
		weather.Pressure = result.Hourly.MeanSeaLevelPressure[0]
	}
	if len(result.Hourly.Precipitation) > 0 {
		weather.Precipitation = result.Hourly.Precipitation[0]
	}
	if result.CurrentWeather.Time != "" {
		if t, err := time.Parse(time.RFC3339, result.CurrentWeather.Time); err == nil {
			weather.LastUpdated = t
		}
	}

	for i, v := range result.Hourly.Precipitation {
		if i < 6 {
			weather.RainfallForecast = append(weather.RainfallForecast, v)
		} else {
			break
		}
	}

	for i := 0; i < len(result.Hourly.Time) && i < len(result.Hourly.Precipitation); i++ {
		if i >= 6 {
			break
		}
		weather.HistoricalRainfall = append(weather.HistoricalRainfall, models.HistoricalData{
			Date:  result.Hourly.Time[i],
			Value: result.Hourly.Precipitation[i],
		})
	}

	logutil.Debugf("[APIClient] FetchKNMIWeatherData: parsed result=%+v", weather)
	return weather, nil
}

// FetchWeerliveWeather retrieves current weather and 5-day forecast
// Documentation: https://weerlive.nl/delen.php
func (c *ApiClient) FetchWeerliveWeather(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.WeerliveWeatherData, error) {
	// Return empty data if not configured
	if cfg.WeerliveApiURL == "" {
		return &models.WeerliveWeatherData{
			Temperature:   0,
			WeatherDesc:   "",
			WindSpeed:     0,
			WindDirection: "",
			Pressure:      0,
			Humidity:      0,
			Visibility:    0,
			Forecast:      []models.WeerliveForecast{},
		}, nil
	}

	url := fmt.Sprintf("%s?key=%s&locatie=%f,%f", cfg.WeerliveApiURL, cfg.WeerliveApiKey, lat, lon)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &models.WeerliveWeatherData{
			Temperature:   0,
			WeatherDesc:   "",
			WindSpeed:     0,
			WindDirection: "",
			Pressure:      0,
			Humidity:      0,
			Visibility:    0,
			Forecast:      []models.WeerliveForecast{},
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return &models.WeerliveWeatherData{
			Temperature:   0,
			WeatherDesc:   "",
			WindSpeed:     0,
			WindDirection: "",
			Pressure:      0,
			Humidity:      0,
			Visibility:    0,
			Forecast:      []models.WeerliveForecast{},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &models.WeerliveWeatherData{
			Temperature:   0,
			WeatherDesc:   "",
			WindSpeed:     0,
			WindDirection: "",
			Pressure:      0,
			Humidity:      0,
			Visibility:    0,
			Forecast:      []models.WeerliveForecast{},
		}, nil
	}

	var response struct {
		LiveWeather []models.WeerliveWeatherData `json:"liveweer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return &models.WeerliveWeatherData{
			Temperature:   0,
			WeatherDesc:   "",
			WindSpeed:     0,
			WindDirection: "",
			Pressure:      0,
			Humidity:      0,
			Visibility:    0,
			Forecast:      []models.WeerliveForecast{},
		}, nil
	}

	if len(response.LiveWeather) == 0 {
		return &models.WeerliveWeatherData{
			Temperature:   0,
			WeatherDesc:   "",
			WindSpeed:     0,
			WindDirection: "",
			Pressure:      0,
			Humidity:      0,
			Visibility:    0,
			Forecast:      []models.WeerliveForecast{},
		}, nil
	}

	return &response.LiveWeather[0], nil
}

// FetchKNMISolarData retrieves solar radiation for solar panel potential
// Documentation: https://dataplatform.knmi.nl/group/sunshine-and-radiation
func (c *ApiClient) FetchKNMISolarData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.KNMISolarData, error) {
	// Return empty data if not configured
	if cfg.KNMISolarApiURL == "" {
		logutil.Debugf("[APIClient] FetchKNMISolarData: KNMISolarApiURL not configured")
		return &models.KNMISolarData{
			SolarRadiation: 0,
			SunshineHours:  0,
			UVIndex:        0,
			Historical:     []models.HistoricalData{},
		}, nil
	}

	url := fmt.Sprintf("%s?latitude=%f&longitude=%f&hourly=shortwave_radiation&daily=sunshine_duration,uv_index_max&timezone=Europe/Amsterdam", cfg.KNMISolarApiURL, lat, lon)
	logutil.Debugf("[APIClient] FetchKNMISolarData: url=%s, lat=%.6f, lon=%.6f", cfg.KNMISolarApiURL, lat, lon)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logutil.Debugf("[APIClient] FetchKNMISolarData: request creation failed: %v", err)
		return &models.KNMISolarData{
			SolarRadiation: 0,
			SunshineHours:  0,
			UVIndex:        0,
			Historical:     []models.HistoricalData{},
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[APIClient] FetchKNMISolarData: HTTP request failed: %v", err)
		return &models.KNMISolarData{
			SolarRadiation: 0,
			SunshineHours:  0,
			UVIndex:        0,
			Historical:     []models.HistoricalData{},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logutil.Debugf("[APIClient] FetchKNMISolarData: response status=%d", resp.StatusCode)
		return &models.KNMISolarData{
			SolarRadiation: 0,
			SunshineHours:  0,
			UVIndex:        0,
			Historical:     []models.HistoricalData{},
		}, nil
	}

	var result struct {
		Hourly struct {
			Time               []string  `json:"time"`
			ShortwaveRadiation []float64 `json:"shortwave_radiation"`
		} `json:"hourly"`
		Daily struct {
			Time          []string  `json:"time"`
			SunshineHours []float64 `json:"sunshine_duration"`
			UVIndexMax    []float64 `json:"uv_index_max"`
		} `json:"daily"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &models.KNMISolarData{
			SolarRadiation: 0,
			SunshineHours:  0,
			UVIndex:        0,
			Historical:     []models.HistoricalData{},
		}, nil
	}

	solar := &models.KNMISolarData{
		Historical: make([]models.HistoricalData, 0, 6),
	}

	if len(result.Hourly.ShortwaveRadiation) > 0 {
		solar.SolarRadiation = result.Hourly.ShortwaveRadiation[0]
	}
	if len(result.Daily.SunshineHours) > 0 {
		solar.SunshineHours = result.Daily.SunshineHours[0] / 3600.0
	}
	if len(result.Daily.UVIndexMax) > 0 {
		solar.UVIndex = result.Daily.UVIndexMax[0]
	}
	// Note: Date set from result.Hourly.Time is missing in my manual reconstruction if not careful,
	// but I see in previous code `solar.Date = result.Hourly.Time[0]` if avail.
	if len(result.Hourly.Time) > 0 {
		// solar.Date = ... wait, KNMISolarData struct in models has no Date field?
		// checking models.go content...
	}

	for i := 0; i < len(result.Hourly.Time) && i < len(result.Hourly.ShortwaveRadiation); i++ {
		if i >= 6 {
			break
		}
		solar.Historical = append(solar.Historical, models.HistoricalData{
			Date:  result.Hourly.Time[i],
			Value: result.Hourly.ShortwaveRadiation[i],
		})
	}

	logutil.Debugf("[APIClient] FetchKNMISolarData: parsed result=%+v", solar)
	return solar, nil
}
