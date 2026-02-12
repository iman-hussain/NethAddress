package apiclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
)

func TestFetchNDWTrafficData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": [
				{
					"locationId": "A2-123",
					"intensity": 2500,
					"averageSpeed": 85.5,
					"congestionLevel": "Free",
					"lastUpdated": "2024-01-15T10:00:00Z",
					"coordinates": {"lat": 52.1, "lon": 5.1}
				},
				{
					"locationId": "N201-42",
					"intensity": 850,
					"averageSpeed": 62.3,
					"congestionLevel": "Moderate",
					"lastUpdated": "2024-01-15T10:00:00Z",
					"coordinates": {"lat": 52.09, "lon": 5.12}
				}
			]
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		NDWTrafficApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	data, err := client.FetchNDWTrafficData(context.Background(), cfg, 52.0907, 5.1214, 2000)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(data) != 2 {
		t.Fatalf("Expected 2 traffic data points, got %d", len(data))
	}

	if data[0].LocationID != "A2-123" {
		t.Errorf("Expected location ID 'A2-123', got '%s'", data[0].LocationID)
	}
	if data[0].AverageSpeed != 85.5 {
		t.Errorf("Expected average speed 85.5, got %f", data[0].AverageSpeed)
	}
}

func TestFetchOpenOVData(t *testing.T) {
	// Note: The OpenOV function uses hardcoded OVapi URLs with coordinate-based lookups
	// The function first finds nearby stop area codes via PDOK BAG, then queries OVapi
	// This test verifies that the function returns a valid structure
	cfg := &config.Config{}
	client := NewApiClient(http.DefaultClient, cfg)

	data, err := client.FetchOpenOVData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// The result depends on real OVapi response
	// Just check the structure is valid
	if data == nil {
		t.Error("Expected non-nil data")
	}
}

func TestFetchParkingData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"totalSpaces": 120,
			"availableSpaces": 45,
			"occupancyRate": 62.5,
			"parkingZones": [{
				"zoneId": "P1",
				"name": "P+R Utrecht Science Park",
				"type": "Garage",
				"capacity": 500,
				"available": 120,
				"hourlyRate": 3.5,
				"coordinates": {"lat": 52.08, "lon": 5.17}
			}]
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		ParkingApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	data, err := client.FetchParkingData(context.Background(), cfg, 52.0907, 5.1214, 1000)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.AvailableSpaces != 45 {
		t.Errorf("Expected 45 available spaces, got %d", data.AvailableSpaces)
	}
	if len(data.ParkingZones) != 1 {
		t.Errorf("Expected 1 parking zone, got %d", len(data.ParkingZones))
	}
}
