package apiclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
)

func TestFetchPDOKPlatformData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"cadastralData": {
				"parcelId": "UTR01-A-1234",
				"municipality": "Utrecht",
				"section": "A",
				"parcelNumber": "1234",
				"area": 245,
				"landUse": "Residential"
			},
			"addressData": {
				"bagId": "0123456789012345",
				"fullAddress": "Julianalaan 1, Utrecht",
				"postalCode": "3581BB",
				"latitude": 52.0907,
				"longitude": 5.1214
			},
			"topographyData": {
				"landType": "Urban",
				"terrainFeatures": ["Flat"],
				"waterBodies": ["Canal"]
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		PDOKApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	data, err := client.FetchPDOKPlatformData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.CadastralData.ParcelID != "UTR01-A-1234" {
		t.Errorf("Expected parcel ID 'UTR01-A-1234', got '%s'", data.CadastralData.ParcelID)
	}
	if data.CadastralData.LandUse != "Residential" {
		t.Errorf("Expected land use 'Residential', got '%s'", data.CadastralData.LandUse)
	}
}

func TestFetchStratopoEnvironmentData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"environmentScore": 78.5,
			"totalVariables": 725,
			"pollutionIndex": 15.2,
			"urbanizationLevel": "Urban",
			"environmentFactors": {
				"noise": 42,
				"airQuality": "Good"
			},
			"esgRating": "A-",
			"recommendations": ["Plant more trees"]
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		StratopoApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	data, err := client.FetchStratopoEnvironmentData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.EnvironmentScore != 78.5 {
		t.Errorf("Expected environment score 78.5, got %f", data.EnvironmentScore)
	}
	if data.ESGRating != "A-" {
		t.Errorf("Expected ESG rating 'A-', got '%s'", data.ESGRating)
	}
}

func TestFetchLandUseData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"primaryUse": "Residential",
			"zoningCode": "WR-1",
			"zoningDetails": "Woongebied",
			"restrictions": ["Max height 15m"],
			"allowedUses": ["Residential", "Small retail"],
			"buildingRights": {
				"maxHeight": 15,
				"maxBuildArea": 350,
				"floorAreaRatio": 1.6,
				"groundCoverage": 60,
				"canSubdivide": false,
				"canExpand": true
			},
			"futurePlans": [{
				"planName": "Urban Renewal 2025",
				"type": "Redevelopment",
				"status": "Approved",
				"expectedDate": "2025-09-01",
				"impact": "Positive"
			}]
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		LandUseApiURL: server.URL,
	}
	client := NewApiClient(server.Client(), cfg)

	data, err := client.FetchLandUseData(context.Background(), cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.PrimaryUse != "Residential" {
		t.Errorf("Expected primary use 'Residential', got '%s'", data.PrimaryUse)
	}
	if !data.BuildingRights.CanExpand {
		t.Error("Expected CanExpand to be true")
	}
	if len(data.FuturePlans) != 1 {
		t.Errorf("Expected 1 future plan, got %d", len(data.FuturePlans))
	}
}
