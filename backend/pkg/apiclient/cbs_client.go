package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"

	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
)

type CBSData struct {
	AvgIncome         float64
	PopulationDensity float64
	AvgWOZValue       float64
}

func (c *ApiClient) FetchCBSData(cfg *config.Config, neighborhoodCode string) (*CBSData, error) {
	logutil.Debugf("[CBS] FetchCBSData: URL=%s, neighborhoodCode=%s", cfg.CBSApiURL, neighborhoodCode)
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
	logutil.Debugf("[CBS] Request URL: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logutil.Debugf("[CBS] Request error: %v", err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[CBS] HTTP error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logutil.Debugf("[CBS] Non-200 status: %d", resp.StatusCode)
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
		logutil.Debugf("[CBS] Decode error: %v", err)
		return nil, err
	}
	logutil.Debugf("[CBS] Response: %+v", result)
	if len(result.Value) == 0 {
		logutil.Debugf("[CBS] No results for %s", neighborhoodCode)
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
	logutil.Debugf("[CBS] Final data: %+v", data)
	return data, nil
}

// NeighborhoodRegionCodes contains CBS region codes for a location
type NeighborhoodRegionCodes struct {
	NeighborhoodCode string `json:"neighborhoodCode"` // Buurt code (BU...)
	DistrictCode     string `json:"districtCode"`     // Wijk code (WK...)
	MunicipalityCode string `json:"municipalityCode"` // Gemeente code (GM...)
	NeighborhoodName string `json:"neighborhoodName"`
	DistrictName     string `json:"districtName"`
	MunicipalityName string `json:"municipalityName"`
}

// LookupNeighborhoodCode attempts to find the CBS neighborhood code for given coordinates
// It uses PDOK WFS service to query CBS neighborhoods (buurten) data
// Falls back to municipality code if neighborhood lookup fails
func (c *ApiClient) LookupNeighborhoodCode(cfg *config.Config, lat, lon float64) (*NeighborhoodRegionCodes, error) {
	logutil.Debugf("[CBS] LookupNeighborhoodCode: lat=%.6f, lon=%.6f", lat, lon)

	// Use PDOK WFS service for CBS neighborhoods
	// Documentation: https://www.pdok.nl/datasets/cbs-gebiedsindelingen
	wfsURL := "https://service.pdok.nl/cbs/gebiedsindelingen/2022/wfs/v1_0"
	
	// Build WFS GetFeature request for buurten (neighborhoods) using coordinates
	// Using EPSG:4326 (WGS84) coordinate system
	url := fmt.Sprintf("%s?service=WFS&version=2.0.0&request=GetFeature&typeName=cbs_buurten_2022&outputFormat=application/json&srsName=EPSG:4326&bbox=%.6f,%.6f,%.6f,%.6f,EPSG:4326",
		wfsURL, lon-0.001, lat-0.001, lon+0.001, lat+0.001)

	logutil.Debugf("[CBS] WFS Request URL: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logutil.Debugf("[CBS] Request error: %v", err)
		return nil, fmt.Errorf("failed to create WFS request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[CBS] HTTP error: %v", err)
		return nil, fmt.Errorf("WFS request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logutil.Debugf("[CBS] Non-200 status: %d", resp.StatusCode)
		return nil, fmt.Errorf("WFS returned status %d", resp.StatusCode)
	}

	// Parse GeoJSON response
	var geoJSON struct {
		Type     string `json:"type"`
		Features []struct {
			Type       string `json:"type"`
			Properties struct {
				Buurtcode    string `json:"buurtcode"`
				Buurtnaam    string `json:"buurtnaam"`
				Wijkcode     string `json:"wijkcode"`
				Wijknaam     string `json:"wijknaam"`
				Gemeentecode string `json:"gemeentecode"`
				Gemeentenaam string `json:"gemeentenaam"`
			} `json:"properties"`
			Geometry interface{} `json:"geometry"`
		} `json:"features"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geoJSON); err != nil {
		logutil.Debugf("[CBS] Decode error: %v", err)
		return nil, fmt.Errorf("failed to parse WFS response: %w", err)
	}

	logutil.Debugf("[CBS] Found %d features", len(geoJSON.Features))

	if len(geoJSON.Features) == 0 {
		return nil, fmt.Errorf("no neighborhood found for coordinates")
	}

	// Use the first matching feature
	props := geoJSON.Features[0].Properties

	result := &NeighborhoodRegionCodes{
		NeighborhoodCode: props.Buurtcode,
		DistrictCode:     props.Wijkcode,
		MunicipalityCode: props.Gemeentecode,
		NeighborhoodName: props.Buurtnaam,
		DistrictName:     props.Wijknaam,
		MunicipalityName: props.Gemeentenaam,
	}

	logutil.Debugf("[CBS] Resolved codes: neighborhood=%s, district=%s, municipality=%s",
		result.NeighborhoodCode, result.DistrictCode, result.MunicipalityCode)

	return result, nil
}
