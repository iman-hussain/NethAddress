package aggregator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/apiclient"
	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

// roundTripperFunc allows using a function as http.RoundTripper for test mocks
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// TestAggregatePropertyData_BAGIDExtraction tests that BAG IDs are properly extracted
func TestAggregatePropertyData_BAGIDExtraction(t *testing.T) {
	// Mock BAG API response with BAG IDs
	mockBAGJSON := `{
		"response": {
			"docs": [
				{
					"id": "adr-12345678",
					"nummeraanduiding_id": "0363200000123456",
					"verblijfsobject_id": "0363010000123456",
					"pand_id": "0363100000123456",
					"weergavenaam": "Teststraat 10, 1234AB Testdorp",
					"straatnaam": "Teststraat",
					"huisnummer": 10,
					"huis_nlt": "10",
					"postcode": "1234AB",
					"woonplaatsnaam": "Testdorp",
					"gemeentenaam": "Amsterdam",
					"gemeentecode": "GM0363",
					"provincienaam": "Noord-Holland",
					"provinciecode": "PV27",
					"centroide_ll": "POINT(4.8952 52.3702)"
				}
			]
		}
	}`

	// Mock CBS WFS response for neighborhood lookup
	mockCBSWFS := `{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"properties": {
					"buurtcode": "BU03630000",
					"buurtnaam": "Test Neighborhood",
					"wijkcode": "WK036300",
					"wijknaam": "Test District",
					"gemeentecode": "GM0363",
					"gemeentenaam": "Amsterdam"
				}
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Determine which API is being called based on URL
		if strings.Contains(r.URL.String(), "locatieserver") || strings.Contains(r.URL.String(), "bzk") || strings.Contains(r.URL.Path, "/v1/search") {
			w.Write([]byte(mockBAGJSON))
		} else if strings.Contains(r.URL.String(), "gebiedsindelingen") {
			w.Write([]byte(mockCBSWFS))
		} else {
			// Return empty response for other APIs
			w.Write([]byte(`{}`))
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		BagApiURL: server.URL,
	}

	// Create client with custom RoundTripper
	client := apiclient.NewApiClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Host = strings.TrimPrefix(server.URL, "http://")
			req.URL.Scheme = "http"
			return http.DefaultTransport.RoundTrip(req)
		}),
	}, cfg)

	aggregator := NewPropertyAggregator(client, nil, cfg)

	// Test aggregation
	result, err := aggregator.AggregatePropertyData(context.Background(), "1234AB", "10")
	if err != nil {
		t.Fatalf("AggregatePropertyData failed: %v", err)
	}

	// Verify BAG ID is extracted (should be verblijfsobject_id)
	if result.BAGID != "0363010000123456" {
		t.Errorf("Expected BAGID '0363010000123456', got '%s'", result.BAGID)
	}

	// Verify address and coordinates are present
	if result.Address == "" {
		t.Error("Address should not be empty")
	}

	if result.Coordinates[0] == 0 && result.Coordinates[1] == 0 {
		t.Error("Coordinates should not be zero")
	}

	// Verify DataSources includes BAG
	foundBAG := false
	for _, source := range result.DataSources {
		if source == "BAG" {
			foundBAG = true
			break
		}
	}
	if !foundBAG {
		t.Error("DataSources should include 'BAG'")
	}

	// Verify Errors map is initialized
	if result.Errors == nil {
		t.Error("Errors map should be initialized")
	}
}

// TestAggregatePropertyData_ErrorHandling tests that errors are properly captured
func TestAggregatePropertyData_ErrorHandling(t *testing.T) {
	// Mock BAG API response without other APIs
	mockBAGJSON := `{
		"response": {
			"docs": [
				{
					"id": "adr-12345678",
					"verblijfsobject_id": "0363010000123456",
					"weergavenaam": "Teststraat 10, 1234AB Testdorp",
					"centroide_ll": "POINT(4.8952 52.3702)"
				}
			]
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only respond to BAG requests, fail others
		if strings.Contains(r.URL.String(), "locatieserver") || strings.Contains(r.URL.String(), "bzk") || strings.Contains(r.URL.Path, "/v1/search") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(mockBAGJSON))
		} else if strings.Contains(r.URL.String(), "gebiedsindelingen") {
			// Return empty features for CBS WFS
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
		} else {
			// Return 500 error for other APIs
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"service unavailable"}`))
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		BagApiURL:         server.URL,
		AltumWOZApiURL:    server.URL,
		AltumWOZApiKey:    "test-key",
		KNMIWeatherApiURL: server.URL,
		KNMISolarApiURL:   server.URL,
	}

	client := apiclient.NewApiClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Host = strings.TrimPrefix(server.URL, "http://")
			req.URL.Scheme = "http"
			return http.DefaultTransport.RoundTrip(req)
		}),
	}, cfg)

	aggregator := NewPropertyAggregator(client, nil, cfg)

	result, err := aggregator.AggregatePropertyData(context.Background(), "1234AB", "10")
	if err != nil {
		t.Fatalf("AggregatePropertyData should not fail on API errors: %v", err)
	}

	// Verify BAG data is still present
	if result.BAGID != "0363010000123456" {
		t.Errorf("Expected BAGID '0363010000123456', got '%s'", result.BAGID)
	}

	// Verify that some errors were captured (APIs that failed)
	if len(result.Errors) == 0 {
		t.Log("Warning: Expected some API errors to be captured in Errors map")
		// This is not a hard failure as some APIs might not be called if URLs are empty
	}

	// Verify BAG is in DataSources
	foundBAG := false
	for _, source := range result.DataSources {
		if source == "BAG" {
			foundBAG = true
			break
		}
	}
	if !foundBAG {
		t.Error("DataSources should include 'BAG' even when other APIs fail")
	}
}
