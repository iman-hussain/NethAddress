package apiclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchAirQualityData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/stations":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": [{
					"number": "NL123",
					"location": "Utrecht-Griftpark"
				}]
			}`))
		case r.URL.Path == "/stations/NL123/measurements":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": [{
					"formula": "PM25",
					"value": 12.5,
					"timestamp_measured": "2024-01-15T10:00:00Z"
				}]
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		LuchtmeetnetApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	data, err := client.FetchAirQualityData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.AQI != 51 {
		t.Errorf("Expected AQI 51, got %d", data.AQI)
	}
	if data.StationName != "Utrecht-Griftpark" {
		t.Errorf("Expected station 'Utrecht-Griftpark', got '%s'", data.StationName)
	}
	if data.Category != "Moderate" {
		t.Errorf("Expected category 'Moderate', got '%s'", data.Category)
	}
	if len(data.Measurements) != 1 {
		t.Fatalf("Expected 1 measurement, got %d", len(data.Measurements))
	}
	if data.Measurements[0].Parameter != "PM25" {
		t.Errorf("Expected PM25 measurement, got %s", data.Measurements[0].Parameter)
	}
}

func TestFetchNoisePollutionData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"roadNoise": 55.3,
			"railNoise": 42.1,
			"aircraftNoise": 38.5,
			"industryNoise": 45.2,
			"totalNoise": 58.7
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		NoisePollutionApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	data, err := client.FetchNoisePollutionData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.RoadNoise != 55.3 {
		t.Errorf("Expected road noise 55.3, got %f", data.RoadNoise)
	}
	if data.TotalNoise != 58.7 {
		t.Errorf("Expected total noise 58.7, got %f", data.TotalNoise)
	}
	if data.NoiseCategory != "Loud" {
		t.Errorf("Expected noise category 'Loud', got '%s'", data.NoiseCategory)
	}
}

func TestFetchAirQualityData_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.Config{
		LuchtmeetnetApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	data, err := client.FetchAirQualityData(context.Background(), cfg, 0, 0)
	if err != nil {
		t.Errorf("Expected no error with graceful degradation, got %v", err)
	}
	if data.StationID != "" || data.Category != "Unknown" {
		t.Error("Expected empty data with Unknown category for invalid coordinates")
	}
}
