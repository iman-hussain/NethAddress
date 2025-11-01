package apiclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchKadasterObjectInfo(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/objecten/test-bag-id" {
			t.Errorf("Expected path /objecten/test-bag-id, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"eigenaar": {"naam": "John Doe"},
			"kadaster": {"referentie": "KAD123"},
			"woz": {"waarde": 350000},
			"energie": {"label": "B"},
			"belastingen": {"gemeentelijk": 1200},
			"oppervlakte": {"wonen": 120, "perceel": 250},
			"gebouw": {"type": "Vrijstaand", "bouwjaar": 1995}
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		KadasterObjectInfoApiURL: server.URL,
	}
	client := NewApiClient(nil)

	result, err := client.FetchKadasterObjectInfo(cfg, "test-bag-id")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.OwnerName != "John Doe" {
		t.Errorf("Expected owner 'John Doe', got %s", result.OwnerName)
	}
	if result.WOZValue != 350000 {
		t.Errorf("Expected WOZ value 350000, got %f", result.WOZValue)
	}
	if result.EnergyLabel != "B" {
		t.Errorf("Expected energy label 'B', got %s", result.EnergyLabel)
	}
	if result.BuildYear != 1995 {
		t.Errorf("Expected build year 1995, got %d", result.BuildYear)
	}
}

func TestFetchKadasterObjectInfo_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.Config{
		KadasterObjectInfoApiURL: server.URL,
	}
	client := NewApiClient(nil)

	_, err := client.FetchKadasterObjectInfo(cfg, "nonexistent")
	if err == nil {
		t.Error("Expected error for 404 response, got nil")
	}
}
