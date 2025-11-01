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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"totalPopulation": 350000,
			"households": 150000,
			"averageHouseholdSize": 2.3,
			"ageDistribution": {
				"0-14": 22,
				"15-24": 18,
				"25-44": 30,
				"45-64": 20,
				"65+": 10
			},
			"demographics": {
				"age0to14": 22000,
				"age15to24": 18000,
				"age25to44": 30000,
				"age45to64": 20000,
				"age65plus": 10000
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		CBSPopulationApiURL: server.URL,
	}
	client := NewApiClient(&http.Client{})

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"gridId": "100m_35821",
			"population": 245,
			"households": 106,
			"averageWOZ": 325000,
			"averageIncome": 42000,
			"housingDensity": 85
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

	if data.GridID != "100m_35821" {
		t.Errorf("Expected grid ID 100m_35821, got %s", data.GridID)
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
