package aggregator

import (
	"fmt"
	"time"

	"github.com/iman-hussain/AddressIQ/backend/pkg/apiclient"
	"github.com/iman-hussain/AddressIQ/backend/pkg/cache"
	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
)

// PropertyAggregator combines data from multiple API sources
type PropertyAggregator struct {
	apiClient *apiclient.ApiClient
	cache     *cache.CacheService
	config    *config.Config
}

// NewPropertyAggregator creates a new aggregator instance
func NewPropertyAggregator(apiClient *apiclient.ApiClient, cacheService *cache.CacheService, cfg *config.Config) *PropertyAggregator {
	return &PropertyAggregator{
		apiClient: apiClient,
		cache:     cacheService,
		config:    cfg,
	}
}

// ComprehensivePropertyData represents all collected property data
type ComprehensivePropertyData struct {
	// Core Property Data
	Address     string     `json:"address"`
	Coordinates [2]float64 `json:"coordinates"`
	BAGID       string     `json:"bagId"`

	// Property Details
	KadasterInfo       *apiclient.KadasterObjectInfo     `json:"kadasterInfo,omitempty"`
	WOZData            *apiclient.AltumWOZData           `json:"wozData,omitempty"`
	MarketValuation    *apiclient.MatrixianPropertyValue `json:"marketValuation,omitempty"`
	TransactionHistory *apiclient.TransactionHistory     `json:"transactionHistory,omitempty"`
	MonumentStatus     *apiclient.MonumentData           `json:"monumentStatus,omitempty"`

	// Environmental Data
	Weather        *apiclient.KNMIWeatherData    `json:"weather,omitempty"`
	SolarPotential *apiclient.KNMISolarData      `json:"solarPotential,omitempty"`
	SoilData       *apiclient.WURSoilData        `json:"soilData,omitempty"`
	Subsidence     *apiclient.SubsidenceData     `json:"subsidence,omitempty"`
	SoilQuality    *apiclient.SoilQualityData    `json:"soilQuality,omitempty"`
	BROSoilMap     *apiclient.BROSoilMapData     `json:"broSoilMap,omitempty"`
	AirQuality     *apiclient.AirQualityData     `json:"airQuality,omitempty"`
	NoisePollution *apiclient.NoisePollutionData `json:"noisePollution,omitempty"`

	// Energy & Sustainability
	EnergyClimate  *apiclient.EnergyClimateData  `json:"energyClimate,omitempty"`
	Sustainability *apiclient.SustainabilityData `json:"sustainability,omitempty"`

	// Risk Assessment
	FloodRisk    *apiclient.FloodRiskData    `json:"floodRisk,omitempty"`
	WaterQuality *apiclient.WaterQualityData `json:"waterQuality,omitempty"`
	Safety       *apiclient.SafetyData       `json:"safety,omitempty"`

	// Mobility & Accessibility
	TrafficData     []apiclient.NDWTrafficData     `json:"trafficData,omitempty"`
	PublicTransport *apiclient.OpenOVTransportData `json:"publicTransport,omitempty"`
	ParkingData     *apiclient.ParkingData         `json:"parkingData,omitempty"`

	// Demographics & Neighborhood
	Population   *apiclient.CBSPopulationData  `json:"population,omitempty"`
	StatLineData *apiclient.CBSStatLineData    `json:"statLineData,omitempty"`
	SquareStats  *apiclient.CBSSquareStatsData `json:"squareStats,omitempty"`
	CBSData      *apiclient.CBSData            `json:"cbsData,omitempty"`

	// Infrastructure
	GreenSpaces     *apiclient.GreenSpacesData     `json:"greenSpaces,omitempty"`
	Education       *apiclient.EducationData       `json:"education,omitempty"`
	BuildingPermits *apiclient.BuildingPermitsData `json:"buildingPermits,omitempty"`
	Facilities      *apiclient.FacilitiesData      `json:"facilities,omitempty"`
	Elevation       *apiclient.AHNHeightData       `json:"elevation,omitempty"`

	// Comprehensive Platforms
	PDOKData            *apiclient.PDOKPlatformData        `json:"pdokData,omitempty"`
	StratopoEnvironment *apiclient.StratopoEnvironmentData `json:"stratopoEnvironment,omitempty"`
	LandUse             *apiclient.LandUseData             `json:"landUse,omitempty"`

	// Aviation
	SchipholFlights *apiclient.SchipholFlightData `json:"schipholFlights,omitempty"`

	// Metadata
	AggregatedAt time.Time         `json:"aggregatedAt"`
	DataSources  []string          `json:"dataSources"`
	Errors       map[string]string `json:"errors,omitempty"`
}

// AggregatePropertyData fetches and combines data from all available sources
func (pa *PropertyAggregator) AggregatePropertyData(postcode, houseNumber string) (*ComprehensivePropertyData, error) {
	logutil.Debugf("[AGGREGATOR] Starting aggregation for %s %s", postcode, houseNumber)

	// Check cache first (if available)
	if pa.cache != nil {
		cacheKey := cache.CacheKey{}.AggregatedKey(postcode, houseNumber)
		var cached ComprehensivePropertyData
		if err := pa.cache.Get(cacheKey, &cached); err == nil {
			logutil.Debugf("[AGGREGATOR] Cache hit for %s %s - returning cached data", postcode, houseNumber)
			return &cached, nil
		}
		logutil.Debugf("[AGGREGATOR] Cache miss for %s %s - fetching fresh data", postcode, houseNumber)
	}

	// Start with BAG data (essential)
	bagData, err := pa.apiClient.FetchBAGData(postcode, houseNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch BAG data: %w", err)
	}

	lat := bagData.Coordinates[1]
	lon := bagData.Coordinates[0]

	logutil.Debugf("[AGGREGATOR] Coordinates from BAG: lat=%.6f, lon=%.6f", lat, lon)

	// Extract BAG ID from response (prefer verblijfsobject_id, fallback to id)
	bagID := bagData.ID
	if bagID == "" {
		logutil.Debugf("[AGGREGATOR] Warning: No BAG ID found in response, some APIs may fail")
	}

	// Lookup neighborhood and region codes dynamically
	var neighborhoodCode, regionCode string
	regionCodes, err := pa.apiClient.LookupNeighborhoodCode(pa.config, lat, lon)
	if err == nil && regionCodes != nil {
		neighborhoodCode = regionCodes.NeighborhoodCode
		regionCode = regionCodes.MunicipalityCode
		logutil.Debugf("[AGGREGATOR] Resolved neighborhood=%s, region=%s", neighborhoodCode, regionCode)
	} else {
		logutil.Debugf("[AGGREGATOR] Warning: Could not resolve neighborhood codes: %v", err)
		// Fallback to municipality code from BAG data if available
		if bagData.MunicipalityCode != "" {
			regionCode = bagData.MunicipalityCode
			logutil.Debugf("[AGGREGATOR] Using municipality code from BAG: %s", regionCode)
		}
	}

	data := &ComprehensivePropertyData{
		Address:      bagData.Address,
		Coordinates:  bagData.Coordinates,
		BAGID:        bagID,
		AggregatedAt: time.Now(),
		DataSources:  []string{"BAG"},
		Errors:       make(map[string]string),
	}

	// Property & Valuation Data (parallel fetches where possible)
	pa.fetchPropertyData(data, bagID, lat, lon)

	// Environmental Data
	pa.fetchEnvironmentalData(data, lat, lon)

	// Energy & Sustainability
	pa.fetchEnergyData(data, bagID)

	// Risk Assessment
	pa.fetchRiskData(data, lat, lon, neighborhoodCode)

	// Mobility & Accessibility
	pa.fetchMobilityData(data, lat, lon)

	// Demographics & Neighborhood
	pa.fetchDemographicsData(data, lat, lon, neighborhoodCode, regionCode)

	// Infrastructure
	pa.fetchInfrastructureData(data, lat, lon)

	// Comprehensive Platforms
	pa.fetchPlatformData(data, lat, lon)

	// Cache the aggregated result (if caching is available)
	if pa.cache != nil {
		cacheKey := cache.CacheKey{}.AggregatedKey(postcode, houseNumber)
		pa.cache.Set(cacheKey, data, cache.PropertyDataTTL)
	}

	return data, nil
}

func (pa *PropertyAggregator) fetchPropertyData(data *ComprehensivePropertyData, bagID string, lat, lon float64) {
	// Kadaster Object Info
	if kadasterInfo, err := pa.apiClient.FetchKadasterObjectInfo(pa.config, bagID); err == nil {
		data.KadasterInfo = kadasterInfo
		data.DataSources = append(data.DataSources, "Kadaster")
	} else {
		logutil.Debugf("[AGGREGATOR] Kadaster fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Kadaster"] = err.Error()
		}
	}

	// WOZ Data
	if wozData, err := pa.apiClient.FetchAltumWOZData(pa.config, bagID); err == nil {
		data.WOZData = wozData
		data.DataSources = append(data.DataSources, "Altum WOZ")
	} else {
		logutil.Debugf("[AGGREGATOR] Altum WOZ fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Altum WOZ"] = err.Error()
		}
	}

	// Market Valuation
	if valuation, err := pa.apiClient.FetchPropertyValuePlus(pa.config, bagID, lat, lon); err == nil {
		data.MarketValuation = valuation
		data.DataSources = append(data.DataSources, "Matrixian")
	} else {
		logutil.Debugf("[AGGREGATOR] Matrixian fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Matrixian"] = err.Error()
		}
	}

	// Transaction History
	if transactions, err := pa.apiClient.FetchTransactionHistory(pa.config, bagID); err == nil {
		data.TransactionHistory = transactions
		data.DataSources = append(data.DataSources, "Altum Transactions")
	} else {
		logutil.Debugf("[AGGREGATOR] Altum Transactions fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Altum Transactions"] = err.Error()
		}
	}

	// Monument Status
	if monument, err := pa.apiClient.FetchMonumentData(pa.config, bagID); err == nil {
		data.MonumentStatus = monument
		data.DataSources = append(data.DataSources, "Monument Register")
	} else {
		logutil.Debugf("[AGGREGATOR] Monument fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Monument Register"] = err.Error()
		}
	}
}

func (pa *PropertyAggregator) fetchEnvironmentalData(data *ComprehensivePropertyData, lat, lon float64) {
	logutil.Debugf("[AGGREGATOR] fetchEnvironmentalData called with lat=%.6f, lon=%.6f", lat, lon)

	// Weather

	logutil.Debugf("[AGGREGATOR] Calling FetchKNMIWeatherData...")
	weather, weatherErr := pa.apiClient.FetchKNMIWeatherData(pa.config, lat, lon)
	if weatherErr == nil && weather != nil {
		data.Weather = weather
		data.DataSources = append(data.DataSources, "KNMI Weather")
		logutil.Debugf("[AGGREGATOR] ✓ Weather data fetched successfully: %+v", weather)
	} else {
		logutil.Debugf("[AGGREGATOR] ✗ Weather fetch failed: %v", weatherErr)
		if data.Errors != nil && weatherErr != nil {
			data.Errors["KNMI Weather"] = weatherErr.Error()
		}
	}

	logutil.Debugf("[AGGREGATOR] Calling FetchKNMISolarData...")
	solar, solarErr := pa.apiClient.FetchKNMISolarData(pa.config, lat, lon)
	if solarErr == nil && solar != nil {
		data.SolarPotential = solar
		data.DataSources = append(data.DataSources, "KNMI Solar")
		logutil.Debugf("[AGGREGATOR] ✓ Solar data fetched successfully: %+v", solar)
	} else {
		logutil.Debugf("[AGGREGATOR] ✗ Solar fetch failed: %v", solarErr)
		if data.Errors != nil && solarErr != nil {
			data.Errors["KNMI Solar"] = solarErr.Error()
		}
	}

	// Soil Data
	if soil, err := pa.apiClient.FetchWURSoilData(pa.config, lat, lon); err == nil {
		data.SoilData = soil
		data.DataSources = append(data.DataSources, "WUR Soil")
	} else {
		logutil.Debugf("[AGGREGATOR] WUR Soil fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["WUR Soil"] = err.Error()
		}
	}

	// Subsidence
	if subsidence, err := pa.apiClient.FetchSubsidenceData(pa.config, lat, lon); err == nil {
		data.Subsidence = subsidence
		data.DataSources = append(data.DataSources, "SkyGeo")
	} else {
		logutil.Debugf("[AGGREGATOR] SkyGeo fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["SkyGeo Subsidence"] = err.Error()
		}
	}

	// Soil Quality
	if soilQuality, err := pa.apiClient.FetchSoilQualityData(pa.config, lat, lon); err == nil {
		data.SoilQuality = soilQuality
		data.DataSources = append(data.DataSources, "Soil Quality")
	} else {
		logutil.Debugf("[AGGREGATOR] Soil Quality fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Soil Quality"] = err.Error()
		}
	}

	// BRO Soil Map
	if broSoil, err := pa.apiClient.FetchBROSoilMapData(pa.config, lat, lon); err == nil {
		data.BROSoilMap = broSoil
		data.DataSources = append(data.DataSources, "BRO")
	} else {
		logutil.Debugf("[AGGREGATOR] BRO Soil fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["BRO Soil Map"] = err.Error()
		}
	}

	// Air Quality
	logutil.Debugf("[AGGREGATOR] Calling FetchAirQualityData...")
	if airQuality, err := pa.apiClient.FetchAirQualityData(pa.config, lat, lon); err == nil {
		data.AirQuality = airQuality
		data.DataSources = append(data.DataSources, "Air Quality")
		logutil.Debugf("[AGGREGATOR] ✓ Air quality data fetched successfully")
	} else {
		logutil.Debugf("[AGGREGATOR] ✗ Air quality fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Luchtmeetnet Air Quality"] = err.Error()
		}
	}

	// Noise Pollution
	if noise, err := pa.apiClient.FetchNoisePollutionData(pa.config, lat, lon); err == nil {
		data.NoisePollution = noise
		data.DataSources = append(data.DataSources, "Noise Register")
	} else {
		logutil.Debugf("[AGGREGATOR] Noise Pollution fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Noise Pollution"] = err.Error()
		}
	}
}

func (pa *PropertyAggregator) fetchEnergyData(data *ComprehensivePropertyData, bagID string) {
	// Energy & Climate
	if energy, err := pa.apiClient.FetchEnergyClimateData(pa.config, bagID); err == nil {
		data.EnergyClimate = energy
		data.DataSources = append(data.DataSources, "Altum Energy")
	} else {
		logutil.Debugf("[AGGREGATOR] Altum Energy fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Altum Energy"] = err.Error()
		}
	}

	// Sustainability
	if sustainability, err := pa.apiClient.FetchSustainabilityData(pa.config, bagID); err == nil {
		data.Sustainability = sustainability
		data.DataSources = append(data.DataSources, "Altum Sustainability")
	} else {
		logutil.Debugf("[AGGREGATOR] Altum Sustainability fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Altum Sustainability"] = err.Error()
		}
	}
}

func (pa *PropertyAggregator) fetchRiskData(data *ComprehensivePropertyData, lat, lon float64, neighborhoodCode string) {
	// Flood Risk
	if flood, err := pa.apiClient.FetchFloodRiskData(pa.config, lat, lon); err == nil {
		data.FloodRisk = flood
		data.DataSources = append(data.DataSources, "Flood Risk")
	} else {
		logutil.Debugf("[AGGREGATOR] Flood risk fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Flood Risk"] = err.Error()
		}
	}

	// Water Quality
	if water, err := pa.apiClient.FetchWaterQualityData(pa.config, lat, lon); err == nil {
		data.WaterQuality = water
		data.DataSources = append(data.DataSources, "Digital Delta")
	} else {
		logutil.Debugf("[AGGREGATOR] Water quality fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Water Quality"] = err.Error()
		}
	}

	// Safety (uses neighborhood code if available)
	if neighborhoodCode != "" {
		if safety, err := pa.apiClient.FetchSafetyData(pa.config, neighborhoodCode); err == nil {
			data.Safety = safety
			data.DataSources = append(data.DataSources, "CBS Safety")
		} else {
			logutil.Debugf("[AGGREGATOR] Safety fetch failed: %v", err)
			if data.Errors != nil {
				data.Errors["CBS Safety"] = err.Error()
			}
		}
	} else {
		logutil.Debugf("[AGGREGATOR] Skipping safety data: no neighborhood code available")
		if data.Errors != nil {
			data.Errors["CBS Safety"] = "neighborhood code not available"
		}
	}

	// Schiphol Flights
	if flights, err := pa.apiClient.FetchSchipholFlightData(pa.config, lat, lon); err == nil {
		data.SchipholFlights = flights
		data.DataSources = append(data.DataSources, "Schiphol")
	} else {
		logutil.Debugf("[AGGREGATOR] Schiphol flight data fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Schiphol"] = err.Error()
		}
	}
}

func (pa *PropertyAggregator) fetchMobilityData(data *ComprehensivePropertyData, lat, lon float64) {
	// Traffic Data
	if traffic, err := pa.apiClient.FetchNDWTrafficData(pa.config, lat, lon, 1000); err == nil {
		data.TrafficData = traffic
		data.DataSources = append(data.DataSources, "NDW Traffic")
	} else {
		logutil.Debugf("[AGGREGATOR] NDW Traffic fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["NDW Traffic"] = err.Error()
		}
	}

	// Public Transport
	if transport, err := pa.apiClient.FetchOpenOVData(pa.config, lat, lon); err == nil {
		data.PublicTransport = transport
		data.DataSources = append(data.DataSources, "OpenOV")
	} else {
		logutil.Debugf("[AGGREGATOR] OpenOV fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["OpenOV"] = err.Error()
		}
	}

	// Parking
	if parking, err := pa.apiClient.FetchParkingData(pa.config, lat, lon, 500); err == nil {
		data.ParkingData = parking
		data.DataSources = append(data.DataSources, "Parking")
	} else {
		logutil.Debugf("[AGGREGATOR] Parking fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Parking"] = err.Error()
		}
	}
}

func (pa *PropertyAggregator) fetchDemographicsData(data *ComprehensivePropertyData, lat, lon float64, neighborhoodCode, regionCode string) {
	// Population
	if population, err := pa.apiClient.FetchCBSPopulationData(pa.config, lat, lon); err == nil {
		data.Population = population
		data.DataSources = append(data.DataSources, "CBS Population")
	} else {
		logutil.Debugf("[AGGREGATOR] Population fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["CBS Population"] = err.Error()
		}
	}

	// Square Stats
	if squareStats, err := pa.apiClient.FetchCBSSquareStats(pa.config, lat, lon); err == nil {
		data.SquareStats = squareStats
		data.DataSources = append(data.DataSources, "CBS Square Stats")
	} else {
		logutil.Debugf("[AGGREGATOR] Square stats fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["CBS Square Stats"] = err.Error()
		}
	}

	// StatLine (uses region/municipality code if available)
	if regionCode != "" {
		if statLine, err := pa.apiClient.FetchCBSStatLineData(pa.config, regionCode); err == nil {
			data.StatLineData = statLine
			data.DataSources = append(data.DataSources, "CBS StatLine")
		} else {
			logutil.Debugf("[AGGREGATOR] StatLine fetch failed: %v", err)
			if data.Errors != nil {
				data.Errors["CBS StatLine"] = err.Error()
			}
		}
	} else {
		logutil.Debugf("[AGGREGATOR] Skipping StatLine data: no region code available")
		if data.Errors != nil {
			data.Errors["CBS StatLine"] = "region code not available"
		}
	}

	// Legacy CBS Data (uses neighborhood code if available)
	if neighborhoodCode != "" {
		if cbsData, err := pa.apiClient.FetchCBSData(pa.config, neighborhoodCode); err == nil {
			data.CBSData = cbsData
			data.DataSources = append(data.DataSources, "CBS")
		} else {
			logutil.Debugf("[AGGREGATOR] CBS data fetch failed: %v", err)
			if data.Errors != nil {
				data.Errors["CBS"] = err.Error()
			}
		}
	} else {
		logutil.Debugf("[AGGREGATOR] Skipping CBS data: no neighborhood code available")
		if data.Errors != nil {
			data.Errors["CBS"] = "neighborhood code not available"
		}
	}
}

func (pa *PropertyAggregator) fetchInfrastructureData(data *ComprehensivePropertyData, lat, lon float64) {
	// Green Spaces
	if greenSpaces, err := pa.apiClient.FetchGreenSpacesData(pa.config, lat, lon, 1000); err == nil {
		data.GreenSpaces = greenSpaces
		data.DataSources = append(data.DataSources, "Green Spaces")
	} else {
		logutil.Debugf("[AGGREGATOR] Green Spaces fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Green Spaces"] = err.Error()
		}
	}

	// Education
	if education, err := pa.apiClient.FetchEducationData(pa.config, lat, lon); err == nil {
		data.Education = education
		data.DataSources = append(data.DataSources, "Education")
	} else {
		logutil.Debugf("[AGGREGATOR] Education fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Education"] = err.Error()
		}
	}

	// Building Permits
	if permits, err := pa.apiClient.FetchBuildingPermitsData(pa.config, lat, lon, 1000); err == nil {
		data.BuildingPermits = permits
		data.DataSources = append(data.DataSources, "Building Permits")
	} else {
		logutil.Debugf("[AGGREGATOR] Building Permits fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Building Permits"] = err.Error()
		}
	}

	// Facilities
	if facilities, err := pa.apiClient.FetchFacilitiesData(pa.config, lat, lon); err == nil {
		data.Facilities = facilities
		data.DataSources = append(data.DataSources, "Facilities")
	} else {
		logutil.Debugf("[AGGREGATOR] Facilities fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Facilities"] = err.Error()
		}
	}

	// Elevation (AHN)
	if elevation, err := pa.apiClient.FetchAHNHeightData(pa.config, lat, lon); err == nil {
		data.Elevation = elevation
		data.DataSources = append(data.DataSources, "AHN")
	} else {
		logutil.Debugf("[AGGREGATOR] AHN Elevation fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["AHN Elevation"] = err.Error()
		}
	}
}

func (pa *PropertyAggregator) fetchPlatformData(data *ComprehensivePropertyData, lat, lon float64) {
	// PDOK Platform
	if pdok, err := pa.apiClient.FetchPDOKPlatformData(pa.config, lat, lon); err == nil {
		data.PDOKData = pdok
		data.DataSources = append(data.DataSources, "PDOK Platform")
	} else {
		logutil.Debugf("[AGGREGATOR] PDOK Platform fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["PDOK Platform"] = err.Error()
		}
	}

	// Stratopo Environment
	if stratopo, err := pa.apiClient.FetchStratopoEnvironmentData(pa.config, lat, lon); err == nil {
		data.StratopoEnvironment = stratopo
		data.DataSources = append(data.DataSources, "Stratopo")
	} else {
		logutil.Debugf("[AGGREGATOR] Stratopo fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Stratopo"] = err.Error()
		}
	}

	// Land Use
	if landUse, err := pa.apiClient.FetchLandUseData(pa.config, lat, lon); err == nil {
		data.LandUse = landUse
		data.DataSources = append(data.DataSources, "Land Use")
	} else {
		logutil.Debugf("[AGGREGATOR] Land Use fetch failed: %v", err)
		if data.Errors != nil {
			data.Errors["Land Use"] = err.Error()
		}
	}
}
