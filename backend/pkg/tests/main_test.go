package tests

import (
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/apiclient"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
	"github.com/iman-hussain/AddressIQ/backend/pkg/scoring"
)

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func TestCalculateScoreResidential(t *testing.T) {
	agg := &models.AggregatedData{
		PDOKData: models.PDOKData{ZoningInfo: "Residential"},
	}
	score := scoring.CalculateScore(agg)
	if !floatEquals(score.Viability, 8.6) {
		t.Errorf("Expected Viability 8.6, got %v", score.Viability)
	}
	if !floatEquals(score.Investment, 5.3) {
		t.Errorf("Expected Investment 5.3, got %v", score.Investment)
	}
	if !floatEquals(score.ESG, 8.0) {
		t.Errorf("Expected ESG 8.0, got %v", score.ESG)
	}
}

func TestCalculateScoreNonResidential(t *testing.T) {
	agg := &models.AggregatedData{
		PDOKData: models.PDOKData{ZoningInfo: "Industrial"},
	}
	score := scoring.CalculateScore(agg)
	if !floatEquals(score.Viability, 4.4) {
		t.Errorf("Expected Viability 4.4, got %v", score.Viability)
	}
	if !floatEquals(score.Investment, 5.3) {
		t.Errorf("Expected Investment 5.3, got %v", score.Investment)
	}
	if !floatEquals(score.ESG, 8.0) {
		t.Errorf("Expected ESG 8.0, got %v", score.ESG)
	}
}

func TestFetchBAGDataMock(t *testing.T) {
	// Mock BAG API response
	mockResponse := `{"response":{"docs":[{"weergavenaam":"Teststraat 10, 1234AB Testdorp","straatnaam":"Teststraat","huisnummer":10,"huis_nlt":"10","postcode":"1234AB","woonplaatsnaam":"Testdorp","centroide_ll":"POINT(4.8952 52.3702)"}]}}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, strings.NewReader(mockResponse))
	}))
	defer ts.Close()
	httpClient := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		// Redirect all requests to the test server
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(ts.URL, "http://")
		return http.DefaultTransport.RoundTrip(req)
	})}
	client := apiclient.NewApiClient(httpClient)
	data, err := client.FetchBAGData("1234AB", "10")
	if err != nil {
		t.Fatalf("FetchBAGData failed: %v. Check if the mock server is running and the response format matches expected BAG API output.", err)
	}
	if data.Address == "" {
		t.Error("FetchBAGData: Address field is empty. Ensure mock response includes address fields.")
	}
	if data.GeoJSON == "" {
		t.Error("FetchBAGData: GeoJSON field is empty. Ensure mock response includes geometry.")
	}
}

func TestFetchPDOKDataMock(t *testing.T) {
	// Mock PDOK API response (minimal XML with <omschrijving> and <beperkingen>)
	mockXML := `<root><omschrijving>Wonen</omschrijving><beperkingen>Geen</beperkingen></root>`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		io.Copy(w, strings.NewReader(mockXML))
	}))
	defer ts.Close()
	httpClient := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(ts.URL, "http://")
		return http.DefaultTransport.RoundTrip(req)
	})}
	client := apiclient.NewApiClient(httpClient)
	data, err := client.FetchPDOKData("52.3702,4.8952")
	if err != nil {
		t.Fatalf("FetchPDOKData failed: %v. Check if the mock server is running and the response format matches expected PDOK API output.", err)
	}
	if data.ZoningInfo == "" {
		t.Error("FetchPDOKData: ZoningInfo field is empty. Ensure mock response includes <omschrijving>.")
	}
	if len(data.Restrictions) == 0 {
		t.Error("FetchPDOKData: Restrictions field is empty. Ensure mock response includes <beperkingen>.")
	}
}
