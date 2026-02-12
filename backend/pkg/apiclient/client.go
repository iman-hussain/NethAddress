package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
	"github.com/iman-hussain/nethaddress/backend/pkg/logutil"

	"github.com/iman-hussain/nethaddress/backend/pkg/models"
)

// ApiClient for external API calls
type ApiClient struct {
	HTTP *http.Client
	cfg  *config.Config
}

func NewApiClient(client *http.Client, cfg *config.Config) *ApiClient {
	if client == nil {
		// default client with reasonable timeout to avoid hanging requests
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &ApiClient{
		HTTP: client,
		cfg:  cfg,
	}
}

// BearerAuthHeader returns a header map with Bearer token authorization.
// Returns nil if token is empty.
func BearerAuthHeader(token string) map[string]string {
	if token == "" {
		return nil
	}
	return map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)}
}

// GetJSON performs a GET request to the given URL, sets standard headers (merged with
// any custom headers provided), validates the response status is 2xx, and decodes the
// JSON response body into the target interface.
//
// Returns an error if the request fails, the status is non-2xx, or JSON decoding fails.
// Callers should handle errors by returning appropriate empty/default models to preserve
// the "soft failure" behaviour.
func (c *ApiClient) GetJSON(ctx context.Context, apiName, url string, headers map[string]string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logutil.Debugf("[%s] Request creation failed: %v", apiName, err)
		return fmt.Errorf("request creation failed: %w", err)
	}

	// Set standard header
	req.Header.Set("Accept", "application/json")

	// Merge custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[%s] HTTP request failed: %v", apiName, err)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		logutil.Debugf("[%s] Non-2xx status: %d", apiName, resp.StatusCode)
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		logutil.Debugf("[%s] JSON decode failed: %v", apiName, err)
		return fmt.Errorf("JSON decode failed: %w", err)
	}

	return nil
}

// PostJSON performs a POST request with a JSON body to the given URL.
// Sets Content-Type and Accept to application/json, merges custom headers,
// validates the response status is 2xx, and decodes the JSON response body into target.
func (c *ApiClient) PostJSON(ctx context.Context, apiName, url string, body interface{}, headers map[string]string, target interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		logutil.Debugf("[%s] JSON marshal failed: %v", apiName, err)
		return fmt.Errorf("JSON marshal failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		logutil.Debugf("[%s] Request creation failed: %v", apiName, err)
		return fmt.Errorf("request creation failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[%s] HTTP request failed: %v", apiName, err)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		logutil.Debugf("[%s] Non-2xx status: %d", apiName, resp.StatusCode)
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		logutil.Debugf("[%s] JSON decode failed: %v", apiName, err)
		return fmt.Errorf("JSON decode failed: %w", err)
	}

	return nil
}

// PostFormJSON performs a POST request with form-encoded body (application/x-www-form-urlencoded).
// Expects a JSON response and decodes it into target.
func (c *ApiClient) PostFormJSON(ctx context.Context, apiName, url, formData string, headers map[string]string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(formData))
	if err != nil {
		logutil.Debugf("[%s] Request creation failed: %v", apiName, err)
		return fmt.Errorf("request creation failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[%s] HTTP request failed: %v", apiName, err)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		logutil.Debugf("[%s] Non-2xx status: %d", apiName, resp.StatusCode)
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		logutil.Debugf("[%s] JSON decode failed: %v", apiName, err)
		return fmt.Errorf("JSON decode failed: %w", err)
	}

	return nil
}

// GetJSONWithRetry performs a GET request with exponential backoff retry logic.
// Combines retryWithBackoff + GetJSON for convenience.
func (c *ApiClient) GetJSONWithRetry(ctx context.Context, apiName, url string, headers map[string]string, maxAttempts int, initialDelay time.Duration, target interface{}) error {
	return c.retryWithBackoff(ctx, apiName, maxAttempts, initialDelay, func(retryCtx context.Context) error {
		return c.GetJSON(retryCtx, apiName, url, headers, target)
	})
}

// PostFormJSONWithRetry performs a POST form request with exponential backoff retry logic.
func (c *ApiClient) PostFormJSONWithRetry(ctx context.Context, apiName, url, formData string, headers map[string]string, maxAttempts int, initialDelay time.Duration, target interface{}) error {
	return c.retryWithBackoff(ctx, apiName, maxAttempts, initialDelay, func(retryCtx context.Context) error {
		return c.PostFormJSON(retryCtx, apiName, url, formData, headers, target)
	})
}

// retryWithBackoff executes fn with exponential backoff retries on failure.
// Retries up to maxAttempts times with the first retry after retryDelay.
// Returns early if context is cancelled or if fn succeeds.
func (c *ApiClient) retryWithBackoff(ctx context.Context, apiName string, maxAttempts int, initialDelay time.Duration, fn func(context.Context) error) error {
	var lastErr error
	delay := initialDelay

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn(ctx)
		if err == nil {
			return nil // Success on this attempt
		}

		lastErr = err
		if attempt < maxAttempts {
			logutil.Debugf("[%s] Attempt %d failed (%v), retrying in %v...", apiName, attempt, err, delay)
			select {
			case <-time.After(delay):
				// Exponential backoff: double the delay for next attempt
				delay = delay * 2
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	logutil.Debugf("[%s] All %d attempts failed, last error: %v", apiName, maxAttempts, lastErr)
	return lastErr
}

func (c *ApiClient) FetchBAGData(ctx context.Context, postcode, number string) (*models.BAGData, error) {
	postcode = strings.ToUpper(strings.TrimSpace(postcode))
	number = strings.TrimSpace(number)
	if postcode == "" || number == "" {
		return nil, fmt.Errorf("postcode and house number are required")
	}

	logutil.Debugf("[BAG] FetchBAGData: postcode=%s, number=%s", postcode, number)
	endpoint := c.cfg.BagApiURL

	params := url.Values{}
	params.Set("q", fmt.Sprintf("postcode:%s AND huisnummer:%s", postcode, number))
	params.Set("fq", "type:adres")
	params.Set("rows", "1")
	params.Set("wt", "json")

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
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
	var apiResp models.BagResponse
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

func (c *ApiClient) FetchPDOKData(ctx context.Context, coordinates string) (*models.PDOKData, error) {
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

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
				Naam               string `json:"naam"`
				Plantype           string `json:"plantype"`
				PlanStatus         string `json:"planstatus"`
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
