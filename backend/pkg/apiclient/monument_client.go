package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// Default PDOK RCE Monuments API endpoint (free, no auth required)
const defaultMonumentenApiURL = "https://api.pdok.nl/rce/beschermde-gebieden-cultuurhistorie/ogc/v1"

// FetchMonumentData queries the PDOK RCE monuments API to check if an address is a monument
// Uses bbox query with coordinates from BAG data
func (c *ApiClient) FetchMonumentData(ctx context.Context, cfg *config.Config, bagPandID string) (*models.MonumentData, error) {
	// This method uses BAG Pand ID - we'll also provide a coordinate-based fallback
	logutil.Debugf("[Monument] FetchMonumentData called with bagPandID: %s", bagPandID)

	// For now, return not a monument - the coordinate-based method is more reliable
	// The BAG Pand ID lookup requires Amsterdam's specific API
	return &models.MonumentData{
		IsMonument: false,
		Type:       "",
		Date:       "",
	}, nil
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

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logutil.Debugf("[Monument] Request error: %v", err)
		return &models.MonumentData{IsMonument: false}, nil
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[Monument] HTTP error: %v", err)
		return &models.MonumentData{IsMonument: false}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logutil.Debugf("[Monument] Non-200 status: %d", resp.StatusCode)
		return &models.MonumentData{IsMonument: false}, nil
	}

	var apiResp models.MonumentResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[Monument] Decode error: %v", err)
		return &models.MonumentData{IsMonument: false}, nil
	}

	logutil.Debugf("[Monument] Found %d monuments near coordinates", len(apiResp.Features))

	if len(apiResp.Features) == 0 {
		return &models.MonumentData{
			IsMonument: false,
			Type:       "",
			Date:       "",
		}, nil
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
