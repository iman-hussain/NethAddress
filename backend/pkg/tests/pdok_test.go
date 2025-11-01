package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/apiclient"
)

// Use roundTripperFunc from bag_test.go

func TestFetchPDOKData_RealAPI(t *testing.T) {
	// Mock WMS XML response
	mockXML := `<?xml version="1.0" encoding="UTF-8"?>
	<FeatureInfo>
		<omschrijving>Wonen</omschrijving>
		<beperkingen>Geen bedrijfsmatige activiteiten</beperkingen>
	</FeatureInfo>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(mockXML))
	}))
	defer server.Close()

	// Use a custom RoundTripper to redirect requests to the mock server
	client := apiclient.NewApiClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Host = server.URL[len("http://"):]
			req.URL.Scheme = "http"
			return http.DefaultTransport.RoundTrip(req)
		}),
	})

	pdokData, err := client.FetchPDOKData("4.8952,52.3702")
	if err != nil {
		t.Fatalf("FetchPDOKData failed: %v", err)
	}
	if pdokData.ZoningInfo != "Wonen" {
		t.Errorf("Expected zoning 'Wonen', got '%s'", pdokData.ZoningInfo)
	}
	if len(pdokData.Restrictions) == 0 || pdokData.Restrictions[0] != "Geen bedrijfsmatige activiteiten" {
		t.Errorf("Expected restriction 'Geen bedrijfsmatige activiteiten', got '%v'", pdokData.Restrictions)
	}
}
