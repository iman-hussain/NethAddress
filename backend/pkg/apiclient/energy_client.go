package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// FetchEnergyClimateData retrieves energy labels and climate risk for ESG scoring
// Documentation: https://docs.altum.ai/english/apis/energy-and-climate-api
func (c *ApiClient) FetchEnergyClimateData(ctx context.Context, cfg *config.Config, bagID string) (*models.EnergyClimateData, error) {
	if cfg.AltumEnergyApiURL == "" {
		// Return default data when API is not configured (paid service)
		return &models.EnergyClimateData{
			EnergyLabel:      "Unknown",
			ClimateRisk:      "Unknown",
			EfficiencyScore:  0,
			AnnualEnergyCost: 0,
			CO2Emissions:     0,
			HeatLoss:         0,
		}, nil
	}

	url := fmt.Sprintf("%s/energy/%s", cfg.AltumEnergyApiURL, bagID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if cfg.AltumEnergyApiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AltumEnergyApiKey))
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		// Return default data on network error
		return &models.EnergyClimateData{
			EnergyLabel:      "Unknown",
			ClimateRisk:      "Unknown",
			EfficiencyScore:  0,
			AnnualEnergyCost: 0,
			CO2Emissions:     0,
			HeatLoss:         0,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("energy data not found for BAG ID: %s", bagID)
	}

	if resp.StatusCode != 200 {
		// Return default data on non-200 status
		return &models.EnergyClimateData{
			EnergyLabel:      "Unknown",
			ClimateRisk:      "Unknown",
			EfficiencyScore:  0,
			AnnualEnergyCost: 0,
			CO2Emissions:     0,
			HeatLoss:         0,
		}, nil
	}

	var result models.EnergyClimateData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// Return default data on parse error
		return &models.EnergyClimateData{
			EnergyLabel:      "Unknown",
			ClimateRisk:      "Unknown",
			EfficiencyScore:  0,
			AnnualEnergyCost: 0,
			CO2Emissions:     0,
			HeatLoss:         0,
		}, nil
	}

	return &result, nil
}

// FetchSustainabilityData retrieves sustainability recommendations and potential savings
// Documentation: https://docs.altum.ai/english/apis/sustainability-api
func (c *ApiClient) FetchSustainabilityData(ctx context.Context, cfg *config.Config, bagID string) (*models.SustainabilityData, error) {
	if cfg.AltumSustainabilityApiURL == "" {
		// Return default data when API is not configured (paid service)
		return &models.SustainabilityData{
			CurrentRating:       "Unknown",
			PotentialRating:     "Unknown",
			RecommendedMeasures: []models.SustainabilityMeasure{},
			TotalCO2Savings:     0,
			TotalCostSavings:    0,
			InvestmentCost:      0,
			PaybackPeriod:       0,
		}, nil
	}

	url := fmt.Sprintf("%s/sustainability/%s", cfg.AltumSustainabilityApiURL, bagID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if cfg.AltumSustainabilityApiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AltumSustainabilityApiKey))
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		// Return default data on network error
		return &models.SustainabilityData{
			CurrentRating:       "Unknown",
			PotentialRating:     "Unknown",
			RecommendedMeasures: []models.SustainabilityMeasure{},
			TotalCO2Savings:     0,
			TotalCostSavings:    0,
			InvestmentCost:      0,
			PaybackPeriod:       0,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("sustainability data not found for BAG ID: %s", bagID)
	}

	if resp.StatusCode != 200 {
		// Return default data on non-200 status
		return &models.SustainabilityData{
			CurrentRating:       "Unknown",
			PotentialRating:     "Unknown",
			RecommendedMeasures: []models.SustainabilityMeasure{},
			TotalCO2Savings:     0,
			TotalCostSavings:    0,
			InvestmentCost:      0,
			PaybackPeriod:       0,
		}, nil
	}

	var result models.SustainabilityData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// Return default data on parse error
		return &models.SustainabilityData{
			CurrentRating:       "Unknown",
			PotentialRating:     "Unknown",
			RecommendedMeasures: []models.SustainabilityMeasure{},
			TotalCO2Savings:     0,
			TotalCostSavings:    0,
			InvestmentCost:      0,
			PaybackPeriod:       0,
		}, nil
	}

	return &result, nil
}
