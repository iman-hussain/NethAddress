package apiclient

import (
	"context"
	"fmt"

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

	headers := make(map[string]string)
	if cfg.AltumWOZApiKey != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", cfg.AltumWOZApiKey)
	}

	var result models.AltumWOZData
	if err := c.GetJSON(ctx, "Altum WOZ", url, headers, &result); err != nil {
		return nil, fmt.Errorf("altum WOZ API request failed: %w", err)
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

	headers := make(map[string]string)
	if cfg.AltumTransactionApiKey != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", cfg.AltumTransactionApiKey)
	}

	var result models.TransactionHistory
	if err := c.GetJSON(ctx, "Altum Transaction", url, headers, &result); err != nil {
		// Return empty transaction history on failure (soft failure for 404)
		return &models.TransactionHistory{Transactions: []models.TransactionData{}, TotalCount: 0}, nil
	}

	return &result, nil
}
