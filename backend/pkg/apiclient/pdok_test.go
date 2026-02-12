package apiclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
)

func TestFetchPDOKData_RealAPI(t *testing.T) {
	// Mock WFS GeoJSON response
	mockJSON := `{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"properties": {
					"naam": "Wonen",
					"plantype": "bestemmingsplan",
					"planstatus": "vastgesteld",
					"beleidsmatigstatus": "Geen bedrijfsmatige activiteiten"
				},
				"geometry": {
					"type": "Polygon",
					"coordinates": [[[4.895, 52.370], [4.896, 52.370], [4.896, 52.371], [4.895, 52.371], [4.895, 52.370]]]
				}
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	cfg := &config.Config{}
	// Use a custom RoundTripper to redirect requests to the mock server
	client := NewApiClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Host = server.URL[len("http://"):]
			req.URL.Scheme = "http"
			return http.DefaultTransport.RoundTrip(req)
		}),
	}, cfg)

	pdokData, err := client.FetchPDOKData(context.Background(), "4.8952,52.3702")
	if err != nil {
		t.Fatalf("FetchPDOKData failed: %v", err)
	}
	if pdokData.ZoningInfo != "Wonen" {
		t.Errorf("Expected zoning 'Wonen', got '%s'", pdokData.ZoningInfo)
	}
	if len(pdokData.Restrictions) == 0 {
		t.Errorf("Expected restrictions, got none")
	}
}
