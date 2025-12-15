package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
)

// Default PDOK CBS API endpoint (free, no auth required)
const defaultCBSBuurtenApiURL = "https://api.pdok.nl/cbs/wijken-en-buurten-2024/ogc/v1"

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

// cbsBuurtenResponse represents the PDOK CBS wijken-en-buurten OGC API response
type cbsBuurtenResponse struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Properties struct {
			Buurtcode                    string   `json:"buurtcode"`
			Buurtnaam                    string   `json:"buurtnaam"`
			Wijkcode                     string   `json:"wijkcode"`
			Gemeentecode                 string   `json:"gemeentecode"`
			Gemeentenaam                 string   `json:"gemeentenaam"`
			AantalInwoners               *int     `json:"aantal_inwoners"`
			AantalHuishoudens            *int     `json:"aantal_huishoudens"`
			GemiddeldeHuishoudensgrootte *float64 `json:"gemiddelde_huishoudsgrootte"`
			Bevolkingsdichtheid          *int     `json:"bevolkingsdichtheid_inwoners_per_km2"`
			// Age distribution fields (percentage)
			Perc0Tot15Jaar  *int `json:"percentage_personen_0_tot_15_jaar"`
			Perc15Tot25Jaar *int `json:"percentage_personen_15_tot_25_jaar"`
			Perc25Tot45Jaar *int `json:"percentage_personen_25_tot_45_jaar"`
			Perc45Tot65Jaar *int `json:"percentage_personen_45_tot_65_jaar"`
			Perc65Plus      *int `json:"percentage_personen_65_jaar_en_ouder"`
		} `json:"properties"`
	} `json:"features"`
	NumberReturned int `json:"numberReturned"`
}

// FetchCBSPopulationData retrieves population data using the free PDOK CBS OGC API
// Documentation: https://api.pdok.nl/cbs/wijken-en-buurten-2024/ogc/v1
func (c *ApiClient) FetchCBSPopulationData(cfg *config.Config, lat, lon float64) (*CBSPopulationData, error) {
	// Always use PDOK CBS API default (free, no auth) - ignore config overrides which may have bad URLs
	baseURL := defaultCBSBuurtenApiURL

	// Create a small bounding box around the point (approximately 200m)
	delta := 0.001 // ~100m in latitude
	bbox := fmt.Sprintf("%.6f,%.6f,%.6f,%.6f", lon-delta, lat-delta, lon+delta, lat+delta)

	url := fmt.Sprintf("%s/collections/buurten/items?bbox=%s&f=json&limit=1", baseURL, bbox)
	logutil.Debugf("[CBS Population] Request URL: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logutil.Debugf("[CBS Population] Request error: %v", err)
		return emptyPopulationData(), nil
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[CBS Population] HTTP error: %v", err)
		return emptyPopulationData(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logutil.Debugf("[CBS Population] Non-200 status: %d", resp.StatusCode)
		return emptyPopulationData(), nil
	}

	var apiResp cbsBuurtenResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[CBS Population] Decode error: %v", err)
		return emptyPopulationData(), nil
	}

	if len(apiResp.Features) == 0 {
		logutil.Debugf("[CBS Population] No features found for coordinates")
		return emptyPopulationData(), nil
	}

	props := apiResp.Features[0].Properties
	logutil.Debugf("[CBS Population] Found buurt: %s (%s)", props.Buurtnaam, props.Buurtcode)

	// Extract population data, handling nil pointers and negative CBS sentinel values
	population := 0
	if props.AantalInwoners != nil && *props.AantalInwoners >= 0 {
		population = *props.AantalInwoners
	}

	households := 0
	if props.AantalHuishoudens != nil && *props.AantalHuishoudens >= 0 {
		households = *props.AantalHuishoudens
	}

	avgHHSize := 0.0
	if props.GemiddeldeHuishoudensgrootte != nil && *props.GemiddeldeHuishoudensgrootte >= 0 {
		avgHHSize = *props.GemiddeldeHuishoudensgrootte
	}

	// Calculate age distribution from percentages
	// CBS stores values as percentage (e.g., 15 = 15%)
	age0to14 := 0
	age15to24 := 0
	age25to44 := 0
	age45to64 := 0
	age65plus := 0

	if population > 0 {
		if props.Perc0Tot15Jaar != nil {
			age0to14 = (population * (*props.Perc0Tot15Jaar)) / 100
		}
		if props.Perc15Tot25Jaar != nil {
			age15to24 = (population * (*props.Perc15Tot25Jaar)) / 100
		}
		if props.Perc25Tot45Jaar != nil {
			age25to44 = (population * (*props.Perc25Tot45Jaar)) / 100
		}
		if props.Perc45Tot65Jaar != nil {
			age45to64 = (population * (*props.Perc45Tot65Jaar)) / 100
		}
		if props.Perc65Plus != nil {
			age65plus = (population * (*props.Perc65Plus)) / 100
		}
	}

	result := &CBSPopulationData{
		TotalPopulation: population,
		AgeDistribution: map[string]int{
			"0-14":  age0to14,
			"15-24": age15to24,
			"25-44": age25to44,
			"45-64": age45to64,
			"65+":   age65plus,
		},
		Households:    households,
		AverageHHSize: avgHHSize,
		Demographics: PopulationDemographics{
			Age0to14:  age0to14,
			Age15to24: age15to24,
			Age25to44: age25to44,
			Age45to64: age45to64,
			Age65Plus: age65plus,
		},
	}

	logutil.Debugf("[CBS Population] Result: pop=%d, households=%d", population, households)
	return result, nil
}

func emptyPopulationData() *CBSPopulationData {
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
	}
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
	// Using dataset 84286NED for compatibility
	url := fmt.Sprintf("%s/ODataFeed/v4/CBS/84286NED/Observations?$filter=RegioS eq '%s'&$orderby=Perioden desc&$top=1",
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

// CBSSquareStatsData represents hyperlocal neighbourhood statistics
type CBSSquareStatsData struct {
	GridID         string  `json:"gridId"`
	Population     int     `json:"population"`
	Households     int     `json:"households"`
	AverageWOZ     float64 `json:"averageWOZ"`
	AverageIncome  float64 `json:"averageIncome"`
	HousingDensity int     `json:"housingDensity"` // units per hectare
}

// cbsSquareResponse for parsing CBS grid statistics
type cbsSquareResponse struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Properties struct {
			Buurtcode                  string `json:"buurtcode"`
			AantalInwoners             *int   `json:"aantal_inwoners"`
			AantalHuishoudens          *int   `json:"aantal_huishoudens"`
			GemiddeldeWozWaardeWoning  *int   `json:"gemiddelde_woningwaarde"`
			Omgevingsadressendichtheid *int   `json:"omgevingsadressendichtheid"`
			// Income data
			GemHuishoudinkomen *int `json:"gemiddeld_gestandaardiseerd_inkomen_van_huishoudens"`
		} `json:"properties"`
	} `json:"features"`
}

// FetchCBSSquareStats retrieves neighbourhood-level statistics using PDOK CBS API
// Documentation: https://api.pdok.nl/cbs/wijken-en-buurten-2024/ogc/v1
func (c *ApiClient) FetchCBSSquareStats(cfg *config.Config, lat, lon float64) (*CBSSquareStatsData, error) {
	// Use config URL if provided (for testing), otherwise use PDOK CBS API default
	baseURL := defaultCBSBuurtenApiURL
	if cfg.CBSSquareStatsApiURL != "" {
		baseURL = cfg.CBSSquareStatsApiURL
	}

	// Create a small bounding box around the point
	delta := 0.001
	bbox := fmt.Sprintf("%.6f,%.6f,%.6f,%.6f", lon-delta, lat-delta, lon+delta, lat+delta)

	url := fmt.Sprintf("%s/collections/buurten/items?bbox=%s&f=json&limit=1", baseURL, bbox)
	logutil.Debugf("[CBS Square] Request URL: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logutil.Debugf("[CBS Square] Request error: %v", err)
		return emptySquareStats(), nil
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[CBS Square] HTTP error: %v", err)
		return emptySquareStats(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logutil.Debugf("[CBS Square] Non-200 status: %d", resp.StatusCode)
		return emptySquareStats(), nil
	}

	var apiResp cbsSquareResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		logutil.Debugf("[CBS Square] Decode error: %v", err)
		return emptySquareStats(), nil
	}

	if len(apiResp.Features) == 0 {
		logutil.Debugf("[CBS Square] No features found")
		return emptySquareStats(), nil
	}

	props := apiResp.Features[0].Properties

	population := 0
	if props.AantalInwoners != nil && *props.AantalInwoners >= 0 {
		population = *props.AantalInwoners
	}

	households := 0
	if props.AantalHuishoudens != nil && *props.AantalHuishoudens >= 0 {
		households = *props.AantalHuishoudens
	}

	avgWOZ := 0.0
	if props.GemiddeldeWozWaardeWoning != nil && *props.GemiddeldeWozWaardeWoning >= 0 {
		avgWOZ = float64(*props.GemiddeldeWozWaardeWoning) // Already in EUR
	}

	density := 0
	if props.Omgevingsadressendichtheid != nil && *props.Omgevingsadressendichtheid >= 0 {
		density = *props.Omgevingsadressendichtheid
	}

	avgIncome := 0.0
	if props.GemHuishoudinkomen != nil && *props.GemHuishoudinkomen >= 0 {
		avgIncome = float64(*props.GemHuishoudinkomen) * 100 // Convert from x100
	}

	result := &CBSSquareStatsData{
		GridID:         props.Buurtcode,
		Population:     population,
		Households:     households,
		AverageWOZ:     avgWOZ,
		AverageIncome:  avgIncome,
		HousingDensity: density,
	}

	logutil.Debugf("[CBS Square] Result: grid=%s, pop=%d, woz=%.0f", props.Buurtcode, population, avgWOZ)
	return result, nil
}

func emptySquareStats() *CBSSquareStatsData {
	return &CBSSquareStatsData{
		GridID:         "",
		Population:     0,
		Households:     0,
		AverageWOZ:     0,
		AverageIncome:  0,
		HousingDensity: 0,
	}
}
