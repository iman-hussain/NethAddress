package apiclient

import (
	"context"
	"fmt"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
	"github.com/iman-hussain/nethaddress/backend/pkg/logutil"
	"github.com/iman-hussain/nethaddress/backend/pkg/models"
)

// Default PDOK RCE Monuments API endpoint (free, no auth required)
const defaultMonumentenApiURL = "https://api.pdok.nl/rce/beschermde-gebieden-cultuurhistorie/ogc/v1"

// emptyMonumentData returns a default MonumentData struct for soft failures.
func emptyMonumentData() *models.MonumentData {
	return &models.MonumentData{
		IsMonument: false,
		Type:       "",
		Date:       "",
	}
}

// FetchMonumentData queries the PDOK RCE monuments API to check if an address is a monument
// Uses bbox query with coordinates from BAG data
func (c *ApiClient) FetchMonumentData(ctx context.Context, cfg *config.Config, bagPandID string) (*models.MonumentData, error) {
	// This method uses BAG Pand ID - we'll also provide a coordinate-based fallback
	logutil.Debugf("[Monument] FetchMonumentData called with bagPandID: %s", bagPandID)

	// For now, return not a monument - the coordinate-based method is more reliable
	// The BAG Pand ID lookup requires Amsterdam's specific API
	return emptyMonumentData(), nil
}

// FetchMonumentDataByCoords queries PDOK RCE monuments API using coordinates
// Documentation: https://api.pdok.nl/rce/beschermde-gebieden-cultuurhistorie/ogc/v1
func (c *ApiClient) FetchMonumentDataByCoords(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.MonumentData, error) {
	// Use config URL if provided (for testing), otherwise use PDOK RCE API default
	baseURL := defaultMonumentenApiURL
	if cfg.MonumentenApiURL != "" {
		baseURL = cfg.MonumentenApiURL
	}

	// Create a small bounding box around the point (approximately 50m)
	delta := 0.0005
	bbox := fmt.Sprintf("%.6f,%.6f,%.6f,%.6f", lon-delta, lat-delta, lon+delta, lat+delta)

	url := fmt.Sprintf("%s/collections/rce_inspire_points/items?bbox=%s&f=json&limit=5", baseURL, bbox)
	logutil.Debugf("[Monument] Request URL: %s", url)

	var apiResp models.MonumentResponse
	if err := c.GetJSON(ctx, "Monument", url, nil, &apiResp); err != nil {
		return emptyMonumentData(), nil
	}

	logutil.Debugf("[Monument] Found %d monuments near coordinates", len(apiResp.Features))

	if len(apiResp.Features) == 0 {
		return emptyMonumentData(), nil
	}

	// Return the first monument found
	monument := apiResp.Features[0]
	result := &models.MonumentData{
		IsMonument: true,
		Type:       "Rijksmonument",
		Name:       monument.Properties.Text,
		Date:       monument.Properties.LegalFoundationDate,
	}

	logutil.Debugf("[Monument] Found: %s (registered %s)", result.Name, result.Date)
	return result, nil
}
