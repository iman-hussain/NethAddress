package apiclient

import (
	"context"
	"fmt"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
	"github.com/iman-hussain/nethaddress/backend/pkg/models"
)

// FetchPDOKPlatformData retrieves comprehensive PDOK platform data
// Documentation: https://api.pdok.nl
func (c *ApiClient) FetchPDOKPlatformData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.PDOKPlatformData, error) {
	if cfg.PDOKApiURL == "" {
		return nil, fmt.Errorf("PDOKApiURL not configured")
	}

	url := fmt.Sprintf("%s/comprehensive?lat=%f&lon=%f", cfg.PDOKApiURL, lat, lon)

	var result models.PDOKPlatformData
	if err := c.GetJSON(ctx, "PDOK Platform", url, nil, &result); err != nil {
		return nil, fmt.Errorf("PDOK platform API request failed: %w", err)
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

	var result models.StratopoEnvironmentData
	if err := c.GetJSON(ctx, "Stratopo", url, BearerAuthHeader(cfg.StratopoApiKey), &result); err != nil {
		return nil, fmt.Errorf("stratopo API request failed: %w", err)
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

	var result models.LandUseData
	if err := c.GetJSON(ctx, "Land Use", url, nil, &result); err != nil {
		return nil, fmt.Errorf("land use API request failed: %w", err)
	}

	return &result, nil
}
