package apiclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchMonumentData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"_embedded": map[string]any{
				"monumenten": []map[string]any{
					{
						"monumentnummer":  12345,
						"adressering":     "Singel 1",
						"type":            "Pand",
						"status":          "Rijksmonument",
						"datumAanwijzing": "1985-06-01",
						"betreftBagPand": []map[string]any{
							{"identificatie": "1234567890"},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewApiClient(server.Client())
	cfg := &config.Config{MonumentenApiURL: server.URL}

	data, err := client.FetchMonumentData(cfg, "1234567890")
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
