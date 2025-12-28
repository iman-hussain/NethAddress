package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// FetchAltumWOZData retrieves official WOZ tax values and building characteristics
// Documentation: https://docs.altum.ai/english/apis/woz-api
func (c *ApiClient) FetchAltumWOZData(ctx context.Context, cfg *config.Config, bagID string) (*models.AltumWOZData, error) {
	if cfg.AltumWOZApiURL == "" {
		return nil, fmt.Errorf("AltumWOZApiURL not configured")
	}

	url := fmt.Sprintf("%s/woz/%s", cfg.AltumWOZApiURL, bagID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if cfg.AltumWOZApiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AltumWOZApiKey))
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("WOZ data not found for BAG ID: %s", bagID)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("altum WOZ API returned status %d", resp.StatusCode)
	}

	var result models.AltumWOZData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode altum WOZ response: %w", err)
	}

	return &result, nil
}

// FetchTransactionHistory retrieves historical transactions from 1993-present
// Documentation: https://docs.altum.ai/english/apis/transaction-api
func (c *ApiClient) FetchTransactionHistory(ctx context.Context, cfg *config.Config, bagID string) (*models.TransactionHistory, error) {
	if cfg.AltumTransactionApiURL == "" {
		return nil, fmt.Errorf("AltumTransactionApiURL not configured")
	}

	url := fmt.Sprintf("%s/transactions/%s", cfg.AltumTransactionApiURL, bagID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if cfg.AltumTransactionApiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AltumTransactionApiKey))
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return &models.TransactionHistory{Transactions: []models.TransactionData{}, TotalCount: 0}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("altum transaction API returned status %d", resp.StatusCode)
	}

	var result models.TransactionHistory
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode transaction response: %w", err)
	}

	return &result, nil
}
