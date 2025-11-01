package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

type MonumentData struct {
	IsMonument bool
	Type       string
	Date       string
}

// Amsterdam Monuments API response structure
type amsterdamMonumentsResponse struct {
	Embedded struct {
		Monumenten []struct {
			Monumentnummer  int    `json:"monumentnummer"`
			Adressering     string `json:"adressering"`
			Type            string `json:"type"`
			Status          string `json:"status"` // "Rijksmonument" or "Gemeentelijk monument"
			DatumAanwijzing string `json:"datumAanwijzing"`
			BetreftBagPand  []struct {
				Identificatie string `json:"identificatie"` // BAG Pand ID
			} `json:"betreftBagPand"`
		} `json:"monumenten"`
	} `json:"_embedded"`
}

func (c *ApiClient) FetchMonumentData(cfg *config.Config, bagPandID string) (*MonumentData, error) {
	// Amsterdam Data API - query monuments by BAG pand ID
	// Note: This currently only works for Amsterdam addresses
	// For national coverage, would need to query RCE SPARQL endpoint
	url := fmt.Sprintf("%s?betreftBagPand.identificatie=%s", cfg.MonumentenApiURL, bagPandID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// Not found - not a monument
		return &MonumentData{
			IsMonument: false,
			Type:       "",
			Date:       "",
		}, nil
	}

	var result amsterdamMonumentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode monument response: %w", err)
	}

	if len(result.Embedded.Monumenten) == 0 {
		// No monuments found for this BAG pand ID
		return &MonumentData{
			IsMonument: false,
			Type:       "",
			Date:       "",
		}, nil
	}

	// Return first monument found
	monument := result.Embedded.Monumenten[0]
	return &MonumentData{
		IsMonument: true,
		Type:       monument.Status, // "Rijksmonument" or "Gemeentelijk monument"
		Date:       monument.DatumAanwijzing,
	}, nil
}
