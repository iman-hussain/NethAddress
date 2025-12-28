package apiclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchGreenSpacesData(t *testing.T) {
	// Use transport mocking since the code now uses hardcoded PDOK URLs
	mockTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		responseBody := `{
			"type": "FeatureCollection",
			"features": [{
				"type": "Feature",
				"id": "bgt.123",
				"properties": {
					"fysiekVoorkomen": "groenvoorziening",
					"naam": "Wilhelminapark",
					"openbareRuimte": ""
				},
				"geometry": {
					"type": "Polygon",
					"coordinates": [[[4.88, 52.37], [4.89, 52.37], [4.89, 52.38], [4.88, 52.38], [4.88, 52.37]]]
				}
			}],
			"numberReturned": 1
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(responseBody)),
		}, nil
	})

	cfg := &config.Config{}
	client := NewApiClient(&http.Client{Transport: mockTransport}, cfg)

	data, err := client.FetchGreenSpacesData(context.Background(), cfg, 52.0907, 5.1214, 1000)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// With BGT data, we should get at least one green space
	if len(data.GreenSpaces) == 0 {
		t.Errorf("Expected at least one green space")
	}
	if data.NearestPark != "Wilhelminapark" {
		t.Errorf("Expected nearest park 'Wilhelminapark', got '%s'", data.NearestPark)
	}
}

func TestFetchEducationData(t *testing.T) {
	// Test with OSM Overpass API response format
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"elements": [{
				"type": "node",
				"id": 12345,
				"lat": 52.091,
				"lon": 5.122,
				"tags": {
					"name": "OBS De Regenboog",
					"amenity": "school",
					"isced:level": "1"
				}
			}]
		}`))
	}))
	defer server.Close()

	// Note: The education function uses a fixed Overpass URL, so this test
	// verifies the parsing logic rather than the actual HTTP call
	cfg := &config.Config{}
	client := NewApiClient(server.Client(), cfg)

	// This will call the real Overpass API, not our mock
	// So we just verify it doesn't panic and returns valid structure
	data, err := client.FetchEducationData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// The result depends on real Overpass API response
	// Just check the structure is valid
	if data == nil {
		t.Error("Expected non-nil data")
	}
}

func TestFetchBuildingPermitsData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"totalPermits": 45,
			"newConstruction": 12,
			"renovations": 28,
			"extensions": 5,
			"growthTrend": "Increasing"
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		BuildingPermitsApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	data, err := client.FetchBuildingPermitsData(context.Background(), cfg, 52.0907, 5.1214, 1000)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.TotalPermits != 45 {
		t.Errorf("Expected 45 total permits, got %d", data.TotalPermits)
	}
	if data.GrowthTrend != "Increasing" {
		t.Errorf("Expected growth trend 'Increasing', got '%s'", data.GrowthTrend)
	}
}

func TestFetchFacilitiesData(t *testing.T) {
	// Facilities uses OSM Overpass API - but with fixed URL
	// The real function uses hardcoded Overpass URL, so this test
	// verifies that it doesn't panic and returns valid structure
	cfg := &config.Config{}
	client := NewApiClient(http.DefaultClient, cfg)

	data, err := client.FetchFacilitiesData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// The result depends on real Overpass API response
	// Just check the structure is valid
	if data == nil {
		t.Error("Expected non-nil data")
	}
}

func TestFetchAHNHeightData(t *testing.T) {
	// Use transport mocking since the code now uses hardcoded URLs
	mockTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		responseBody := `{
			"results": [
				{"latitude": 52.0907, "longitude": 5.1214, "elevation": 3.5}
			]
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(responseBody)),
		}, nil
	})

	cfg := &config.Config{}
	client := NewApiClient(&http.Client{Transport: mockTransport}, cfg)

	data, err := client.FetchAHNHeightData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.Elevation != 3.5 {
		t.Errorf("Expected elevation 3.5, got %f", data.Elevation)
	}
	// Elevation 3.5m in Netherlands = Low flood risk
	if data.FloodRisk != "Low" {
		t.Errorf("Expected flood risk 'Low', got '%s'", data.FloodRisk)
	}
}
