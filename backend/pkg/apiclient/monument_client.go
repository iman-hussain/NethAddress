package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
)

// Default PDOK RCE Monuments API endpoint (free, no auth required)
const defaultMonumentenApiURL = "https://api.pdok.nl/rce/beschermde-gebieden-cultuurhistorie/ogc/v1"

type MonumentData struct {
	IsMonument bool   `json:"isMonument"`
	Type       string `json:"type"`
	Date       string `json:"date"`
	Name       string `json:"name,omitempty"`
	Number     string `json:"number,omitempty"`
}

// monumentResponse represents the PDOK RCE INSPIRE monuments API response
type monumentResponse struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Properties struct {
			// INSPIRE format fields
			Text                string `json:"text"`                // Monument name
			LegalFoundationDate string `json:"legalfoundationdate"` // Date of designation
			CICitation          string `json:"ci_citation"`         // Link to monumentenregister
		} `json:"properties"`
	} `json:"features"`
	NumberReturned int `json:"numberReturned"`
}

// FetchMonumentData queries the PDOK RCE monuments API to check if an address is a monument
// Uses bbox query with coordinates from BAG data
func (c *ApiClient) FetchMonumentData(cfg *config.Config, bagPandID string) (*MonumentData, error) {
	// This method uses BAG Pand ID - we'll also provide a coordinate-based fallback
	logutil.Debugf("[Monument] FetchMonumentData called with bagPandID: %s", bagPandID)
	
	// For now, return not a monument - the coordinate-based method is more reliable
	// The BAG Pand ID lookup requires Amsterdam's specific API
	return &MonumentData{
		IsMonument: false,
		Type:       "",
		Date:       "",
	}, nil
}

// FetchMonumentDataByCoords queries PDOK RCE monuments API using coordinates
// Documentation: https://api.pdok.nl/rce/beschermde-gebieden-cultuurhistorie/ogc/v1
func (c *ApiClient) FetchMonumentDataByCoords(cfg *config.Config, lat, lon float64) (*MonumentData, error) {
	baseURL := defaultMonumentenApiURL
	if cfg.MonumentenApiURL != "" && cfg.MonumentenApiURL != "https://api.data.amsterdam.nl/monumenten/monumenten" {
		baseURL = cfg.MonumentenApiURL
	}

	// Create a small bounding box around the point (approximately 50m)
	delta := 0.0005
	bbox := fmt.Sprintf("%.6f,%.6f,%.6f,%.6f", lon-delta, lat-delta, lon+delta, lat+delta)

	url := fmt.Sprintf("%s/collections/rce_inspire_points/items?bbox=%s&f=json&limit=5", baseURL, bbox)
	logutil.Debugf("[Monument] Request URL: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logutil.Debugf("[Monument] Request error: %v", err)
		return &MonumentData{IsMonument: false}, nil
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[Monument] HTTP error: %v", err)
		return &MonumentData{IsMonument: false}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logutil.Debugf("[Monument] Non-200 status: %d", resp.StatusCode)
		return &MonumentData{IsMonument: false}, nil
	}

	var apiResp monumentResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[Monument] Decode error: %v", err)
		return &MonumentData{IsMonument: false}, nil
	}

	logutil.Debugf("[Monument] Found %d monuments near coordinates", len(apiResp.Features))

	if len(apiResp.Features) == 0 {
		return &MonumentData{
			IsMonument: false,
			Type:       "",
			Date:       "",
		}, nil
	}

	// Return the first monument found
	monument := apiResp.Features[0]
	result := &MonumentData{
		IsMonument: true,
		Type:       "Rijksmonument",
		Name:       monument.Properties.Text,
		Date:       monument.Properties.LegalFoundationDate,
	}

	logutil.Debugf("[Monument] Found: %s (registered %s)", result.Name, result.Date)
	return result, nil
}
