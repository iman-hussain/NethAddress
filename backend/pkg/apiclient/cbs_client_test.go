package apiclient

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchCBSData(t *testing.T) {
	called := false
	stub := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			called = true
			body := `{"value": [{"GemiddeldInkomenPerInkomensontvanger_68": 35.5, "Bevolkingsdichtheid_33": 3200.0, "GemiddeldeWOZWaardeVanWoningen_35": 312.0}]}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	cfg := &config.Config{CBSApiURL: "http://example.com"}
	client := NewApiClient(stub, cfg)

	data, err := client.FetchCBSData(context.Background(), cfg, "GM0344")
	if !called {
		t.Fatalf("Expected CBS API test server to be called, got err: %v", err)
	}
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.AvgIncome != 35500 {
		t.Errorf("Expected average income 35500, got %f", data.AvgIncome)
	}
	if data.PopulationDensity != 3200 {
		t.Errorf("Expected population density 3200, got %f", data.PopulationDensity)
	}
	if data.AvgWOZValue != 312000 {
		t.Errorf("Expected average WOZ 312000, got %f", data.AvgWOZValue)
	}
}
