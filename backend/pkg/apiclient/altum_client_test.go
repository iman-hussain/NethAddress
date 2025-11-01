package apiclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchAltumWOZData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"wozValue": 400000,
			"valueYear": 2024,
			"buildingType": "Tussenwoning",
			"buildYear": 1980,
			"surfaceArea": 110,
			"latitude": 52.37,
			"longitude": 4.89
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		AltumWOZApiURL: server.URL,
		AltumWOZApiKey: "test-key",
	}
	client := NewApiClient(nil)

	result, err := client.FetchAltumWOZData(cfg, "test-bag-id")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.WOZValue != 400000 {
		t.Errorf("Expected WOZ value 400000, got %f", result.WOZValue)
	}
	if result.BuildingType != "Tussenwoning" {
		t.Errorf("Expected building type 'Tussenwoning', got %s", result.BuildingType)
	}
}

func TestFetchTransactionHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"transactions": [
				{
					"transactionId": "TX123",
					"date": "2023-05-15",
					"purchasePrice": 380000,
					"propertyType": "Woning",
					"surfaceArea": 110,
					"bagObjectId": "test-bag-id"
				}
			],
			"totalCount": 1
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		AltumTransactionApiURL: server.URL,
		AltumTransactionApiKey: "test-key",
	}
	client := NewApiClient(nil)

	result, err := client.FetchTransactionHistory(cfg, "test-bag-id")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.TotalCount != 1 {
		t.Errorf("Expected 1 transaction, got %d", result.TotalCount)
	}
	if len(result.Transactions) == 0 {
		t.Fatal("Expected transactions, got empty array")
	}
	if result.Transactions[0].PurchasePrice != 380000 {
		t.Errorf("Expected price 380000, got %f", result.Transactions[0].PurchasePrice)
	}
}

func TestFetchTransactionHistory_NoData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.Config{
		AltumTransactionApiURL: server.URL,
	}
	client := NewApiClient(nil)

	result, err := client.FetchTransactionHistory(cfg, "test-bag-id")
	if err != nil {
		t.Fatalf("Expected no error for 404, got %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("Expected 0 transactions for 404, got %d", result.TotalCount)
	}
}
