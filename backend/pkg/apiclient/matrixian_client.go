package apiclient

import (
	"context"
	"fmt"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
	"github.com/iman-hussain/nethaddress/backend/pkg/models"
)

// FetchPropertyValuePlus retrieves market value and 30+ features
// Documentation: https://matrixian.com/en/api/property-value-plus-api/
func (c *ApiClient) FetchPropertyValuePlus(ctx context.Context, cfg *config.Config, bagID string, lat, lon float64) (*models.MatrixianPropertyValue, error) {
	if cfg.MatrixianApiURL == "" {
		return nil, fmt.Errorf("MatrixianApiURL not configured")
	}

	url := fmt.Sprintf("%s/property-value-plus?bagId=%s&lat=%f&lon=%f", cfg.MatrixianApiURL, bagID, lat, lon)

	headers := make(map[string]string)
	if cfg.MatrixianApiKey != "" {
		headers["X-API-Key"] = cfg.MatrixianApiKey
	}

	var result models.MatrixianPropertyValue
	if err := c.GetJSON(ctx, "Matrixian", url, headers, &result); err != nil {
		return nil, fmt.Errorf("matrixian API request failed: %w", err)
	}

	return &result, nil
}
