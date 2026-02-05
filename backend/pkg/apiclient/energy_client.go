package apiclient

import (
	"context"
	"fmt"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// emptyEnergyClimateData returns a default EnergyClimateData struct for soft failures.
func emptyEnergyClimateData() *models.EnergyClimateData {
	return &models.EnergyClimateData{
		EnergyLabel:      "Unknown",
		ClimateRisk:      "Unknown",
		EfficiencyScore:  0,
		AnnualEnergyCost: 0,
		CO2Emissions:     0,
		HeatLoss:         0,
	}
}

// FetchEnergyClimateData retrieves energy labels and climate risk for ESG scoring
// Documentation: https://docs.altum.ai/english/apis/energy-and-climate-api
func (c *ApiClient) FetchEnergyClimateData(ctx context.Context, cfg *config.Config, bagID string) (*models.EnergyClimateData, error) {
	if cfg.AltumEnergyApiURL == "" {
		// Return default data when API is not configured (paid service)
		return emptyEnergyClimateData(), nil
	}

	url := fmt.Sprintf("%s/energy/%s", cfg.AltumEnergyApiURL, bagID)

	var result models.EnergyClimateData
	if err := c.GetJSON(ctx, "Energy Climate", url, BearerAuthHeader(cfg.AltumEnergyApiKey), &result); err != nil {
		return emptyEnergyClimateData(), nil
	}

	return &result, nil
}

// emptySustainabilityData returns a default SustainabilityData struct for soft failures.
func emptySustainabilityData() *models.SustainabilityData {
	return &models.SustainabilityData{
		CurrentRating:       "Unknown",
		PotentialRating:     "Unknown",
		RecommendedMeasures: []models.SustainabilityMeasure{},
		TotalCO2Savings:     0,
		TotalCostSavings:    0,
		InvestmentCost:      0,
		PaybackPeriod:       0,
	}
}

// FetchSustainabilityData retrieves sustainability recommendations and potential savings
// Documentation: https://docs.altum.ai/english/apis/sustainability-api
func (c *ApiClient) FetchSustainabilityData(ctx context.Context, cfg *config.Config, bagID string) (*models.SustainabilityData, error) {
	if cfg.AltumSustainabilityApiURL == "" {
		// Return default data when API is not configured (paid service)
		return emptySustainabilityData(), nil
	}

	url := fmt.Sprintf("%s/sustainability/%s", cfg.AltumSustainabilityApiURL, bagID)

	var result models.SustainabilityData
	if err := c.GetJSON(ctx, "Sustainability", url, BearerAuthHeader(cfg.AltumSustainabilityApiKey), &result); err != nil {
		return emptySustainabilityData(), nil
	}

	return &result, nil
}
