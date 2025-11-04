package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

// CBSPopulationData represents grid-based population data
type CBSPopulationData struct {
	TotalPopulation int                    `json:"totalPopulation"`
	AgeDistribution map[string]int         `json:"ageDistribution"`
	Households      int                    `json:"households"`
	AverageHHSize   float64                `json:"averageHouseholdSize"`
	Demographics    PopulationDemographics `json:"demographics"`
}

// PopulationDemographics represents demographic breakdown
type PopulationDemographics struct {
	Age0to14  int `json:"age0to14"`
	Age15to24 int `json:"age15to24"`
	Age25to44 int `json:"age25to44"`
	Age45to64 int `json:"age45to64"`
	Age65Plus int `json:"age65plus"`
}

// FetchCBSPopulationData retrieves grid population data for target market analysis
// Documentation: https://api.pdok.nl/cbs/population-distribution
func (c *ApiClient) FetchCBSPopulationData(cfg *config.Config, lat, lon float64) (*CBSPopulationData, error) {
	if cfg.CBSPopulationApiURL == "" {
		return &CBSPopulationData{
			TotalPopulation: 0,
			AgeDistribution: make(map[string]int),
			Households:      0,
			AverageHHSize:   0,
			Demographics: PopulationDemographics{
				Age0to14:  0,
				Age15to24: 0,
				Age25to44: 0,
				Age45to64: 0,
				Age65Plus: 0,
			},
		}, nil
	}

	// Note: This would need WFS query implementation for actual PDOK service
	// For now, return empty data gracefully
	return &CBSPopulationData{
		TotalPopulation: 0,
		AgeDistribution: make(map[string]int),
		Households:      0,
		AverageHHSize:   0,
		Demographics: PopulationDemographics{
			Age0to14:  0,
			Age15to24: 0,
			Age25to44: 0,
			Age45to64: 0,
			Age65Plus: 0,
		},
	}, nil
}

// CBSStatLineData represents comprehensive socioeconomic data
type CBSStatLineData struct {
	RegionCode     string  `json:"regionCode"`
	RegionName     string  `json:"regionName"`
	Population     int     `json:"population"`
	AverageIncome  float64 `json:"averageIncome"`  // EUR per household
	EmploymentRate float64 `json:"employmentRate"` // percentage
	EducationLevel string  `json:"educationLevel"` // Low, Medium, High
	HousingStock   int     `json:"housingStock"`
	AverageWOZ     float64 `json:"averageWOZ"`
	Year           int     `json:"year"`
}

// FetchCBSStatLineData retrieves socioeconomic data via OData API
// Documentation: https://www.cbs.nl/en-gb/our-services/open-data/statline-as-open-data
func (c *ApiClient) FetchCBSStatLineData(cfg *config.Config, regionCode string) (*CBSStatLineData, error) {
	if cfg.CBSStatLineApiURL == "" {
		return &CBSStatLineData{
			RegionCode:     regionCode,
			RegionName:     regionCode,
			Population:     0,
			AverageIncome:  0,
			EmploymentRate: 0,
			EducationLevel: "Unknown",
			HousingStock:   0,
			AverageWOZ:     0,
			Year:           2024,
		}, nil
	}

	// CBS OData API endpoint for regional statistics
	// Using dataset 85039NED (Kerncijfers wijken en buurten 2023)
	url := fmt.Sprintf("%s/odata/85039NED/Observations?$filter=RegioS eq '%s'&$orderby=Perioden desc&$top=1",
		cfg.CBSStatLineApiURL, regionCode)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Return default data if API fails
		return &CBSStatLineData{
			RegionCode:     regionCode,
			RegionName:     regionCode,
			Population:     0,
			AverageIncome:  0,
			EmploymentRate: 0,
			EducationLevel: "Unknown",
			HousingStock:   0,
			AverageWOZ:     0,
			Year:           2024,
		}, nil
	}

	var result struct {
		Value []struct {
			RegioS                               string  `json:"RegioS"`
			Perioden                             string  `json:"Perioden"`
			BevolkingAanHetBeginVanDePeriode_1   int     `json:"BevolkingAanHetBeginVanDePeriode_1"`
			GemiddeldInkomenPerInwoner_66        float64 `json:"GemiddeldInkomenPerInwoner_66"`
			PercentageWerkloosPerLeeftijdsklasse float64 `json:"PercentageWerkloosPerLeeftijdsklasse"`
			GemiddeldeWOZWaardeVanWoningen_35    float64 `json:"GemiddeldeWOZWaardeVanWoningen_35"`
			Woningvoorraad_31                    int     `json:"Woningvoorraad_31"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode CBS StatLine response: %w", err)
	}

	if len(result.Value) == 0 {
		// Return default data instead of error
		return &CBSStatLineData{
			RegionCode:     regionCode,
			RegionName:     regionCode,
			Population:     0,
			AverageIncome:  0,
			EmploymentRate: 0,
			EducationLevel: "Unknown",
			HousingStock:   0,
			AverageWOZ:     0,
			Year:           2024,
		}, nil
	}

	data := result.Value[0]
	return &CBSStatLineData{
		RegionCode:     data.RegioS,
		RegionName:     regionCode, // Would need lookup table for full names
		Population:     data.BevolkingAanHetBeginVanDePeriode_1,
		AverageIncome:  data.GemiddeldInkomenPerInwoner_66,
		EmploymentRate: 100.0 - data.PercentageWerkloosPerLeeftijdsklasse,
		HousingStock:   data.Woningvoorraad_31,
		AverageWOZ:     data.GemiddeldeWOZWaardeVanWoningen_35 * 1000, // Convert from k EUR
		Year:           2024,                                          // Parse from Perioden field
	}, nil
}

// CBSSquareStatsData represents hyperlocal 100x100m grid statistics
type CBSSquareStatsData struct {
	GridID         string  `json:"gridId"`
	Population     int     `json:"population"`
	Households     int     `json:"households"`
	AverageWOZ     float64 `json:"averageWOZ"`
	AverageIncome  float64 `json:"averageIncome"`
	HousingDensity int     `json:"housingDensity"` // units per hectare
}

// FetchCBSSquareStats retrieves 100x100m microgrid statistics
// Documentation: https://api.store (CBS Square Statistics)
func (c *ApiClient) FetchCBSSquareStats(cfg *config.Config, lat, lon float64) (*CBSSquareStatsData, error) {
	if cfg.CBSSquareStatsApiURL == "" {
		return &CBSSquareStatsData{
			GridID:         "",
			Population:     0,
			Households:     0,
			AverageWOZ:     0,
			AverageIncome:  0,
			HousingDensity: 0,
		}, nil
	}

	// Note: This would need WFS query implementation for actual PDOK service
	// For now, return empty data gracefully
	return &CBSSquareStatsData{
		GridID:         "",
		Population:     0,
		Households:     0,
		AverageWOZ:     0,
		AverageIncome:  0,
		HousingDensity: 0,
	}, nil
}
