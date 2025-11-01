package apiclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchGreenSpacesData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"totalGreenArea": 18500,
			"greenPercentage": 35.5,
			"nearestPark": "Wilhelminapark",
			"parkDistance": 420,
			"treeCanopyCover": 28.4,
			"greenSpaces": [{
				"name": "Wilhelminapark",
				"type": "Park",
				"area": 52000,
				"distance": 420,
				"facilities": ["playground", "pond"]
			}]
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		GreenSpacesApiURL: server.URL,
	}
	client := NewApiClient(server.Client())

	data, err := client.FetchGreenSpacesData(cfg, 52.0907, 5.1214, 1000)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.GreenPercentage != 35.5 {
		t.Errorf("Expected green percentage 35.5, got %f", data.GreenPercentage)
	}
	if data.NearestPark != "Wilhelminapark" {
		t.Errorf("Expected nearest park 'Wilhelminapark', got '%s'", data.NearestPark)
	}
}

func TestFetchEducationData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"nearestPrimarySchool": {
				"name": "OBS De Regenboog",
				"type": "Primary",
				"distance": 450,
				"qualityScore": 7.6,
				"students": 320,
				"address": "Julianalaan 10",
				"denomination": "Public"
			},
			"allSchools": [{
				"name": "Het Baarnsch Lyceum",
				"type": "Secondary",
				"distance": 850,
				"qualityScore": 8.2,
				"students": 950,
				"address": "Stationsweg 16",
				"denomination": "Special"
			}],
			"averageQuality": 7.8
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		EducationApiURL: server.URL,
	}
	client := NewApiClient(server.Client())

	data, err := client.FetchEducationData(cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(data.AllSchools) != 1 {
		t.Fatalf("Expected 1 school, got %d", len(data.AllSchools))
	}
	if data.AverageQuality != 7.8 {
		t.Errorf("Expected average quality 7.8, got %f", data.AverageQuality)
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
	client := NewApiClient(server.Client())

	data, err := client.FetchBuildingPermitsData(cfg, 52.0907, 5.1214, 1000)
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"supermarkets": 5,
			"restaurants": 12,
			"healthcare": 3,
			"amenitiesScore": 82
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		FacilitiesApiURL: server.URL,
	}
	client := NewApiClient(server.Client())

	data, err := client.FetchFacilitiesData(cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.AmenitiesScore != 82 {
		t.Errorf("Expected amenities score 82, got %f", data.AmenitiesScore)
	}
}

func TestFetchAHNHeightData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"elevation": 5.2,
			"terrainSlope": 1.8,
			"surrounding": [4.9, 5.3, 5.0]
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		AHNHeightModelApiURL: server.URL,
	}
	client := NewApiClient(server.Client())

	data, err := client.FetchAHNHeightData(cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.Elevation != 5.2 {
		t.Errorf("Expected elevation 5.2, got %f", data.Elevation)
	}
	if data.FloodRisk != "Low" {
		t.Errorf("Expected flood risk 'Low', got '%s'", data.FloodRisk)
	}
}
