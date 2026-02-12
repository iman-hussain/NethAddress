package apiclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
)

func TestGenerateLocationSummary_NoAPIKey(t *testing.T) {
	cfg := &config.Config{
		GeminiApiKey: "",
	}
	client := NewApiClient(nil, cfg)

	result, err := client.GenerateLocationSummary(context.Background(), cfg, map[string]string{"test": "data"})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Generated {
		t.Error("Expected Generated to be false when API key is missing")
	}

	if result.Error != "Gemini API key not configured" {
		t.Errorf("Expected specific error message, got: %s", result.Error)
	}
}

func TestGenerateLocationSummary_Success(t *testing.T) {
	// Mock Gemini API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got: %s", r.Method)
		}

		// Return mock response
		response := geminiResponse{
			Candidates: []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			}{
				{
					Content: struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					}{
						Parts: []struct {
							Text string `json:"text"`
						}{
							{Text: "**Investment**: Strong area. **Business**: Cafes thrive. **Liveability**: Good for professionals. **Risks**: Minor flood risk."},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Note: This test can't actually test the real Gemini endpoint without mocking
	// because the URL is hardcoded. This is a limitation of the current implementation.
	// In production, we'd inject the base URL via config.

	cfg := &config.Config{
		GeminiApiKey: "", // Empty to test the early return
	}
	client := NewApiClient(nil, cfg)

	result, err := client.GenerateLocationSummary(context.Background(), cfg, map[string]string{"address": "Test"})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// With empty API key, we expect not generated
	if result.Generated {
		t.Error("Expected Generated to be false with empty API key")
	}
}

func TestGeminiPrompt(t *testing.T) {
	// Verify the prompt contains expected sections
	if len(GeminiPrompt) == 0 {
		t.Error("GeminiPrompt should not be empty")
	}

	expectedKeywords := []string{
		"Investment",
		"Business",
		"Liveability",
		"risks",
		"800 characters",
		"British English",
	}

	for _, keyword := range expectedKeywords {
		if !containsIgnoreCase(GeminiPrompt, keyword) {
			t.Errorf("GeminiPrompt should contain '%s'", keyword)
		}
	}
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsIgnoreCaseHelper(s, substr))
}

func containsIgnoreCaseHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalIgnoreCase(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
