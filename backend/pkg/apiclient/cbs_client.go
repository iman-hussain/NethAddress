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
	// CBS OData API for neighborhood statistics (Kerncijfers wijken en buurten 2018)
	// Dataset 84286NED - https://opendata.cbs.nl/ODataApi/odata/84286NED/WijkenEnBuurten
	url := fmt.Sprintf("%s/84286NED/WijkenEnBuurten?$filter=WijkenEnBuurten eq '%s'", cfg.CBSApiURL, neighborhoodCode)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			// GemiddeldInkomenPerInkomensontvanger_68: Average income per income recipient (x 1000 euro)
			// Bevolkingsdichtheid_33: Population density (inhabitants per km²)
			// GemiddeldeWOZWaardeVanWoningen_35: Average WOZ value of homes (x 1000 euro)
			GemiddeldInkomen    float64 `json:"GemiddeldInkomenPerInkomensontvanger_68"`
			Bevolkingsdichtheid float64 `json:"Bevolkingsdichtheid_33"`
			GemiddeldeWOZ       float64 `json:"GemiddeldeWOZWaardeVanWoningen_35"`
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Value) == 0 {
		return nil, fmt.Errorf("No CBS data found for neighborhood %s", neighborhoodCode)
	}
	data := &CBSData{
		AvgIncome:         result.Value[0].GemiddeldInkomen * 1000, // Convert from x1000 euro to euro
		PopulationDensity: result.Value[0].Bevolkingsdichtheid,
		AvgWOZValue:       result.Value[0].GemiddeldeWOZ * 1000, // Convert from x1000 euro to euro
	}
	return data, nil
}
