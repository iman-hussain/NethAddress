package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"

	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

// ApiClient for external API calls
type ApiClient struct {
	HTTP *http.Client
}

func NewApiClient(client *http.Client) *ApiClient {
	if client == nil {
		// default client with reasonable timeout to avoid hanging requests
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &ApiClient{HTTP: client}
}

const defaultBAGEndpoint = "https://api.pdok.nl/bzk/locatieserver/search/v3_1/free"

// bagResponse mirrors the PDOK Locatieserver free endpoint JSON payload.
type bagResponse struct {
	Response struct {
		Docs []bagDocument `json:"docs"`
	} `json:"response"`
}

// bagDocument contains only the fields we actually use downstream.
type bagDocument struct {
	ID                      string  `json:"id"`
	NummeraanduidingID      string  `json:"nummeraanduiding_id"`
	VerblijfsobjectID       string  `json:"verblijfsobject_id"`
	PandID                  string  `json:"pand_id"`
	Weergavenaam            string  `json:"weergavenaam"`
	Straatnaam              string  `json:"straatnaam"`
	Huisnummer              float64 `json:"huisnummer"`
	HuisNLT                 string  `json:"huis_nlt"`
	Huisletter              string  `json:"huisletter"`
	Huisnummertoevoeg       string  `json:"huisnummertoevoeging"`
	Postcode                string  `json:"postcode"`
	WoonplaatsNaam          string  `json:"woonplaatsnaam"`
	Gemeentenaam            string  `json:"gemeentenaam"`
	Gemeentecode            string  `json:"gemeentecode"`
	Provincienaam           string  `json:"provincienaam"`
	Provinciecode           string  `json:"provinciecode"`
	CentroidLL              string  `json:"centroide_ll"`
	GeometriePolygoon       string  `json:"geometrie_polygoon"`
}

func (c *ApiClient) FetchBAGData(postcode, number string) (*models.BAGData, error) {
	postcode = strings.ToUpper(strings.TrimSpace(postcode))
	number = strings.TrimSpace(number)
	if postcode == "" || number == "" {
		return nil, fmt.Errorf("postcode and house number are required")
	}

	logutil.Debugf("[BAG] FetchBAGData: postcode=%s, number=%s", postcode, number)
	endpoint := strings.TrimSpace(os.Getenv("BAG_API_URL"))
	if endpoint == "" {
		endpoint = defaultBAGEndpoint
	}

	params := url.Values{}
	params.Set("q", fmt.Sprintf("postcode:%s AND huisnummer:%s", postcode, number))
	params.Set("fq", "type:adres")
	params.Set("rows", "1")
	params.Set("wt", "json")

	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		logutil.Debugf("[BAG] Request error: %v", err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[BAG] HTTP error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		logutil.Debugf("[BAG] Non-200 status: %d", resp.StatusCode)
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("BAG API returned status %d: %s", resp.StatusCode, string(b))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logutil.Debugf("[BAG] Read body error: %v", err)
		return nil, err
	}

	logutil.Debugf("[BAG] Raw response: %s", string(body))
	var apiResp bagResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		logutil.Debugf("[BAG] Unmarshal error: %v", err)
		return nil, fmt.Errorf("failed to parse BAG API response: %w", err)
	}

	logutil.Debugf("[BAG] Parsed response: %+v", apiResp)
	if len(apiResp.Response.Docs) == 0 {
		logutil.Debugf("[BAG] No results for %s %s", postcode, number)
		return nil, fmt.Errorf("no results from BAG API for %s %s", postcode, number)
	}

	doc := apiResp.Response.Docs[0]

	// Derive coordinates from the WKT POINT value provided by the API.
	var coordinates [2]float64
	if strings.HasPrefix(doc.CentroidLL, "POINT(") {
		inner := strings.TrimSuffix(strings.TrimPrefix(doc.CentroidLL, "POINT("), ")")
		parts := strings.Fields(inner)
		if len(parts) == 2 {
			if lon, err := strconv.ParseFloat(parts[0], 64); err == nil {
				coordinates[0] = lon
			}
			if lat, err := strconv.ParseFloat(parts[1], 64); err == nil {
				coordinates[1] = lat
			}
		}
	}
	if coordinates[0] == 0 && coordinates[1] == 0 {
		return nil, fmt.Errorf("failed to parse coordinates from BAG response")
	}

	address := strings.TrimSpace(doc.Weergavenaam)
	if address == "" {
		// Fall back to manually constructed address if weergavenaam is missing.
		var builder strings.Builder
		builder.WriteString(strings.TrimSpace(doc.Straatnaam))
		builder.WriteString(" ")
		if doc.HuisNLT != "" {
			builder.WriteString(strings.TrimSpace(doc.HuisNLT))
		} else if doc.Huisnummer > 0 {
			builder.WriteString(strconv.FormatInt(int64(doc.Huisnummer), 10))
			if doc.Huisletter != "" {
				builder.WriteString(strings.ToUpper(doc.Huisletter))
			}
			if doc.Huisnummertoevoeg != "" {
				builder.WriteString(strings.ToUpper(doc.Huisnummertoevoeg))
			}
		}
		if doc.Postcode != "" || doc.WoonplaatsNaam != "" {
			builder.WriteString(", ")
			builder.WriteString(strings.TrimSpace(doc.Postcode))
			if doc.WoonplaatsNaam != "" {
				builder.WriteString(" ")
				builder.WriteString(strings.TrimSpace(doc.WoonplaatsNaam))
			}
		}
		address = strings.TrimSpace(builder.String())
	}

	geoJSON := strings.TrimSpace(doc.GeometriePolygoon)
	if geoJSON == "" {
		geoJSON = fmt.Sprintf(`{"type":"Point","coordinates":[%f,%f]}`, coordinates[0], coordinates[1])
	}

	// Extract BAG IDs - prefer verblijfsobject_id, fallback to nummeraanduiding_id or id
	bagID := strings.TrimSpace(doc.VerblijfsobjectID)
	if bagID == "" {
		bagID = strings.TrimSpace(doc.NummeraanduidingID)
	}
	if bagID == "" {
		bagID = strings.TrimSpace(doc.ID)
	}

	logutil.Debugf("[BAG] Extracted BAG ID: %s (verblijfsobject: %s, nummeraanduiding: %s, pand: %s)",
		bagID, doc.VerblijfsobjectID, doc.NummeraanduidingID, doc.PandID)

	return &models.BAGData{
		Address:            address,
		Coordinates:        coordinates,
		GeoJSON:            geoJSON,
		ID:                 bagID,
		NummeraanduidingID: strings.TrimSpace(doc.NummeraanduidingID),
		VerblijfsobjectID:  strings.TrimSpace(doc.VerblijfsobjectID),
		PandID:             strings.TrimSpace(doc.PandID),
		Municipality:       strings.TrimSpace(doc.Gemeentenaam),
		MunicipalityCode:   strings.TrimSpace(doc.Gemeentecode),
		Province:           strings.TrimSpace(doc.Provincienaam),
		ProvinceCode:       strings.TrimSpace(doc.Provinciecode),
	}, nil
}

func (c *ApiClient) FetchPDOKData(coordinates string) (*models.PDOKData, error) {
	logutil.Debugf("[PDOK] FetchPDOKData: coordinates=%s", coordinates)
	
	// Parse coordinates (expected format: "lon,lat")
	parts := strings.Split(coordinates, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid coordinates format, expected 'lon,lat'")
	}
	
	lon := strings.TrimSpace(parts[0])
	lat := strings.TrimSpace(parts[1])
	
	// Use PDOK WFS for bestemmingsplannen (zoning plans)
	// Documentation: https://www.nationaalgeoregister.nl/geonetwork/srv/dut/catalog.search#/metadata/c10a2c55-972f-4309-a038-0e5286934877
	baseURL := "https://service.pdok.nl/roo/ruimtelijkeplannen/wfs/v1_0"
	
	// Build WFS GetFeature request with proper BBOX
	// Using a small buffer around the point to ensure we catch the containing polygon
	lonFloat, err := strconv.ParseFloat(lon, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude '%s': %w", lon, err)
	}
	latFloat, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude '%s': %w", lat, err)
	}
	buffer := 0.001 // ~100m buffer
	
	bbox := fmt.Sprintf("%.6f,%.6f,%.6f,%.6f", 
		lonFloat-buffer, latFloat-buffer, lonFloat+buffer, latFloat+buffer)
	
	url := fmt.Sprintf("%s?service=WFS&version=2.0.0&request=GetFeature&typeName=plangebied&outputFormat=application/json&srsName=EPSG:4326&bbox=%s,EPSG:4326&count=1",
		baseURL, bbox)
	
	logutil.Debugf("[PDOK] WFS Request URL: %s", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logutil.Debugf("[PDOK] Request error: %v", err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	
	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[PDOK] HTTP error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logutil.Debugf("[PDOK] Non-200 status: %d", resp.StatusCode)
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("PDOK API returned status %d: %s", resp.StatusCode, string(b))
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logutil.Debugf("[PDOK] Read body error: %v", err)
		return nil, err
	}
	
	// Log first 500 chars or less of response
	bodyPreview := string(body)
	if len(bodyPreview) > 500 {
		bodyPreview = bodyPreview[:500]
	}
	logutil.Debugf("[PDOK] Raw response (first 500 chars): %s", bodyPreview)
	
	// Parse GeoJSON response
	var geoJSON struct {
		Type     string `json:"type"`
		Features []struct {
			Type       string `json:"type"`
			Properties struct {
				Naam              string `json:"naam"`
				Plantype          string `json:"plantype"`
				PlanStatus        string `json:"planstatus"`
				Beleidsmatigstatus string `json:"beleidsmatigstatus"`
			} `json:"properties"`
		} `json:"features"`
	}
	
	if err := json.Unmarshal(body, &geoJSON); err != nil {
		logutil.Debugf("[PDOK] JSON unmarshal error: %v", err)
		return nil, fmt.Errorf("failed to parse PDOK response: %w", err)
	}
	
	logutil.Debugf("[PDOK] Found %d features", len(geoJSON.Features))
	
	zoning := "Unknown"
	restrictions := []string{}
	
	if len(geoJSON.Features) > 0 {
		props := geoJSON.Features[0].Properties
		zoning = props.Naam
		if props.Plantype != "" {
			restrictions = append(restrictions, fmt.Sprintf("Type: %s", props.Plantype))
		}
		if props.PlanStatus != "" {
			restrictions = append(restrictions, fmt.Sprintf("Status: %s", props.PlanStatus))
		}
	}
	
	logutil.Debugf("[PDOK] Final data: zoning=%s, restrictions=%v", zoning, restrictions)
	
	return &models.PDOKData{
		ZoningInfo:   zoning,
		Restrictions: restrictions,
	}, nil
}
