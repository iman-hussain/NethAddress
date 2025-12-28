package tests

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

func TestFetchBAGData_RealAPI(t *testing.T) {
	// Mock BAG API response with BAG IDs
	mockJSON := `{
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	cfg := &config.Config{}
	// Use a custom RoundTripper to redirect requests to the mock server
	client := apiclient.NewApiClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			// Replace the host with the mock server
			req.URL.Host = strings.TrimPrefix(server.URL, "http://")
			req.URL.Scheme = "http"
			return http.DefaultTransport.RoundTrip(req)
		}),
	}, cfg)

	bagData, err := client.FetchBAGData(context.Background(), "1234AB", "10")
	if err != nil {
		t.Fatalf("FetchBAGData failed: %v", err)
	}
	if bagData.Address == "" || bagData.GeoJSON == "" {
		t.Errorf("Expected non-empty address and geojson")
	}
	if len(bagData.Coordinates) < 2 || bagData.Coordinates[0] != 4.8952 || bagData.Coordinates[1] != 52.3702 {
		t.Errorf("Coordinates not parsed correctly: %v", bagData.Coordinates)
	}
	// Verify BAG IDs are extracted
	if bagData.VerblijfsobjectID != "0363010000123456" {
		t.Errorf("Expected verblijfsobject_id '0363010000123456', got '%s'", bagData.VerblijfsobjectID)
	}
	if bagData.ID != "0363010000123456" {
		t.Errorf("Expected ID to be verblijfsobject_id '0363010000123456', got '%s'", bagData.ID)
	}
	if bagData.MunicipalityCode != "GM0363" {
		t.Errorf("Expected municipality code 'GM0363', got '%s'", bagData.MunicipalityCode)
	}
}
