package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/iman-hussain/AddressIQ/backend/pkg/aggregator"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
)

// HandleSearchStream handles the /api/search/stream endpoint for SSE
func (h *SearchHandler) HandleSearchStream(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Flush immediately to establish connection
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	// Normalize inputs for consistent caching
	postcode := strings.ToUpper(strings.ReplaceAll(r.URL.Query().Get("postcode"), " ", ""))
	houseNumber := strings.TrimSpace(r.URL.Query().Get("houseNumber"))
	bypassCache := r.URL.Query().Get("bypassCache") == "true"

	if postcode == "" || houseNumber == "" {
		sendSSEError(w, flusher, "Missing postcode or houseNumber")
		return
	}

	// Parse optional user provided API keys
	var userKeys map[string]string
	if keysParam := r.URL.Query().Get("apiKeys"); keysParam != "" {
		if err := json.Unmarshal([]byte(keysParam), &userKeys); err != nil {
			logutil.Warnf("Failed to parse user API keys: %v", err)
		}
	}

	// Validate admin secret for cache bypass
	if bypassCache {
		adminSecret := h.config.AdminSecret
		authHeader := r.Header.Get("X-Admin-Secret")
		// Also allow query param for SSE since headers might be tricky in pure EventSource (though usually not)
		if authHeader == "" {
			authHeader = r.URL.Query().Get("adminSecret")
		}

		if adminSecret != "" && authHeader != adminSecret {
			logutil.Warnf("Security: Stream cache bypass denied for %s (invalid secret)", r.RemoteAddr)
			bypassCache = false
		} else {
			logutil.Infof("Admin authorized override: Bypassing cache for stream %s %s", postcode, houseNumber)
		}
	}

	logutil.Infof("Starting stream search for %s %s", postcode, houseNumber)

	// Check if data is in cache (quick check before streaming)
	cacheData, isCacheHit := h.aggregator.GetCachedData(postcode, houseNumber)

	// Create progress channel
	progressCh := make(chan aggregator.ProgressEvent, 50) // Buffer to prevent blocking

	// Channel for final result
	resultCh := make(chan *aggregator.ComprehensivePropertyData)
	errCh := make(chan error)

	// If cache hit and not bypassing, return immediately without streaming
	if isCacheHit && !bypassCache && cacheData != nil {
		logutil.Infof("Cache hit detected at stream start for %s %s", postcode, houseNumber)

		// Build full response from cache
		apiResults := h.buildAPIResults(cacheData)
		response := ComprehensiveSearchResponse{
			Address:     cacheData.Address,
			Coordinates: cacheData.Coordinates,
			GeoJSON:     cacheData.GeoJSON,
			APIResults:  apiResults,
			AISummary:   cacheData.AISummary,
		}

		// Send cached data immediately
		responseJSON, _ := json.Marshal(response)
		fmt.Fprintf(w, "event: data\ndata: %s\n\n", responseJSON)
		flusher.Flush()

		// Send completion event
		htmlContent := fmt.Sprintf(`
<div data-target="header">
    <div class="box">
        <h5 class="title is-5">%s</h5>
        <p class="is-size-6"><strong>Coordinates:</strong> %.6f, %.6f</p>
        <p class="is-size-6"><strong>Postcode:</strong> %s | <strong>House Number:</strong> %s</p>
		<div class="address-buttons mt-3">
			<button class="button is-small glass-liquid" onclick="exportCSV()">
				⬇️ Export CSV
			</button>
			<button class="button is-small glass-liquid" onclick="refreshData()">
				🔄 Refresh
			</button>
            <button class="button is-small glass-liquid" onclick="openSettings()">
                ⚙️ Settings
            </button>
        </div>
    </div>
</div>
<div data-target="results">
</div>
<div data-geojson='%s' style="display:none;"></div>`,
			html.EscapeString(cacheData.Address),
			cacheData.Coordinates[1], cacheData.Coordinates[0],
			html.EscapeString(postcode), html.EscapeString(houseNumber),
			html.EscapeString(cacheData.GeoJSON))

		htmlPayload, _ := json.Marshal(htmlContent)
		fmt.Fprintf(w, "event: complete\ndata: %s\n\n", htmlPayload)
		flusher.Flush()
		return
	}

	// Start aggregation in a goroutine
	go func() {
		// Ensure all channels are closed on exit to unblock main loop
		defer close(progressCh)
		defer close(resultCh)
		defer close(errCh)

		// Create a context that is cancelled when this goroutine returns
		// (though we use request context for fetch operations)

		// Use request context for fetching so client disconnect cancels work
		data, err := h.aggregator.AggregatePropertyDataWithOptions(r.Context(), postcode, houseNumber, bypassCache, progressCh, userKeys)
		if err != nil {
			// If context was cancelled, err will effectively be ignored by receiver loop
			errCh <- err
			return
		}
		resultCh <- data
	}()

	// Send initial event
	fmt.Fprintf(w, "event: start\ndata: {\"message\": \"Starting search...\"}\n\n")
	flusher.Flush()

	// Keep-alive ticker
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Main loop: Listen for progress, results, errors, or client disconnect
	for {
		select {
		case ev, ok := <-progressCh:
			if !ok {
				// Progress channel closed, wait for result or error
				progressCh = nil // Disable this case
				continue
			}

			if ev.Data != nil {
				// Send partial update with data
				payload, _ := json.Marshal(ev) // Includes Data field
				fmt.Fprintf(w, "event: update\ndata: %s\n\n", payload)
			} else {
				// Status only update
				payload, _ := json.Marshal(ev)
				fmt.Fprintf(w, "data: %s\n\n", payload)
			}
			flusher.Flush()

		case data := <-resultCh:
			if data == nil {
				return // Should not happen if errCh logic is correct
			}
			// Build full response
			apiResults := h.buildAPIResults(data)

			response := ComprehensiveSearchResponse{
				Address:     data.Address,
				Coordinates: data.Coordinates,
				GeoJSON:     data.GeoJSON, // Use aggregated GeoJSON
				APIResults:  apiResults,
				AISummary:   data.AISummary,
			}

			// Serialize to JSON
			responseJSON, err := json.Marshal(response)
			if err != nil {
				logutil.Errorf("Error marshaling response: %v", err)
				sendSSEError(w, flusher, "Failed to serialize response")
				return
			}

			// 1. Send Data Event (Raw JSON) - Efficient transport
			fmt.Fprintf(w, "event: data\ndata: %s\n\n", responseJSON)
			flusher.Flush()

			// 2. Send HTML Event (Presentation only) - Lightweight
			// Note: We remove the heavy data-response attribute since data is sent separately
			htmlContent := fmt.Sprintf(`
<div data-target="header">
    <div class="box">
        <h5 class="title is-5">%s</h5>
        <p class="is-size-6"><strong>Coordinates:</strong> %.6f, %.6f</p>
        <p class="is-size-6"><strong>Postcode:</strong> %s | <strong>House Number:</strong> %s</p>
		<div class="address-buttons mt-3">
			<button class="button is-small glass-liquid" onclick="exportCSV()">
				⬇️ Export CSV
			</button>
			<button class="button is-small glass-liquid" onclick="refreshData()">
				🔄 Refresh
			</button>
            <button class="button is-small glass-liquid" onclick="openSettings()">
                ⚙️ Settings
            </button>
        </div>
    </div>
</div>
<div data-target="results">
</div>
<div data-geojson='%s' style="display:none;"></div>`,
				html.EscapeString(data.Address),
				data.Coordinates[1], data.Coordinates[0],
				html.EscapeString(postcode), html.EscapeString(houseNumber),
				html.EscapeString(data.GeoJSON))

			// Send HTML as JSON string for safe transport (frontend expects JSON string of HTML)
			htmlPayload, err := json.Marshal(htmlContent)
			if err != nil {
				logutil.Errorf("Error marshaling html response: %v", err)
				sendSSEError(w, flusher, "Failed to serialize response")
				return
			}
			fmt.Fprintf(w, "event: complete\ndata: %s\n\n", htmlPayload)
			flusher.Flush()
			return

		case err := <-errCh:
			if err != nil {
				logutil.Errorf("Stream aggregation error: %v", err)
				sendSSEError(w, flusher, err.Error())
			}
			return

		case <-ticker.C:
			// Send keep-alive comment
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()

		case <-r.Context().Done():
			logutil.Infof("Client disconnected during stream %s %s", postcode, houseNumber)
			return
		}
	}
}

func sendSSEError(w http.ResponseWriter, flusher http.Flusher, msg string) {
	// Escape the message for JSON string
	safeMsg, _ := json.Marshal(msg)
	fmt.Fprintf(w, "event: error\ndata: {\"message\": %s}\n\n", safeMsg)
	flusher.Flush()
}
