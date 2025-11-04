package apiclient

import (
	"encoding/json"
	"encoding/xml"
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
	Weergavenaam      string  `json:"weergavenaam"`
	Straatnaam        string  `json:"straatnaam"`
	Huisnummer        float64 `json:"huisnummer"`
	HuisNLT           string  `json:"huis_nlt"`
	Huisletter        string  `json:"huisletter"`
	Huisnummertoevoeg string  `json:"huisnummertoevoeging"`
	Postcode          string  `json:"postcode"`
	WoonplaatsNaam    string  `json:"woonplaatsnaam"`
	CentroidLL        string  `json:"centroide_ll"`
	GeometriePolygoon string  `json:"geometrie_polygoon"`
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

	return &models.BAGData{
		Address:     address,
		Coordinates: coordinates,
		GeoJSON:     geoJSON,
	}, nil
}

// PDOK Zoning API response structs (simplified for WMS GetFeatureInfo XML)
type PDOKResponse struct {
	Omschrijving string `xml:"omschrijving"`
	Beperkingen  string `xml:"beperkingen"`
}

type WMSFeatureInfo struct {
	Zoning       string
	Restrictions []string
}

func (c *ApiClient) FetchPDOKData(coordinates string) (*models.PDOKData, error) {
	logutil.Debugf("[PDOK] FetchPDOKData: coordinates=%s", coordinates)
	// Example: coordinates = "4.8952,52.3702"
	// Build WMS GetFeatureInfo request (Ruimtelijke Plannen)
	// For demo, we use a static endpoint and parse a simple XML
	// In reality, you would need to build a proper WMS request
	url := "https://geodata.nationaalgeoregister.nl/plannen/wms?SERVICE=WMS&VERSION=1.1.1&REQUEST=GetFeatureInfo" +
		"&LAYERS=bestemmingsplannen&QUERY_LAYERS=bestemmingsplannen" +
		"&INFO_FORMAT=application/vnd.ogc.gml" +
		"&X=1&Y=1&SRS=EPSG:4326&WIDTH=1&HEIGHT=1" +
		"&BBOX=" + coordinates + "," + coordinates + "&FEATURE_COUNT=1"
	logutil.Debugf("[PDOK] Request URL: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logutil.Debugf("[PDOK] Request error: %v", err)
		return nil, err
	}
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
	// Parse XML using encoding/xml for robustness
	var pdokResp PDOKResponse
	if err := xml.Unmarshal(body, &pdokResp); err != nil {
		logutil.Debugf("[PDOK] XML unmarshal error: %v", err)
		// fallback to string search if unmarshal fails (for demo)
		xmlStr := string(body)
		if idx := strings.Index(xmlStr, "<omschrijving>"); idx != -1 {
			endIdx := strings.Index(xmlStr[idx:], "</omschrijving>")
			if endIdx != -1 {
				pdokResp.Omschrijving = strings.TrimSpace(xmlStr[idx+13 : idx+endIdx])
				pdokResp.Omschrijving = strings.TrimPrefix(pdokResp.Omschrijving, ">")
			}
		}
		if idx := strings.Index(xmlStr, "<beperkingen>"); idx != -1 {
			endIdx := strings.Index(xmlStr[idx:], "</beperkingen>")
			if endIdx != -1 {
				pdokResp.Beperkingen = xmlStr[idx+13 : idx+endIdx]
			}
		}
	}
	logutil.Debugf("[PDOK] Raw response: %s", string(body))
	logutil.Debugf("[PDOK] Parsed: %+v", pdokResp)
	zoning := pdokResp.Omschrijving
	if zoning == "" {
		zoning = "Unknown"
	}
	restrictions := []string{}
	if pdokResp.Beperkingen != "" {
		restrictions = append(restrictions, pdokResp.Beperkingen)
	}
	logutil.Debugf("[PDOK] Final data: zoning=%s, restrictions=%v", zoning, restrictions)
	return &models.PDOKData{
		ZoningInfo:   zoning,
		Restrictions: restrictions,
	}, nil
}
