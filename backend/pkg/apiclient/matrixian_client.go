package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// FetchPropertyValuePlus retrieves market value and 30+ features
// Documentation: https://matrixian.com/en/api/property-value-plus-api/
func (c *ApiClient) FetchPropertyValuePlus(ctx context.Context, cfg *config.Config, bagID string, lat, lon float64) (*models.MatrixianPropertyValue, error) {
	if cfg.MatrixianApiURL == "" {
		return nil, fmt.Errorf("MatrixianApiURL not configured")
	}

	url := fmt.Sprintf("%s/property-value-plus?bagId=%s&lat=%f&lon=%f", cfg.MatrixianApiURL, bagID, lat, lon)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if cfg.MatrixianApiKey != "" {
		req.Header.Set("X-API-Key", cfg.MatrixianApiKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("property value data not found for BAG ID: %s", bagID)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("matrixian API returned status %d", resp.StatusCode)
	}

	var result models.MatrixianPropertyValue
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode matrixian response: %w", err)
	}

	return &result, nil
}
