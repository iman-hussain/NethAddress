package apiclient

import (
	"context"
	"fmt"

	"github.com/iman-hussain/nethaddress/backend/pkg/config"
	"github.com/iman-hussain/nethaddress/backend/pkg/models"
)

// emptyWURSoilData returns a default WURSoilData struct for soft failures.
func emptyWURSoilData() *models.WURSoilData {
	return &models.WURSoilData{
		SoilType:      "Unknown",
		Composition:   "Unknown",
		Permeability:  0,
		OrganicMatter: 0,
		PH:            0,
		Suitability:   "Unknown",
	}
}

// FetchWURSoilData retrieves soil properties for land quality assessment
// Documentation: https://www.soilphysics.wur.nl
func (c *ApiClient) FetchWURSoilData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.WURSoilData, error) {
	if cfg.WURSoilApiURL == "" {
		// Return default data when API is not configured (requires agreement)
		return emptyWURSoilData(), nil
	}

	url := fmt.Sprintf("%s/soil?lat=%f&lon=%f", cfg.WURSoilApiURL, lat, lon)

	var result models.WURSoilData
	if err := c.GetJSON(ctx, "WUR Soil", url, nil, &result); err != nil {
		return emptyWURSoilData(), nil
	}

	return &result, nil
}

// emptySubsidenceData returns a default SubsidenceData struct for soft failures.
func emptySubsidenceData() *models.SubsidenceData {
	return &models.SubsidenceData{
		SubsidenceRate:  0,
		TotalSubsidence: 0,
		StabilityRating: "Unknown",
		MeasurementDate: "",
		GroundMovement:  0,
	}
}

// FetchSubsidenceData retrieves land subsidence and stability data
// Documentation: https://bodemdalingskaart.nl
func (c *ApiClient) FetchSubsidenceData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.SubsidenceData, error) {
	if cfg.SkyGeoSubsidenceApiURL == "" {
		// Return default data when API is not configured (paid service)
		return emptySubsidenceData(), nil
	}

	url := fmt.Sprintf("%s/subsidence?lat=%f&lon=%f", cfg.SkyGeoSubsidenceApiURL, lat, lon)

	var result models.SubsidenceData
	if err := c.GetJSON(ctx, "Subsidence", url, nil, &result); err != nil {
		return emptySubsidenceData(), nil
	}

	return &result, nil
}

// FetchSoilQualityData retrieves soil contamination data
// Documentation: https://api.store (government soil quality API)
func (c *ApiClient) FetchSoilQualityData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.SoilQualityData, error) {
	if cfg.SoilQualityApiURL == "" {
		return nil, fmt.Errorf("SoilQualityApiURL not configured")
	}

	url := fmt.Sprintf("%s/soil-quality?lat=%f&lon=%f", cfg.SoilQualityApiURL, lat, lon)

	var result models.SoilQualityData
	if err := c.GetJSON(ctx, "Soil Quality", url, nil, &result); err != nil {
		// Return empty data for failures rather than error (soft failure)
		return &models.SoilQualityData{
			ContaminationLevel: "Unknown",
			Contaminants:       []string{},
			RestrictedUse:      false,
		}, nil
	}

	return &result, nil
}

// emptyBROSoilMapData returns a default BROSoilMapData struct for soft failures.
func emptyBROSoilMapData() *models.BROSoilMapData {
	return &models.BROSoilMapData{
		SoilType:          "Unknown",
		PeatComposition:   0,
		Profile:           "Unknown",
		FoundationQuality: "Unknown",
		GroundwaterDepth:  0,
	}
}

// FetchBROSoilMapData retrieves BRO soil map data for foundation quality
// Documentation: https://www.dinoloket.nl/en/bro-soil-map
func (c *ApiClient) FetchBROSoilMapData(ctx context.Context, cfg *config.Config, lat, lon float64) (*models.BROSoilMapData, error) {
	// Return default data if not configured
	if cfg.BROSoilMapApiURL == "" {
		return emptyBROSoilMapData(), nil
	}

	url := fmt.Sprintf("%s/bro/soil-map?lat=%f&lon=%f", cfg.BROSoilMapApiURL, lat, lon)

	var result models.BROSoilMapData
	if err := c.GetJSON(ctx, "BRO Soil Map", url, nil, &result); err != nil {
		return emptyBROSoilMapData(), nil
	}

	return &result, nil
}
