package apiclient

import (
	"context"
	"fmt"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
	"github.com/iman-hussain/nethaddress/backend/pkg/models"
)

// FetchAltumWOZData retrieves official WOZ tax values and building characteristics
// Documentation: https://docs.altum.ai/english/apis/woz-api
func (c *ApiClient) FetchAltumWOZData(ctx context.Context, cfg *config.Config, bagID string) (*models.AltumWOZData, error) {
	if cfg.AltumWOZApiURL == "" {
		return nil, fmt.Errorf("AltumWOZApiURL not configured")
	}

	url := fmt.Sprintf("%s/woz/%s", cfg.AltumWOZApiURL, bagID)

	var result models.AltumWOZData
	if err := c.GetJSON(ctx, "Altum WOZ", url, BearerAuthHeader(cfg.AltumWOZApiKey), &result); err != nil {
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

	var result models.TransactionHistory
	if err := c.GetJSON(ctx, "Altum Transaction", url, BearerAuthHeader(cfg.AltumTransactionApiKey), &result); err != nil {
		// Return empty transaction history on failure (soft failure for 404)
		return &models.TransactionHistory{Transactions: []models.TransactionData{}, TotalCount: 0}, nil
	}

	return &result, nil
}
