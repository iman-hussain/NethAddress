package apiclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

func TestFetchEnergyClimateData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"energyLabel": "B",
			"climateRisk": "Medium",
			"efficiencyScore": 74.5,
			"annualEnergyCost": 1850,
			"co2Emissions": 3250,
			"heatLoss": 42.1
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		AltumEnergyApiURL: server.URL,
	}
	client := NewApiClient(server.Client())

	data, err := client.FetchEnergyClimateData(cfg, "0123456789012345")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.EnergyLabel != "B" {
		t.Errorf("Expected energy label 'B', got '%s'", data.EnergyLabel)
	}
	if data.EfficiencyScore != 74.5 {
		t.Errorf("Expected efficiency score 74.5, got %f", data.EfficiencyScore)
	}
	if data.AnnualEnergyCost != 1850 {
		t.Errorf("Expected annual energy cost 1850, got %f", data.AnnualEnergyCost)
	}
}

func TestFetchSustainabilityData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"currentRating": "Label C",
			"potentialRating": "Label A",
			"totalCO2Savings": 2500,
			"totalCostSavings": 850,
			"investmentCost": 12000,
			"paybackPeriod": 8.5,
			"recommendedMeasures": [
				{
					"type": "insulation",
					"description": "Spouwmuurisolatie",
					"co2Savings": 850,
					"costSavings": 220,
					"investment": 3500,
					"paybackYears": 6.5,
					"priority": 1
				}
			]
		}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		AltumSustainabilityApiURL: server.URL,
	}
	client := NewApiClient(server.Client())

	data, err := client.FetchSustainabilityData(cfg, "0123456789012345")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data.CurrentRating != "Label C" {
		t.Errorf("Expected current rating 'Label C', got '%s'", data.CurrentRating)
	}
	if data.TotalCO2Savings != 2500 {
		t.Errorf("Expected total CO2 savings 2500, got %f", data.TotalCO2Savings)
	}
	if data.PaybackPeriod != 8.5 {
		t.Errorf("Expected payback period 8.5, got %f", data.PaybackPeriod)
	}
}

func TestFetchEnergyClimateData_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.Config{
		AltumEnergyApiURL: server.URL,
	}
	client := NewApiClient(server.Client())

	_, err := client.FetchEnergyClimateData(cfg, "0000000000000000")
	if err == nil {
		t.Error("Expected error for non-existent address, got nil")
	}
}
