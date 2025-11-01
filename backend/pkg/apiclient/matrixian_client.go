package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

// MatrixianPropertyValue represents comprehensive market valuation data
type MatrixianPropertyValue struct {
	MarketValue          float64                `json:"marketValue"`
	ValuationDate        string                 `json:"valuationDate"`
	Confidence           float64                `json:"confidence"`
	ComparableProperties []ComparableProperty   `json:"comparableProperties"`
	Features             map[string]interface{} `json:"features"`
	PricePerSqm          float64                `json:"pricePerSqm"`
}

// ComparableProperty represents similar properties in the area
type ComparableProperty struct {
	Address      string  `json:"address"`
	Distance     float64 `json:"distance"`
	SalePrice    float64 `json:"salePrice"`
	SaleDate     string  `json:"saleDate"`
	SurfaceArea  float64 `json:"surfaceArea"`
	PropertyType string  `json:"propertyType"`
}

// FetchPropertyValuePlus retrieves market value and 30+ features
// Documentation: https://matrixian.com/en/api/property-value-plus-api/
func (c *ApiClient) FetchPropertyValuePlus(cfg *config.Config, bagID string, lat, lon float64) (*MatrixianPropertyValue, error) {
	if cfg.MatrixianApiURL == "" {
		return nil, fmt.Errorf("MatrixianApiURL not configured")
	}

	url := fmt.Sprintf("%s/property-value-plus?bagId=%s&lat=%f&lon=%f", cfg.MatrixianApiURL, bagID, lat, lon)
	req, err := http.NewRequest("GET", url, nil)
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

	var result MatrixianPropertyValue
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode matrixian response: %w", err)
	}

	return &result, nil
}
