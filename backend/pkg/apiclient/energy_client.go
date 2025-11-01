package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

// EnergyClimateData represents energy labels and climate risk
type EnergyClimateData struct {
	EnergyLabel      string  `json:"energyLabel"`      // A++++ to G
	ClimateRisk      string  `json:"climateRisk"`      // Low, Medium, High
	EfficiencyScore  float64 `json:"efficiencyScore"`  // 0-100
	AnnualEnergyCost float64 `json:"annualEnergyCost"` // EUR
	CO2Emissions     float64 `json:"co2Emissions"`     // kg/year
	HeatLoss         float64 `json:"heatLoss"`         // W/m²K
}

// FetchEnergyClimateData retrieves energy labels and climate risk for ESG scoring
// Documentation: https://docs.altum.ai/english/apis/energy-and-climate-api
func (c *ApiClient) FetchEnergyClimateData(cfg *config.Config, bagID string) (*EnergyClimateData, error) {
	if cfg.AltumEnergyApiURL == "" {
		return nil, fmt.Errorf("AltumEnergyApiURL not configured")
	}

	url := fmt.Sprintf("%s/energy/%s", cfg.AltumEnergyApiURL, bagID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if cfg.AltumEnergyApiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AltumEnergyApiKey))
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("energy data not found for BAG ID: %s", bagID)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("altum energy API returned status %d", resp.StatusCode)
	}

	var result EnergyClimateData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode energy climate response: %w", err)
	}

	return &result, nil
}

// SustainabilityData represents sustainability measures and CO2 savings potential
type SustainabilityData struct {
	CurrentRating       string                  `json:"currentRating"`
	PotentialRating     string                  `json:"potentialRating"`
	RecommendedMeasures []SustainabilityMeasure `json:"recommendedMeasures"`
	TotalCO2Savings     float64                 `json:"totalCO2Savings"`  // kg/year
	TotalCostSavings    float64                 `json:"totalCostSavings"` // EUR/year
	InvestmentCost      float64                 `json:"investmentCost"`   // EUR
	PaybackPeriod       float64                 `json:"paybackPeriod"`    // years
}

// SustainabilityMeasure represents a single energy improvement measure
type SustainabilityMeasure struct {
	Type         string  `json:"type"`
	Description  string  `json:"description"`
	CO2Savings   float64 `json:"co2Savings"`  // kg/year
	CostSavings  float64 `json:"costSavings"` // EUR/year
	Investment   float64 `json:"investment"`  // EUR
	PaybackYears float64 `json:"paybackYears"`
	Priority     int     `json:"priority"` // 1 = high, 3 = low
}

// FetchSustainabilityData retrieves sustainability recommendations and potential savings
// Documentation: https://docs.altum.ai/english/apis/sustainability-api
func (c *ApiClient) FetchSustainabilityData(cfg *config.Config, bagID string) (*SustainabilityData, error) {
	if cfg.AltumSustainabilityApiURL == "" {
		return nil, fmt.Errorf("AltumSustainabilityApiURL not configured")
	}

	url := fmt.Sprintf("%s/sustainability/%s", cfg.AltumSustainabilityApiURL, bagID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if cfg.AltumSustainabilityApiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AltumSustainabilityApiKey))
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("sustainability data not found for BAG ID: %s", bagID)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("altum sustainability API returned status %d", resp.StatusCode)
	}

	var result SustainabilityData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode sustainability response: %w", err)
	}

	return &result, nil
}
