package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// FetchPDOKPlatformData retrieves comprehensive PDOK platform data
// Documentation: https://api.pdok.nl
func (c *ApiClient) FetchPDOKPlatformData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.PDOKPlatformData, error) {
	if cfg.PDOKApiURL == "" {
		return nil, fmt.Errorf("PDOKApiURL not configured")
	}

	url := fmt.Sprintf("%s/comprehensive?lat=%f&lon=%f", cfg.PDOKApiURL, lat, lon)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("PDOK platform API returned status %d", resp.StatusCode)
	}

	var result models.PDOKPlatformData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode PDOK platform response: %w", err)
	}

	return &result, nil
}

// FetchStratopoEnvironmentData retrieves 700+ environmental variables
// Documentation: https://stratopo.nl/en/environment-api/
func (c *ApiClient) FetchStratopoEnvironmentData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.StratopoEnvironmentData, error) {
	if cfg.StratopoApiURL == "" {
		return nil, fmt.Errorf("StratopoApiURL not configured")
	}

	url := fmt.Sprintf("%s/environment?lat=%f&lon=%f", cfg.StratopoApiURL, lat, lon)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if cfg.StratopoApiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.StratopoApiKey))
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("stratopo API returned status %d", resp.StatusCode)
	}

	var result models.StratopoEnvironmentData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode stratopo response: %w", err)
	}

	return &result, nil
}

// FetchLandUseData retrieves land use, zoning, and planning data
// Documentation: https://www.nationaalgeoregister.nl (CBS/PDOK)
func (c *ApiClient) FetchLandUseData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.LandUseData, error) {
	if cfg.LandUseApiURL == "" {
		return nil, fmt.Errorf("LandUseApiURL not configured")
	}

	url := fmt.Sprintf("%s/land-use?lat=%f&lon=%f", cfg.LandUseApiURL, lat, lon)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("land use API returned status %d", resp.StatusCode)
	}

	var result models.LandUseData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode land use response: %w", err)
	}

	return &result, nil
}
