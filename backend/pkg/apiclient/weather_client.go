	log.Printf("[APIClient] FetchKNMIWeatherData: url=%s, lat=%.6f, lon=%.6f", cfg.KNMIWeatherApiURL, lat, lon)
	log.Printf("[APIClient] FetchKNMIWeatherData: response status=%d", resp.StatusCode)
	log.Printf("[APIClient] FetchKNMIWeatherData: parsed result=%+v", weather)
	log.Printf("[APIClient] FetchKNMISolarData: url=%s, lat=%.6f, lon=%.6f", cfg.KNMISolarApiURL, lat, lon)
	log.Printf("[APIClient] FetchKNMISolarData: response status=%d", resp.StatusCode)
	log.Printf("[APIClient] FetchKNMISolarData: parsed result=%+v", solar)
package apiclient

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

// KNMIWeatherData represents comprehensive weather data from KNMI
type KNMIWeatherData struct {
	Temperature        float64          `json:"temperature"`
	Precipitation      float64          `json:"precipitation"`
	RainfallForecast   []float64        `json:"rainfallForecast"`
	WindSpeed          float64          `json:"windSpeed"`
	WindDirection      int              `json:"windDirection"`
	Humidity           float64          `json:"humidity"`
	Pressure           float64          `json:"pressure"`
	LastUpdated        time.Time        `json:"lastUpdated"`
	HistoricalRainfall []HistoricalData `json:"historicalRainfall"`
}

// HistoricalData represents historical climate data point
type HistoricalData struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// FetchKNMIWeatherData retrieves real-time weather and historical data
// Documentation: https://dataplatform.knmi.nl
func (c *ApiClient) FetchKNMIWeatherData(cfg *config.Config, lat, lon float64) (*KNMIWeatherData, error) {
	// Return empty data if not configured
	if cfg.KNMIWeatherApiURL == "" {
		log.Printf("[WEATHER] KNMIWeatherApiURL not configured")
		return &KNMIWeatherData{
			Temperature:        0,
			Precipitation:      0,
			RainfallForecast:   []float64{},
			WindSpeed:          0,
			WindDirection:      0,
			Humidity:           0,
			Pressure:           0,
			LastUpdated:        time.Time{},
			HistoricalRainfall: []HistoricalData{},
		}, nil
	}

	url := fmt.Sprintf("%s?latitude=%f&longitude=%f&current_weather=true&hourly=precipitation,relativehumidity_2m,pressure_msl&timezone=Europe/Amsterdam", cfg.KNMIWeatherApiURL, lat, lon)
	log.Printf("[WEATHER] Fetching weather from: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[WEATHER] Failed to create request: %v", err)
		return &KNMIWeatherData{
			Temperature:        0,
			Precipitation:      0,
			RainfallForecast:   []float64{},
			WindSpeed:          0,
			WindDirection:      0,
			Humidity:           0,
			Pressure:           0,
			LastUpdated:        time.Time{},
			HistoricalRainfall: []HistoricalData{},
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		log.Printf("[WEATHER] HTTP request failed: %v", err)
		return &KNMIWeatherData{
			Temperature:        0,
			Precipitation:      0,
			RainfallForecast:   []float64{},
			WindSpeed:          0,
			WindDirection:      0,
			Humidity:           0,
			Pressure:           0,
			LastUpdated:        time.Time{},
			HistoricalRainfall: []HistoricalData{},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[WEATHER] API returned status %d", resp.StatusCode)
		return &KNMIWeatherData{
			Temperature:        0,
			Precipitation:      0,
			RainfallForecast:   []float64{},
			WindSpeed:          0,
			WindDirection:      0,
			Humidity:           0,
			Pressure:           0,
			LastUpdated:        time.Time{},
			HistoricalRainfall: []HistoricalData{},
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
		return &KNMIWeatherData{
			Temperature:        0,
			Precipitation:      0,
			RainfallForecast:   []float64{},
			WindSpeed:          0,
			WindDirection:      0,
			Humidity:           0,
			Pressure:           0,
			LastUpdated:        time.Time{},
			HistoricalRainfall: []HistoricalData{},
		}, nil
	}

	weather := &KNMIWeatherData{
		Temperature:        result.CurrentWeather.Temperature,
		WindSpeed:          result.CurrentWeather.WindSpeed,
		WindDirection:      int(result.CurrentWeather.WindDirection),
		RainfallForecast:   make([]float64, 0, 6),
		HistoricalRainfall: make([]HistoricalData, 0, 6),
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
		weather.HistoricalRainfall = append(weather.HistoricalRainfall, HistoricalData{
			Date:  result.Hourly.Time[i],
			Value: result.Hourly.Precipitation[i],
		})
	}

	return weather, nil
}

// WeerliveWeatherData represents weather data from Weerlive API
type WeerliveWeatherData struct {
	Temperature   float64            `json:"temp"`
	WeatherDesc   string             `json:"samenv"`
	WindSpeed     float64            `json:"windsnelheid"`
	WindDirection string             `json:"windrichting"`
	Pressure      float64            `json:"luchtdruk"`
	Humidity      float64            `json:"lv"`
	Visibility    float64            `json:"zicht"`
	Forecast      []WeerliveForecast `json:"verwachting"`
}

// WeerliveForecast represents daily forecast
type WeerliveForecast struct {
	Day           string  `json:"dag"`
	MinTemp       float64 `json:"mintemp"`
	MaxTemp       float64 `json:"maxtemp"`
	Precipitation float64 `json:"neerslag"`
	WindForce     float64 `json:"windkracht"`
}

// FetchWeerliveWeather retrieves current weather and 5-day forecast
// Documentation: https://weerlive.nl/delen.php
func (c *ApiClient) FetchWeerliveWeather(cfg *config.Config, lat, lon float64) (*WeerliveWeatherData, error) {
	// Return empty data if not configured
	if cfg.WeerliveApiURL == "" {
		return &WeerliveWeatherData{
			Temperature:   0,
			WeatherDesc:   "",
			WindSpeed:     0,
			WindDirection: "",
			Pressure:      0,
			Humidity:      0,
			Visibility:    0,
			Forecast:      []WeerliveForecast{},
		}, nil
	}

	url := fmt.Sprintf("%s?key=%s&locatie=%f,%f", cfg.WeerliveApiURL, cfg.WeerliveApiKey, lat, lon)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &WeerliveWeatherData{
			Temperature:   0,
			WeatherDesc:   "",
			WindSpeed:     0,
			WindDirection: "",
			Pressure:      0,
			Humidity:      0,
			Visibility:    0,
			Forecast:      []WeerliveForecast{},
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return &WeerliveWeatherData{
			Temperature:   0,
			WeatherDesc:   "",
			WindSpeed:     0,
			WindDirection: "",
			Pressure:      0,
			Humidity:      0,
			Visibility:    0,
			Forecast:      []WeerliveForecast{},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &WeerliveWeatherData{
			Temperature:   0,
			WeatherDesc:   "",
			WindSpeed:     0,
			WindDirection: "",
			Pressure:      0,
			Humidity:      0,
			Visibility:    0,
			Forecast:      []WeerliveForecast{},
		}, nil
	}

	var response struct {
		LiveWeather []WeerliveWeatherData `json:"liveweer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return &WeerliveWeatherData{
			Temperature:   0,
			WeatherDesc:   "",
			WindSpeed:     0,
			WindDirection: "",
			Pressure:      0,
			Humidity:      0,
			Visibility:    0,
			Forecast:      []WeerliveForecast{},
		}, nil
	}

	if len(response.LiveWeather) == 0 {
		return &WeerliveWeatherData{
			Temperature:   0,
			WeatherDesc:   "",
			WindSpeed:     0,
			WindDirection: "",
			Pressure:      0,
			Humidity:      0,
			Visibility:    0,
			Forecast:      []WeerliveForecast{},
		}, nil
	}

	return &response.LiveWeather[0], nil
}

// KNMISolarData represents solar radiation data for ESG scoring
type KNMISolarData struct {
	SolarRadiation float64          `json:"solarRadiation"` // W/m²
	SunshineHours  float64          `json:"sunshineHours"`
	UVIndex        float64          `json:"uvIndex"`
	Date           string           `json:"date"`
	Historical     []HistoricalData `json:"historical"`
}

// FetchKNMISolarData retrieves solar radiation for solar panel potential
// Documentation: https://dataplatform.knmi.nl/group/sunshine-and-radiation
func (c *ApiClient) FetchKNMISolarData(cfg *config.Config, lat, lon float64) (*KNMISolarData, error) {
	// Return empty data if not configured
	if cfg.KNMISolarApiURL == "" {
		log.Printf("[SOLAR] KNMISolarApiURL not configured")
		return &KNMISolarData{
			SolarRadiation: 0,
			SunshineHours:  0,
			UVIndex:        0,
			Date:           "",
			Historical:     []HistoricalData{},
		}, nil
	}

	url := fmt.Sprintf("%s?latitude=%f&longitude=%f&hourly=shortwave_radiation&daily=sunshine_duration,uv_index_max&timezone=Europe/Amsterdam", cfg.KNMISolarApiURL, lat, lon)
	log.Printf("[SOLAR] Fetching solar data from: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[SOLAR] Failed to create request: %v", err)
		return &KNMISolarData{
			SolarRadiation: 0,
			SunshineHours:  0,
			UVIndex:        0,
			Date:           "",
			Historical:     []HistoricalData{},
		}, nil
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		log.Printf("[SOLAR] HTTP request failed: %v", err)
		return &KNMISolarData{
			SolarRadiation: 0,
			SunshineHours:  0,
			UVIndex:        0,
			Date:           "",
			Historical:     []HistoricalData{},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[SOLAR] API returned status %d", resp.StatusCode)
		return &KNMISolarData{
			SolarRadiation: 0,
			SunshineHours:  0,
			UVIndex:        0,
			Date:           "",
			Historical:     []HistoricalData{},
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
		return &KNMISolarData{
			SolarRadiation: 0,
			SunshineHours:  0,
			UVIndex:        0,
			Date:           "",
			Historical:     []HistoricalData{},
		}, nil
	}

	solar := &KNMISolarData{
		Historical: make([]HistoricalData, 0, 6),
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
	if len(result.Hourly.Time) > 0 {
		solar.Date = result.Hourly.Time[0]
	}

	for i := 0; i < len(result.Hourly.Time) && i < len(result.Hourly.ShortwaveRadiation); i++ {
		if i >= 6 {
			break
		}
		solar.Historical = append(solar.Historical, HistoricalData{
			Date:  result.Hourly.Time[i],
			Value: result.Hourly.ShortwaveRadiation[i],
		})
	}

	return solar, nil
}
