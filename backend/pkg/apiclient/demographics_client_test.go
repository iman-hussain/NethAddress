package apiclient

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchCBSPopulationData(t *testing.T) {
	// Use transport mocking since the code now uses hardcoded PDOK URLs
	mockTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		// Check that the request is formatted correctly for PDOK OGC API
		if !strings.Contains(req.URL.Path, "/collections/buurten/items") {
			t.Errorf("Expected path to contain /collections/buurten/items, got %s", req.URL.Path)
		}

		responseBody := `{
			"type": "FeatureCollection",
			"features": [{
				"type": "Feature",
				"id": "buurt.1234",
				"properties": {
					"buurtcode": "BU03630001",
					"buurtnaam": "Centrum",
					"wijkcode": "WK036300",
					"gemeentecode": "GM0363",
					"gemeentenaam": "Amsterdam",
					"aantal_inwoners": 350000,
					"aantal_huishoudens": 150000,
					"gemiddelde_huishoudsgrootte": 2.3,
					"bevolkingsdichtheid_inwoners_per_km2": 12500,
					"percentage_personen_0_tot_15_jaar": 15,
					"percentage_personen_15_tot_25_jaar": 18,
					"percentage_personen_25_tot_45_jaar": 30,
					"percentage_personen_45_tot_65_jaar": 20,
					"percentage_personen_65_jaar_en_ouder": 17
				}
			}],
			"numberReturned": 1
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(responseBody)),
		}, nil
	})

	cfg := &config.Config{}
	client := NewApiClient(&http.Client{Transport: mockTransport})

	data, err := client.FetchCBSPopulationData(cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.TotalPopulation != 350000 {
		t.Errorf("Expected total population 350000, got %d", data.TotalPopulation)
	}
	if data.Households != 150000 {
		t.Errorf("Expected households 150000, got %d", data.Households)
	}
	if data.AverageHHSize != 2.3 {
		t.Errorf("Expected average household size 2.3, got %f", data.AverageHHSize)
	}
}

func TestFetchCBSStatLineData(t *testing.T) {
	called := false
	stub := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			called = true
			if req.URL.Path != "/ODataFeed/v4/CBS/84286NED/Observations" {
				return &http.Response{StatusCode: http.StatusNotFound, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
			}
			payload := map[string]any{
				"value": []map[string]any{
					{
						"RegioS":                               "UTRECHT",
						"Perioden":                             "2023",
						"BevolkingAanHetBeginVanDePeriode_1":   350000.0,
						"GemiddeldInkomenPerInwoner_66":        42500.0,
						"PercentageWerkloosPerLeeftijdsklasse": 21.5,
						"GemiddeldeWOZWaardeVanWoningen_35":    325.0,
						"Woningvoorraad_31":                    150000.0,
					},
				},
			}
			buf, _ := json.Marshal(payload)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(string(buf))),
			}, nil
		}),
	}

	cfg := &config.Config{
		CBSStatLineApiURL: "http://example.com",
	}
	client := NewApiClient(stub)

	data, err := client.FetchCBSStatLineData(cfg, "Utrecht")
	if !called {
		t.Fatalf("Expected CBS StatLine API stub to be called")
	}
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.AverageIncome != 42500 {
		t.Errorf("Expected average income 42500, got %f", data.AverageIncome)
	}
	if data.EmploymentRate != 78.5 {
		t.Errorf("Expected employment rate 78.5, got %f", data.EmploymentRate)
	}
	if data.AverageWOZ != 325000 {
		t.Errorf("Expected average WOZ 325000, got %f", data.AverageWOZ)
	}
}

func TestFetchCBSSquareStats(t *testing.T) {
	// Test with PDOK CBS OGC API response format
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/collections/buurten/items") {
			t.Errorf("Expected path to contain /collections/buurten/items, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"type": "FeatureCollection",
			"features": [{
				"type": "Feature",
				"id": "buurt.1234",
				"properties": {
					"buurtcode": "BU03630001",
					"aantal_inwoners": 245,
					"aantal_huishoudens": 106,
					"gemiddelde_woningwaarde": 325000,
					"omgevingsadressendichtheid": 85,
					"gemiddeld_gestandaardiseerd_inkomen_van_huishoudens": 420
				}
			}],
			"numberReturned": 1
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		CBSSquareStatsApiURL: server.URL,
	}
	client := NewApiClient(&http.Client{})

	data, err := client.FetchCBSSquareStats(cfg, 52.0907, 5.1214)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.GridID != "BU03630001" {
		t.Errorf("Expected grid ID BU03630001, got %s", data.GridID)
	}
	if data.Population != 245 {
		t.Errorf("Expected population 245, got %d", data.Population)
	}
	if data.Households != 106 {
		t.Errorf("Expected households 106, got %d", data.Households)
	}
	if data.AverageWOZ != 325000 {
		t.Errorf("Expected average WOZ 325000, got %f", data.AverageWOZ)
	}
}
