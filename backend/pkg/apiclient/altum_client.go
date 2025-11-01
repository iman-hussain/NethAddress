package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

// AltumWOZData represents WOZ tax value data from Altum AI
type AltumWOZData struct {
	WOZValue     float64 `json:"wozValue"`
	ValueYear    int     `json:"valueYear"`
	BuildingType string  `json:"buildingType"`
	BuildYear    int     `json:"buildYear"`
	SurfaceArea  float64 `json:"surfaceArea"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

// FetchAltumWOZData retrieves official WOZ tax values and building characteristics
// Documentation: https://docs.altum.ai/english/apis/woz-api
func (c *ApiClient) FetchAltumWOZData(cfg *config.Config, bagID string) (*AltumWOZData, error) {
	if cfg.AltumWOZApiURL == "" {
		return nil, fmt.Errorf("AltumWOZApiURL not configured")
	}

	url := fmt.Sprintf("%s/woz/%s", cfg.AltumWOZApiURL, bagID)
	req, err := http.NewRequest("GET", url, nil)
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

	var result AltumWOZData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode altum WOZ response: %w", err)
	}

	return &result, nil
}

// TransactionData represents historical property transaction
type TransactionData struct {
	TransactionID string  `json:"transactionId"`
	Date          string  `json:"date"`
	PurchasePrice float64 `json:"purchasePrice"`
	PropertyType  string  `json:"propertyType"`
	SurfaceArea   float64 `json:"surfaceArea"`
	BAGObjectID   string  `json:"bagObjectId"`
}

// TransactionHistory contains list of transactions for a property
type TransactionHistory struct {
	Transactions []TransactionData `json:"transactions"`
	TotalCount   int               `json:"totalCount"`
}

// FetchTransactionHistory retrieves historical transactions from 1993-present
// Documentation: https://docs.altum.ai/english/apis/transaction-api
func (c *ApiClient) FetchTransactionHistory(cfg *config.Config, bagID string) (*TransactionHistory, error) {
	if cfg.AltumTransactionApiURL == "" {
		return nil, fmt.Errorf("AltumTransactionApiURL not configured")
	}

	url := fmt.Sprintf("%s/transactions/%s", cfg.AltumTransactionApiURL, bagID)
	req, err := http.NewRequest("GET", url, nil)
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
		return &TransactionHistory{Transactions: []TransactionData{}, TotalCount: 0}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("altum transaction API returned status %d", resp.StatusCode)
	}

	var result TransactionHistory
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode transaction response: %w", err)
	}

	return &result, nil
}
