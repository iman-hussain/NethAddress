package apiclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchPropertyValuePlus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"marketValue": 425000,
			"valuationDate": "2024-11-01",
			"confidence": 0.92,
			"pricePerSqm": 3863,
			"comparableProperties": [
				{
					"address": "Test Street 10",
					"distance": 150,
					"salePrice": 415000,
					"saleDate": "2024-09-15",
					"surfaceArea": 115,
					"propertyType": "Tussenwoning"
				}
			],
			"features": {
				"hasGarden": true,
				"hasParking": false,
				"energyLabel": "B"
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		MatrixianApiURL: server.URL,
		MatrixianApiKey: "test-key",
	}
	client := NewApiClient(nil, cfg)

	result, err := client.FetchPropertyValuePlus(cfg, "test-bag-id", 52.37, 4.89)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.MarketValue != 425000 {
		t.Errorf("Expected market value 425000, got %f", result.MarketValue)
	}
	if result.Confidence != 0.92 {
		t.Errorf("Expected confidence 0.92, got %f", result.Confidence)
	}
	if len(result.ComparableProperties) != 1 {
		t.Errorf("Expected 1 comparable property, got %d", len(result.ComparableProperties))
	}
	if result.PricePerSqm != 3863 {
		t.Errorf("Expected price per sqm 3863, got %f", result.PricePerSqm)
	}
}
