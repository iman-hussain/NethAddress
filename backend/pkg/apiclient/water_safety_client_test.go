package apiclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchFloodRiskData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"riskLevel": "Medium",
			"floodProbability": 0.4,
			"waterDepth": 0.8,
			"nearestDike": 850,
			"dikeQuality": "Good",
			"floodZone": "Zone 3"
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		FloodRiskApiURL: server.URL,
	}
	client := NewApiClient(server.Client())

	data, err := client.FetchFloodRiskData(cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.RiskLevel != "Medium" {
		t.Errorf("Expected risk level 'Medium', got '%s'", data.RiskLevel)
	}
	if data.WaterDepth != 0.8 {
		t.Errorf("Expected water depth 0.8, got %f", data.WaterDepth)
	}
}

func TestFetchWaterQualityData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"waterLevel": 1.2,
			"waterQuality": "Good",
			"parameters": {
				"ph": 7.8,
				"dissolvedOxygen": 9.2
			},
			"nearestWater": "Amsterdam-Rijnkanaal",
			"distance": 320,
			"lastMeasured": "2024-01-15T10:00:00Z"
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		DigitalDeltaApiURL: server.URL,
	}
	client := NewApiClient(server.Client())

	data, err := client.FetchWaterQualityData(cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.WaterQuality != "Good" {
		t.Errorf("Expected water quality 'Good', got '%s'", data.WaterQuality)
	}
	if data.Parameters["ph"] != 7.8 {
		t.Errorf("Expected pH 7.8, got %f", data.Parameters["ph"])
	}
}

func TestFetchSafetyData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"crimeRate": 22.5,
			"safetyScore": 78,
			"crimeTypes": {
				"burglary": 12,
				"theft": 30
			},
			"policeResponse": 8.5,
			"yearOverYearChange": -2.3
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		SafetyExperienceApiURL: server.URL,
	}
	client := NewApiClient(server.Client())

	data, err := client.FetchSafetyData(cfg, "GM0344")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.SafetyScore != 78 {
		t.Errorf("Expected safety score 78, got %f", data.SafetyScore)
	}
	if data.CrimeRate != 22.5 {
		t.Errorf("Expected crime rate 22.5, got %f", data.CrimeRate)
	}
	if data.SafetyPerception != "Safe" {
		t.Errorf("Expected safety perception 'Safe', got '%s'", data.SafetyPerception)
	}
}

func TestFetchSchipholFlightData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"dailyFlights": 45,
			"noiseLevel": 52.3,
			"peakHours": ["07:00-09:00"],
			"flightPaths": [{
				"routeId": "Polderbaan",
				"altitude": 1200,
				"distance": 8500,
				"flightsPerDay": 8
			}],
			"nightFlights": 2,
			"noiseContour": "35 Ke"
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		SchipholApiURL: server.URL,
	}
	client := NewApiClient(server.Client())

	data, err := client.FetchSchipholFlightData(cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.DailyFlights != 45 {
		t.Errorf("Expected daily flights 45, got %d", data.DailyFlights)
	}
	if data.NoiseLevel != 52.3 {
		t.Errorf("Expected noise level 52.3, got %f", data.NoiseLevel)
	}
}
