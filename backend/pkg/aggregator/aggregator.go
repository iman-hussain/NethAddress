package aggregator

import (
	"fmt"
	"time"

	"github.com/iman-hussain/AddressIQ/backend/pkg/apiclient"
	"github.com/iman-hussain/AddressIQ/backend/pkg/cache"
	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
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
	AggregatedAt time.Time `json:"aggregatedAt"`
	DataSources  []string  `json:"dataSources"`
}

// AggregatePropertyData fetches and combines data from all available sources
func (pa *PropertyAggregator) AggregatePropertyData(postcode, houseNumber string) (*ComprehensivePropertyData, error) {
	// Check cache first (if available)
	if pa.cache != nil {
		cacheKey := cache.CacheKey{}.AggregatedKey(postcode, houseNumber)
		var cached ComprehensivePropertyData
		if err := pa.cache.Get(cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	// Start with BAG data (essential)
	bagData, err := pa.apiClient.FetchBAGData(postcode, houseNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch BAG data: %w", err)
	}

	lat := bagData.Coordinates[1]
	lon := bagData.Coordinates[0]

	// Extract BAG ID from response (would need to be added to BAG API response)
	bagID := "extracted-bag-id" // TODO: Get actual BAG ID from response

	data := &ComprehensivePropertyData{
		Address:      bagData.Address,
		Coordinates:  bagData.Coordinates,
		BAGID:        bagID,
		AggregatedAt: time.Now(),
		DataSources:  []string{"BAG"},
	}

	// Property & Valuation Data (parallel fetches where possible)
	pa.fetchPropertyData(data, bagID, lat, lon)

	// Environmental Data
	pa.fetchEnvironmentalData(data, lat, lon)

	// Energy & Sustainability
	pa.fetchEnergyData(data, bagID)

	// Risk Assessment
	pa.fetchRiskData(data, lat, lon)

	// Mobility & Accessibility
	pa.fetchMobilityData(data, lat, lon)

	// Demographics & Neighborhood
	pa.fetchDemographicsData(data, lat, lon)

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
	}

	// WOZ Data
	if wozData, err := pa.apiClient.FetchAltumWOZData(pa.config, bagID); err == nil {
		data.WOZData = wozData
		data.DataSources = append(data.DataSources, "Altum WOZ")
	}

	// Market Valuation
	if valuation, err := pa.apiClient.FetchPropertyValuePlus(pa.config, bagID, lat, lon); err == nil {
		data.MarketValuation = valuation
		data.DataSources = append(data.DataSources, "Matrixian")
	}

	// Transaction History
	if transactions, err := pa.apiClient.FetchTransactionHistory(pa.config, bagID); err == nil {
		data.TransactionHistory = transactions
		data.DataSources = append(data.DataSources, "Altum Transactions")
	}

	// Monument Status
	if monument, err := pa.apiClient.FetchMonumentData(pa.config, bagID); err == nil {
		data.MonumentStatus = monument
		data.DataSources = append(data.DataSources, "Monument Register")
	}
}

func (pa *PropertyAggregator) fetchEnvironmentalData(data *ComprehensivePropertyData, lat, lon float64) {
	// Weather
	if weather, err := pa.apiClient.FetchKNMIWeatherData(pa.config, lat, lon); err == nil {
		data.Weather = weather
		data.DataSources = append(data.DataSources, "KNMI Weather")
	}

	// Solar Potential
	if solar, err := pa.apiClient.FetchKNMISolarData(pa.config, lat, lon); err == nil {
		data.SolarPotential = solar
		data.DataSources = append(data.DataSources, "KNMI Solar")
	}

	// Soil Data
	if soil, err := pa.apiClient.FetchWURSoilData(pa.config, lat, lon); err == nil {
		data.SoilData = soil
		data.DataSources = append(data.DataSources, "WUR Soil")
	}

	// Subsidence
	if subsidence, err := pa.apiClient.FetchSubsidenceData(pa.config, lat, lon); err == nil {
		data.Subsidence = subsidence
		data.DataSources = append(data.DataSources, "SkyGeo")
	}

	// Soil Quality
	if soilQuality, err := pa.apiClient.FetchSoilQualityData(pa.config, lat, lon); err == nil {
		data.SoilQuality = soilQuality
		data.DataSources = append(data.DataSources, "Soil Quality")
	}

	// BRO Soil Map
	if broSoil, err := pa.apiClient.FetchBROSoilMapData(pa.config, lat, lon); err == nil {
		data.BROSoilMap = broSoil
		data.DataSources = append(data.DataSources, "BRO")
	}

	// Air Quality
	if airQuality, err := pa.apiClient.FetchAirQualityData(pa.config, lat, lon); err == nil {
		data.AirQuality = airQuality
		data.DataSources = append(data.DataSources, "Luchtmeetnet")
	}

	// Noise Pollution
	if noise, err := pa.apiClient.FetchNoisePollutionData(pa.config, lat, lon); err == nil {
		data.NoisePollution = noise
		data.DataSources = append(data.DataSources, "Noise Register")
	}
}

func (pa *PropertyAggregator) fetchEnergyData(data *ComprehensivePropertyData, bagID string) {
	// Energy & Climate
	if energy, err := pa.apiClient.FetchEnergyClimateData(pa.config, bagID); err == nil {
		data.EnergyClimate = energy
		data.DataSources = append(data.DataSources, "Altum Energy")
	}

	// Sustainability
	if sustainability, err := pa.apiClient.FetchSustainabilityData(pa.config, bagID); err == nil {
		data.Sustainability = sustainability
		data.DataSources = append(data.DataSources, "Altum Sustainability")
	}
}

func (pa *PropertyAggregator) fetchRiskData(data *ComprehensivePropertyData, lat, lon float64) {
	// Flood Risk
	if flood, err := pa.apiClient.FetchFloodRiskData(pa.config, lat, lon); err == nil {
		data.FloodRisk = flood
		data.DataSources = append(data.DataSources, "Flood Risk")
	}

	// Water Quality
	if water, err := pa.apiClient.FetchWaterQualityData(pa.config, lat, lon); err == nil {
		data.WaterQuality = water
		data.DataSources = append(data.DataSources, "Digital Delta")
	}

	// Safety (needs neighborhood code - would extract from CBS data)
	// Placeholder for now
	neighborhoodCode := "BU00000000" // TODO: Extract from CBS or PDOK data
	if safety, err := pa.apiClient.FetchSafetyData(pa.config, neighborhoodCode); err == nil {
		data.Safety = safety
		data.DataSources = append(data.DataSources, "CBS Safety")
	}

	// Schiphol Flights
	if flights, err := pa.apiClient.FetchSchipholFlightData(pa.config, lat, lon); err == nil {
		data.SchipholFlights = flights
		data.DataSources = append(data.DataSources, "Schiphol")
	}
}

func (pa *PropertyAggregator) fetchMobilityData(data *ComprehensivePropertyData, lat, lon float64) {
	// Traffic Data
	if traffic, err := pa.apiClient.FetchNDWTrafficData(pa.config, lat, lon, 1000); err == nil {
		data.TrafficData = traffic
		data.DataSources = append(data.DataSources, "NDW Traffic")
	}

	// Public Transport
	if transport, err := pa.apiClient.FetchOpenOVData(pa.config, lat, lon); err == nil {
		data.PublicTransport = transport
		data.DataSources = append(data.DataSources, "OpenOV")
	}

	// Parking
	if parking, err := pa.apiClient.FetchParkingData(pa.config, lat, lon, 500); err == nil {
		data.ParkingData = parking
		data.DataSources = append(data.DataSources, "Parking")
	}
}

func (pa *PropertyAggregator) fetchDemographicsData(data *ComprehensivePropertyData, lat, lon float64) {
	// Population
	if population, err := pa.apiClient.FetchCBSPopulationData(pa.config, lat, lon); err == nil {
		data.Population = population
		data.DataSources = append(data.DataSources, "CBS Population")
	}

	// Square Stats
	if squareStats, err := pa.apiClient.FetchCBSSquareStats(pa.config, lat, lon); err == nil {
		data.SquareStats = squareStats
		data.DataSources = append(data.DataSources, "CBS Square Stats")
	}

	// StatLine (needs region code)
	regionCode := "GM0000" // TODO: Extract from PDOK or other source
	if statLine, err := pa.apiClient.FetchCBSStatLineData(pa.config, regionCode); err == nil {
		data.StatLineData = statLine
		data.DataSources = append(data.DataSources, "CBS StatLine")
	}

	// Legacy CBS Data (if needed)
	neighborhoodCode := "BU00000000" // TODO: Extract appropriately
	if cbsData, err := pa.apiClient.FetchCBSData(pa.config, neighborhoodCode); err == nil {
		data.CBSData = cbsData
		data.DataSources = append(data.DataSources, "CBS")
	}
}

func (pa *PropertyAggregator) fetchInfrastructureData(data *ComprehensivePropertyData, lat, lon float64) {
	// Green Spaces
	if greenSpaces, err := pa.apiClient.FetchGreenSpacesData(pa.config, lat, lon, 1000); err == nil {
		data.GreenSpaces = greenSpaces
		data.DataSources = append(data.DataSources, "Green Spaces")
	}

	// Education
	if education, err := pa.apiClient.FetchEducationData(pa.config, lat, lon); err == nil {
		data.Education = education
		data.DataSources = append(data.DataSources, "Education")
	}

	// Building Permits
	if permits, err := pa.apiClient.FetchBuildingPermitsData(pa.config, lat, lon, 1000); err == nil {
		data.BuildingPermits = permits
		data.DataSources = append(data.DataSources, "Building Permits")
	}

	// Facilities
	if facilities, err := pa.apiClient.FetchFacilitiesData(pa.config, lat, lon); err == nil {
		data.Facilities = facilities
		data.DataSources = append(data.DataSources, "Facilities")
	}

	// Elevation (AHN)
	if elevation, err := pa.apiClient.FetchAHNHeightData(pa.config, lat, lon); err == nil {
		data.Elevation = elevation
		data.DataSources = append(data.DataSources, "AHN")
	}
}

func (pa *PropertyAggregator) fetchPlatformData(data *ComprehensivePropertyData, lat, lon float64) {
	// PDOK Platform
	if pdok, err := pa.apiClient.FetchPDOKPlatformData(pa.config, lat, lon); err == nil {
		data.PDOKData = pdok
		data.DataSources = append(data.DataSources, "PDOK Platform")
	}

	// Stratopo Environment
	if stratopo, err := pa.apiClient.FetchStratopoEnvironmentData(pa.config, lat, lon); err == nil {
		data.StratopoEnvironment = stratopo
		data.DataSources = append(data.DataSources, "Stratopo")
	}

	// Land Use
	if landUse, err := pa.apiClient.FetchLandUseData(pa.config, lat, lon); err == nil {
		data.LandUse = landUse
		data.DataSources = append(data.DataSources, "Land Use")
	}
}
