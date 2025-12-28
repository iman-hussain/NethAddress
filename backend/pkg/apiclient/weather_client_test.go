package apiclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchKNMIWeatherData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"current_weather": map[string]any{
				"temperature":   15.5,
				"windspeed":     12.5,
				"winddirection": 180.0,
				"time":          "2024-11-01T10:00:00Z",
			},
			"hourly": map[string]any{
				"time":                []string{"2024-11-01T10:00:00Z", "2024-11-01T11:00:00Z"},
				"precipitation":       []float64{2.3, 1.1},
				"relativehumidity_2m": []float64{75.0, 74.0},
				"pressure_msl":        []float64{1013.2, 1012.8},
			},
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		KNMIWeatherApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	result, err := client.FetchKNMIWeatherData(context.Background(), cfg, 52.37, 4.89)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Temperature != 15.5 {
		t.Errorf("Expected temperature 15.5, got %f", result.Temperature)
	}
	if result.WindSpeed != 12.5 {
		t.Errorf("Expected wind speed 12.5, got %f", result.WindSpeed)
	}
}

func TestFetchWeerliveWeather(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"liveweer": [{
				"temp": 14.2,
				"samenv": "Bewolkt",
				"windsnelheid": 10.5,
				"windrichting": "ZW",
				"luchtdruk": 1015.3,
				"lv": 80,
				"zicht": 10.0,
				"verwachting": [
					{"dag": "Maandag", "mintemp": 10, "maxtemp": 16, "neerslag": 5, "windkracht": 4}
				]
			}]
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		WeerliveApiURL: server.URL,
		WeerliveApiKey: "test-key",
	}
	client := NewApiClient(server.Client(), cfg)

	result, err := client.FetchWeerliveWeather(context.Background(), cfg, 52.37, 4.89)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Temperature != 14.2 {
		t.Errorf("Expected temperature 14.2, got %f", result.Temperature)
	}
	if result.WeatherDesc != "Bewolkt" {
		t.Errorf("Expected weather 'Bewolkt', got %s", result.WeatherDesc)
	}
}

func TestFetchKNMISolarData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"hourly": map[string]any{
				"time":                []string{"2024-11-01T10:00:00Z", "2024-11-01T11:00:00Z"},
				"shortwave_radiation": []float64{450.5, 420.0},
			},
			"daily": map[string]any{
				"time":              []string{"2024-11-01"},
				"sunshine_duration": []float64{22320},
				"uv_index_max":      []float64{3.5},
			},
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		KNMISolarApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	result, err := client.FetchKNMISolarData(context.Background(), cfg, 52.37, 4.89)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.SolarRadiation != 450.5 {
		t.Errorf("Expected solar radiation 450.5, got %f", result.SolarRadiation)
	}
	if result.SunshineHours != 6.2 {
		t.Errorf("Expected sunshine hours 6.2, got %f", result.SunshineHours)
	}
}
