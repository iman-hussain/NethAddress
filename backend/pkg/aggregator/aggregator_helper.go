package aggregator

import "github.com/iman-hussain/AddressIQ/backend/pkg/config"

// countEnabledSources calculates the total number of expected API calls based on configuration
func (pa *PropertyAggregator) countEnabledSources(cfg *config.Config) int {
	count := 0

	// Free APIs (Always enabled)
	count += 1 // BAG Address (already fetched but part of process)
	count += 1 // KNMI Weather
	count += 1 // KNMI Solar
	count += 1 // Luchtmeetnet Air Quality
	count += 1 // CBS Population
	count += 1 // CBS Square Statistics
	count += 1 // BRO Soil Map
	count += 1 // NDW Traffic
	count += 1 // openOV Public Transport
	count += 1 // Flood Risk
	count += 1 // Green Spaces
	count += 1 // Education Facilities
	count += 1 // Facilities & Amenities
	count += 1 // AHN Height Model
	count += 1 // Monument Status
	count += 1 // PDOK Platform
	count += 1 // Land Use & Zoning

	// Freemium/Premium APIs (Check keys/config)
	if cfg.KadasterObjectInfoApiKey != "" {
		count++
	}
	if cfg.AltumWOZApiKey != "" {
		count++ // WOZ
	}
	if cfg.AltumTransactionApiKey != "" {
		count++ // Transactions
	}
	if cfg.AltumEnergyApiKey != "" {
		count++ // Energy
	}
	if cfg.AltumSustainabilityApiKey != "" {
		count++ // Sustainability
	}
	if cfg.MatrixianApiKey != "" {
		count++
	}
	if cfg.SkyGeoApiKey != "" {
		count++
	}
	if cfg.SchipholApiKey != "" {
		count++
	}
	if cfg.StratopoApiKey != "" {
		count++
	}

	// Specifics that might be conditional in fetchers
	// Noise Pollution (Freemium - usually enabled?)
	count++

	// Soil Physicals (WUR)
	// Check if configured (url usually set)
	count++

	// Soil Quality
	count++

	// Parking
	count++

	// Water Quality
	if cfg.DigitalDeltaApiKey != "" {
		count++
	}

	// Safety
	if cfg.CBSApiKey != "" {
		count++
	}

	// Building Permits
	count++

	return count
}
