package apiclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchMonumentData(t *testing.T) {
	// The BAG Pand ID-based method now returns not-a-monument as we use coordinate-based lookup
	// This test verifies the function doesn't error
	client := NewApiClient(http.DefaultClient)
	cfg := &config.Config{}

	data, err := client.FetchMonumentData(cfg, "1234567890")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// BAG Pand ID method returns not-a-monument by design (coordinate method is preferred)
	if data == nil {
		t.Fatal("Expected non-nil data")
	}
}

func TestFetchMonumentDataByCoords(t *testing.T) {
	// Test with PDOK RCE beschermde-gebieden-cultuurhistorie OGC API response format
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// PDOK RCE returns GeoJSON FeatureCollection
		w.Write([]byte(`{
			"type": "FeatureCollection",
			"features": [{
				"type": "Feature",
				"id": "rijksmonument.12345",
				"properties": {
					"monumentnummer": "12345",
					"rijksmonument_naam": "Anne Frank Huis",
					"monumenttype": "Rijksmonument"
				},
				"geometry": {
					"type": "Point",
					"coordinates": [4.8837, 52.3753]
				}
			}],
			"numberReturned": 1
		}`))
	}))
	defer server.Close()

	client := NewApiClient(server.Client())
	cfg := &config.Config{MonumentenApiURL: server.URL}

	data, err := client.FetchMonumentDataByCoords(cfg, 52.3753, 4.8837)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !data.IsMonument {
		t.Fatal("Expected monument flag to be true")
	}
	if data.Type != "Rijksmonument" {
		t.Errorf("Expected monument type 'Rijksmonument', got '%s'", data.Type)
	}
}
