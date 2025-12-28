package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// FetchKadasterObjectInfo retrieves comprehensive property information using BAG ID
// Documentation: https://www.kadaster.nl/-/objectinformatie-api
func (c *ApiClient) FetchKadasterObjectInfo(ctx context.Context, cfg *config.Config, bagID string) (*models.KadasterObjectInfo, error) {
	if cfg.KadasterObjectInfoApiURL == "" {
		return nil, fmt.Errorf("KadasterObjectInfoApiURL not configured")
	}

	url := fmt.Sprintf("%s/objecten/%s", cfg.KadasterObjectInfoApiURL, bagID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add API key if provided
	if cfg.KadasterObjectInfoApiKey != "" {
		req.Header.Set("X-Api-Key", cfg.KadasterObjectInfoApiKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("property not found for BAG ID: %s", bagID)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("kadaster API returned status %d", resp.StatusCode)
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

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode kadaster response: %w", err)
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
