package apiclient

import (
	"context"
	"fmt"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
	"github.com/iman-hussain/nethaddress/backend/pkg/models"
)

// FetchKadasterObjectInfo retrieves comprehensive property information using BAG ID
// Documentation: https://www.kadaster.nl/-/objectinformatie-api
func (c *ApiClient) FetchKadasterObjectInfo(ctx context.Context, cfg *config.Config, bagID string) (*models.KadasterObjectInfo, error) {
	if cfg.KadasterObjectInfoApiURL == "" {
		return nil, fmt.Errorf("KadasterObjectInfoApiURL not configured")
	}

	url := fmt.Sprintf("%s/objecten/%s", cfg.KadasterObjectInfoApiURL, bagID)

	// Build custom headers for API key authentication
	headers := make(map[string]string)
	if cfg.KadasterObjectInfoApiKey != "" {
		headers["X-Api-Key"] = cfg.KadasterObjectInfoApiKey
	}

	var result struct {
		Eigenaar struct {
			Naam string `json:"naam"`
		} `json:"eigenaar"`
		Kadaster struct {
			Referentie string `json:"referentie"`
		} `json:"kadaster"`
		WOZ struct {
			Waarde float64 `json:"waarde"`
		} `json:"woz"`
		Energie struct {
			Label string `json:"label"`
		} `json:"energie"`
		Belastingen struct {
			Gemeentelijk float64 `json:"gemeentelijk"`
		} `json:"belastingen"`
		Oppervlakte struct {
			Wonen   float64 `json:"wonen"`
			Perceel float64 `json:"perceel"`
		} `json:"oppervlakte"`
		Gebouw struct {
			Type     string `json:"type"`
			Bouwjaar int    `json:"bouwjaar"`
		} `json:"gebouw"`
	}

	if err := c.GetJSON(ctx, "Kadaster", url, headers, &result); err != nil {
		return nil, fmt.Errorf("kadaster API request failed: %w", err)
	}

	return &models.KadasterObjectInfo{
		OwnerName:          result.Eigenaar.Naam,
		CadastralReference: result.Kadaster.Referentie,
		WOZValue:           result.WOZ.Waarde,
		EnergyLabel:        result.Energie.Label,
		MunicipalTaxes:     result.Belastingen.Gemeentelijk,
		SurfaceArea:        result.Oppervlakte.Wonen,
		PlotSize:           result.Oppervlakte.Perceel,
		BuildingType:       result.Gebouw.Type,
		BuildYear:          result.Gebouw.Bouwjaar,
	}, nil
}
