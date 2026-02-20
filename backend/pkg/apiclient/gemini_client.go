package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
	"github.com/iman-hussain/nethaddress/backend/pkg/logutil"
	"github.com/iman-hussain/nethaddress/backend/pkg/models"
)

// geminiRequest represents the request body for Gemini API
type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens"`
	Temperature     float64 `json:"temperature"`
}

// geminiResponse represents the response from Gemini API
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// failedGeminiSummary returns a GeminiSummary indicating failure with the given error message.
func failedGeminiSummary(errorMsg string) *models.GeminiSummary {
	return &models.GeminiSummary{
		Generated: false,
		Error:     errorMsg,
	}
}

// GeminiPrompt is the hardcoded prompt template for generating location summaries.
// Location: backend/pkg/apiclient/gemini_client.go, line ~55
// Modify this prompt to change how Gemini analyses and summarises property data.
const GeminiPrompt = `You are an expert Dutch property analyst. Analyse this JSON data about a Dutch property location and provide a concise summary (max 800 characters).

Your response MUST include:
1. **Investment potential** (property value trends, area development)
2. **Business opportunities** (what businesses would thrive here based on demographics, foot traffic, nearby amenities)
3. **Liveability** (is it good for families, professionals, retirees?)
4. **Key risks** (flooding, noise, pollution, crime)

Be direct and specific. Use data from the JSON, and other information you can gather from the location. Refer to key figures and facts from the JSON. Do not mention if any data is missing or unavailable. No fluff. British English.

JSON Data:
%s`

// GenerateLocationSummary sends property data to Gemini and returns an AI-generated summary
func (c *ApiClient) GenerateLocationSummary(ctx context.Context, cfg *config.Config, propertyData interface{}) (*models.GeminiSummary, error) {
	if cfg.GeminiApiKey == "" {
		logutil.Debugf("[Gemini] API key not configured")
		return failedGeminiSummary("Gemini API key not configured"), nil
	}

	// Marshal property data to JSON
	jsonData, err := json.Marshal(propertyData)
	if err != nil {
		logutil.Debugf("[Gemini] Failed to marshal property data: %v", err)
		return failedGeminiSummary("Failed to prepare data for AI analysis"), nil
	}

	// Truncate JSON if too large (Gemini has input limits)
	maxDataSize := 30000 // ~30KB of JSON data
	if len(jsonData) > maxDataSize {
		jsonData = jsonData[:maxDataSize]
		logutil.Debugf("[Gemini] Truncated JSON data to %d bytes", maxDataSize)
	}

	// Build the prompt with property data
	prompt := fmt.Sprintf(GeminiPrompt, string(jsonData))

	// Prepare Gemini API request
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			MaxOutputTokens: 200, // ~500 characters
			Temperature:     0.7,
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		logutil.Debugf("[Gemini] Failed to marshal request: %v", err)
		return failedGeminiSummary("Failed to prepare AI request"), nil
	}

	// Gemini 2.5 Flash-Lite API endpoint (GA model, optimised for low latency)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-lite:generateContent?key=%s", cfg.GeminiApiKey)

	logutil.Debugf("[Gemini] Sending request to Gemini API")

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		logutil.Debugf("[Gemini] Failed to create request: %v", err)
		return failedGeminiSummary("Failed to create AI request"), nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[Gemini] HTTP request failed: %v", err)
		return failedGeminiSummary("Failed to connect to AI service"), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logutil.Debugf("[Gemini] Failed to read response: %v", err)
		return failedGeminiSummary("Failed to read AI response"), nil
	}

	logutil.Debugf("[Gemini] Response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		logutil.Debugf("[Gemini] API error response: %s", string(body))
		return failedGeminiSummary(fmt.Sprintf("AI service returned status %d", resp.StatusCode)), nil
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		logutil.Debugf("[Gemini] Failed to parse response: %v", err)
		return failedGeminiSummary("Failed to parse AI response"), nil
	}

	if geminiResp.Error != nil {
		logutil.Debugf("[Gemini] API error: %s", geminiResp.Error.Message)
		return failedGeminiSummary(geminiResp.Error.Message), nil
	}

	// Extract the generated text
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		logutil.Debugf("[Gemini] No content in response")
		return failedGeminiSummary("AI returned empty response"), nil
	}

	summary := geminiResp.Candidates[0].Content.Parts[0].Text
	logutil.Debugf("[Gemini] Generated summary: %d characters", len(summary))

	return &models.GeminiSummary{
		Summary:   summary,
		Generated: true,
	}, nil
}

// GeminiSolarPrompt is the hardcoded prompt template for generating solar eligibility summaries.
const GeminiSolarPrompt = `You are an expert Dutch solar energy analyst. Analyse this JSON data about a specific roof/area in the Netherlands and provide a concise summary (max 600 characters) of its solar panel viability.

Your response MUST consider:
1. **The physical area:** The provided area is %.2f square meters. Note: This is a 2D top-down footprint; it does not account for 3D roof pitch, so factor this margin of error into your recommendation.
2. **Environmental factors:** Interpret the current/historical precipitation, sunshine duration, and solar radiation from the provided JSON.

Be direct and specific. Focus strictly on whether this surface is viable for solar panels, the approximate potential, and any immediate environmental considerations. Do not mention if data is missing. No fluff. British English.

JSON Data:
%s`

// GenerateSolarEligibilitySummary sends solar/weather data to Gemini to get a solar viability assessment
func (c *ApiClient) GenerateSolarEligibilitySummary(ctx context.Context, cfg *config.Config, areaSqm float64, data interface{}) (*models.GeminiSummary, error) {
	if cfg.GeminiApiKey == "" {
		logutil.Debugf("[Gemini] API key not configured for solar")
		return failedGeminiSummary("Gemini API key not configured"), nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		logutil.Debugf("[Gemini] Failed to marshal solar data: %v", err)
		return failedGeminiSummary("Failed to prepare data for AI analysis"), nil
	}

	maxDataSize := 30000
	if len(jsonData) > maxDataSize {
		jsonData = jsonData[:maxDataSize]
		logutil.Debugf("[Gemini] Truncated JSON data to %d bytes", maxDataSize)
	}

	prompt := fmt.Sprintf(GeminiSolarPrompt, areaSqm, string(jsonData))

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			MaxOutputTokens: 200, // ~500 characters
			Temperature:     0.7,
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		logutil.Debugf("[Gemini] Failed to marshal request: %v", err)
		return failedGeminiSummary("Failed to prepare AI request"), nil
	}

	// Gemini 2.5 Flash-Lite API endpoint
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-lite:generateContent?key=%s", cfg.GeminiApiKey)

	logutil.Debugf("[Gemini] Sending solar request to Gemini API")

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		logutil.Debugf("[Gemini] Failed to create request: %v", err)
		return failedGeminiSummary("Failed to create AI request"), nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		logutil.Debugf("[Gemini] HTTP request failed: %v", err)
		return failedGeminiSummary("Failed to connect to AI service"), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logutil.Debugf("[Gemini] Failed to read response: %v", err)
		return failedGeminiSummary("Failed to read AI response"), nil
	}

	if resp.StatusCode != http.StatusOK {
		logutil.Debugf("[Gemini] API error response: %s", string(body))
		return failedGeminiSummary(fmt.Sprintf("AI service returned status %d", resp.StatusCode)), nil
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		logutil.Debugf("[Gemini] Failed to parse response: %v", err)
		return failedGeminiSummary("Failed to parse AI response"), nil
	}

	if geminiResp.Error != nil {
		logutil.Debugf("[Gemini] API error: %s", geminiResp.Error.Message)
		return failedGeminiSummary(geminiResp.Error.Message), nil
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		logutil.Debugf("[Gemini] No content in response")
		return failedGeminiSummary("AI returned empty response"), nil
	}

	summary := geminiResp.Candidates[0].Content.Parts[0].Text
	logutil.Debugf("[Gemini] Generated solar summary: %d characters", len(summary))

	return &models.GeminiSummary{
		Summary:   summary,
		Generated: true,
	}, nil
}
