package apiclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
)

func TestFetchWURSoilData(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"soilType": "Clay",
			"composition": "Clay with silt",
			"permeability": 1.2,
			"organicMatter": 4.2,
			"ph": 7.1,
			"suitability": "Good"
		}`))
	}))
	defer server.Close()

	// Create client with mock URL
	cfg := &config.Config{
		WURSoilApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	// Test successful fetch
	data, err := client.FetchWURSoilData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.SoilType != "Clay" {
		t.Errorf("Expected soil type 'Clay', got '%s'", data.SoilType)
	}
	if data.OrganicMatter != 4.2 {
		t.Errorf("Expected organic matter 4.2, got %f", data.OrganicMatter)
	}
	if data.Suitability != "Good" {
		t.Errorf("Expected suitability 'Good', got '%s'", data.Suitability)
	}
}

func TestFetchSubsidenceData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"subsidenceRate": -2.5,
			"totalSubsidence": -15.3,
			"stabilityRating": "Moderate",
			"measurementDate": "2024-01-01",
			"groundMovement": -0.8
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		SkyGeoSubsidenceApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	data, err := client.FetchSubsidenceData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.SubsidenceRate != -2.5 {
		t.Errorf("Expected subsidence rate -2.5, got %f", data.SubsidenceRate)
	}
	if data.StabilityRating != "Moderate" {
		t.Errorf("Expected stability rating 'Moderate', got '%s'", data.StabilityRating)
	}
}

func TestFetchSoilQualityData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"contaminationLevel": "None",
			"contaminants": [],
			"qualityZone": "Woonwijk",
			"restrictedUse": false,
			"lastTested": "2023-01-01"
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		SoilQualityApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	data, err := client.FetchSoilQualityData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.ContaminationLevel != "None" {
		t.Errorf("Expected contamination level 'None', got '%s'", data.ContaminationLevel)
	}
}

func TestFetchBROSoilMapData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"soilType": "Sand",
			"peatComposition": 12.5,
			"profile": "Holocene",
			"foundationQuality": "Good",
			"groundwaterDepth": 1.2
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		BROSoilMapApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	data, err := client.FetchBROSoilMapData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.SoilType != "Sand" {
		t.Errorf("Expected soil type 'Sand', got '%s'", data.SoilType)
	}
	if data.GroundwaterDepth != 1.2 {
		t.Errorf("Expected groundwater depth 1.2, got %f", data.GroundwaterDepth)
	}
}
