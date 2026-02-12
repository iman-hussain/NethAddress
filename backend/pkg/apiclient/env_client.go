package apiclient

import (
	"context"
	"fmt"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
	"github.com/iman-hussain/nethaddress/backend/pkg/logutil"
)

// AsbestosData contains information about asbestos presence at a property
type AsbestosData struct {
	HasAsbestosReport bool   `json:"hasAsbestosReport"`
	ReportDate        string `json:"reportDate,omitempty"`
	Status            string `json:"status,omitempty"` // "unknown", "clean", "present", "remediated"
}

// FetchAsbestosData queries for asbestos-related records at a given location.
// Currently returns a default response as there is no public API for asbestos data in NL.
// Future implementation options:
//   - BAG mutations API (check for asbestos-related building modifications)
//   - Municipality-specific APIs (varies by gemeente)
//   - Commercial property data providers
func (ac *ApiClient) FetchAsbestosData(ctx context.Context, cfg *config.Config, lat, lon float64) (*AsbestosData, error) {
	logutil.Debugf("[Asbestos] Checking asbestos data for lat=%.6f, lon=%.6f", lat, lon)

	// No public NL-wide asbestos registry API exists; return unknown status
	return &AsbestosData{
		HasAsbestosReport: false,
		Status:            "unknown",
	}, nil
}

// FetchAsbestosDataLegacy maintains backward compatibility with geometry-based calls
func (ac *ApiClient) FetchAsbestosDataLegacy(cfg *config.Config, geom interface{}) (*AsbestosData, error) {
	// Extract coordinates from geometry if possible, otherwise return default
	logutil.Debugf("[Asbestos] Legacy geometry-based query (returning default)")
	return &AsbestosData{
		HasAsbestosReport: false,
		Status:            "unknown",
	}, nil
}

// emptyAsbestosData returns a default empty AsbestosData for error cases
func emptyAsbestosData() *AsbestosData {
	return &AsbestosData{
		HasAsbestosReport: false,
		Status:            "unknown",
	}
}

// validateAsbestosCoordinates checks if coordinates are valid for Netherlands
func validateAsbestosCoordinates(lat, lon float64) error {
	// Netherlands bounding box (approximate)
	if lat < 50.5 || lat > 53.7 || lon < 3.2 || lon > 7.3 {
		return fmt.Errorf("coordinates outside Netherlands bounds")
	}
	return nil
}
