package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

type CBSData struct {
	AvgIncome         float64
	PopulationDensity float64
	AvgWOZValue       float64
}

func (c *ApiClient) FetchCBSData(cfg *config.Config, neighborhoodCode string) (*CBSData, error) {
	fmt.Printf("[CBS] FetchCBSData: URL=%s, neighborhoodCode=%s\n", cfg.CBSApiURL, neighborhoodCode)
	// Return empty data if no CBS API URL is configured
	if cfg.CBSApiURL == "" {
		return &CBSData{
			AvgIncome:         0,
			PopulationDensity: 0,
			AvgWOZValue:       0,
		}, nil
	}

	// CBS OData API for neighborhood statistics (Kerncijfers wijken en buurten)
	// Using dataset 85039NED (most recent neighborhood stats)
	// Documentation: https://opendata.cbs.nl/statline/#/CBS/nl/dataset/85039NED
	url := fmt.Sprintf("%s/85039NED/Observations?$filter=RegioS eq '%s'&$orderby=Perioden desc&$top=1", cfg.CBSApiURL, neighborhoodCode)
	fmt.Printf("[CBS] Request URL: %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("[CBS] Request error: %v\n", err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		fmt.Printf("[CBS] HTTP error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("[CBS] Non-200 status: %d\n", resp.StatusCode)
		// Return default data if API fails
		return &CBSData{
			AvgIncome:         0,
			PopulationDensity: 0,
			AvgWOZValue:       0,
		}, nil
	}

	var result struct {
		Value []struct {
			// Updated field names for dataset 85039NED
			GemiddeldInkomen    float64 `json:"GemiddeldInkomenPerInkomensontvanger_68"`
			Bevolkingsdichtheid float64 `json:"Bevolkingsdichtheid_33"`
			GemiddeldeWOZ       float64 `json:"GemiddeldeWOZWaardeVanWoningen_35"`
		} `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("[CBS] Decode error: %v\n", err)
		return nil, err
	}
	fmt.Printf("[CBS] Response: %+v\n", result)
	if len(result.Value) == 0 {
		fmt.Printf("[CBS] No results for %s\n", neighborhoodCode)
		// Return empty data instead of error
		return &CBSData{
			AvgIncome:         0,
			PopulationDensity: 0,
			AvgWOZValue:       0,
		}, nil
	}
	data := &CBSData{
		AvgIncome:         result.Value[0].GemiddeldInkomen * 1000, // Convert from x1000 euro to euro
		PopulationDensity: result.Value[0].Bevolkingsdichtheid,
		AvgWOZValue:       result.Value[0].GemiddeldeWOZ * 1000, // Convert from x1000 euro to euro
	}
	fmt.Printf("[CBS] Final data: %+v\n", data)
	return data, nil
}
