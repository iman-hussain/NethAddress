package apiclient

import (
	"context"
	"fmt"
	"time"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// emptyKNMIWeatherData returns a zeroed KNMIWeatherData struct for soft failures.
func emptyKNMIWeatherData() *models.KNMIWeatherData {
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
	}
}

// FetchKNMIWeatherData retrieves real-time weather and historical data
// Documentation: https://dataplatform.knmi.nl
func (c *ApiClient) FetchKNMIWeatherData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.KNMIWeatherData, error) {
	logutil.Debugf("[Weather] FetchKNMIWeatherData: lat=%.6f, lon=%.6f", lat, lon)
	// Return empty data if not configured
	if cfg.KNMIWeatherApiURL == "" {
		logutil.Debugf("[APIClient] FetchKNMIWeatherData: KNMIWeatherApiURL not configured")
		return emptyKNMIWeatherData(), nil
	}

	url := fmt.Sprintf("%s?latitude=%f&longitude=%f&current_weather=true&hourly=precipitation,relativehumidity_2m,pressure_msl&timezone=Europe/Amsterdam", cfg.KNMIWeatherApiURL, lat, lon)
	logutil.Debugf("[APIClient] FetchKNMIWeatherData: url=%s, lat=%.6f, lon=%.6f", cfg.KNMIWeatherApiURL, lat, lon)

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

	if err := c.GetJSON(ctx, "KNMI Weather", url, nil, &result); err != nil {
		return emptyKNMIWeatherData(), nil
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

// emptyWeerliveWeatherData returns a zeroed WeerliveWeatherData struct for soft failures.
func emptyWeerliveWeatherData() *models.WeerliveWeatherData {
	return &models.WeerliveWeatherData{
		Temperature:   0,
		WeatherDesc:   "",
		WindSpeed:     0,
		WindDirection: "",
		Pressure:      0,
		Humidity:      0,
		Visibility:    0,
		Forecast:      []models.WeerliveForecast{},
	}
}

// FetchWeerliveWeather retrieves current weather and 5-day forecast
// Documentation: https://weerlive.nl/delen.php
func (c *ApiClient) FetchWeerliveWeather(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.WeerliveWeatherData, error) {
	// Return empty data if not configured
	if cfg.WeerliveApiURL == "" {
		return emptyWeerliveWeatherData(), nil
	}

	url := fmt.Sprintf("%s?key=%s&locatie=%f,%f", cfg.WeerliveApiURL, cfg.WeerliveApiKey, lat, lon)

	var response struct {
		LiveWeather []models.WeerliveWeatherData `json:"liveweer"`
	}

	if err := c.GetJSON(ctx, "Weerlive", url, nil, &response); err != nil {
		return emptyWeerliveWeatherData(), nil
	}

	if len(response.LiveWeather) == 0 {
		return emptyWeerliveWeatherData(), nil
	}

	return &response.LiveWeather[0], nil
}

// emptyKNMISolarData returns a zeroed KNMISolarData struct for soft failures.
func emptyKNMISolarData() *models.KNMISolarData {
	return &models.KNMISolarData{
		SolarRadiation: 0,
		SunshineHours:  0,
		UVIndex:        0,
		Historical:     []models.HistoricalData{},
	}
}

// FetchKNMISolarData retrieves solar radiation for solar panel potential
// Documentation: https://dataplatform.knmi.nl/group/sunshine-and-radiation
func (c *ApiClient) FetchKNMISolarData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.KNMISolarData, error) {
	// Return empty data if not configured
	if cfg.KNMISolarApiURL == "" {
		logutil.Debugf("[APIClient] FetchKNMISolarData: KNMISolarApiURL not configured")
		return emptyKNMISolarData(), nil
	}

	url := fmt.Sprintf("%s?latitude=%f&longitude=%f&hourly=shortwave_radiation&daily=sunshine_duration,uv_index_max&timezone=Europe/Amsterdam", cfg.KNMISolarApiURL, lat, lon)
	logutil.Debugf("[APIClient] FetchKNMISolarData: url=%s, lat=%.6f, lon=%.6f", cfg.KNMISolarApiURL, lat, lon)

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

	if err := c.GetJSON(ctx, "KNMI Solar", url, nil, &result); err != nil {
		return emptyKNMISolarData(), nil
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
