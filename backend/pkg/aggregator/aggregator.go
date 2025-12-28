package aggregator

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/iman-hussain/AddressIQ/backend/pkg/apiclient"
	"github.com/iman-hussain/AddressIQ/backend/pkg/cache"
	"github.com/iman-hussain/AddressIQ/backend/pkg/config"
	"github.com/iman-hussain/AddressIQ/backend/pkg/logutil"
	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
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
	KadasterInfo       *models.KadasterObjectInfo     `json:"kadasterInfo,omitempty"`
	WOZData            *models.AltumWOZData           `json:"wozData,omitempty"`
	MarketValuation    *models.MatrixianPropertyValue `json:"marketValuation,omitempty"`
	TransactionHistory *models.TransactionHistory     `json:"transactionHistory,omitempty"`
	MonumentStatus     *models.MonumentData           `json:"monumentStatus,omitempty"`

	// Environmental Data
	// Environmental Data
	Weather        *models.KNMIWeatherData    `json:"weather,omitempty"`
	SolarPotential *models.KNMISolarData      `json:"solarPotential,omitempty"`
	SoilData       *models.WURSoilData        `json:"soilData,omitempty"`
	Subsidence     *models.SubsidenceData     `json:"subsidence,omitempty"`
	SoilQuality    *models.SoilQualityData    `json:"soilQuality,omitempty"`
	BROSoilMap     *models.BROSoilMapData     `json:"broSoilMap,omitempty"`
	AirQuality     *models.AirQualityData     `json:"airQuality,omitempty"`
	NoisePollution *models.NoisePollutionData `json:"noisePollution,omitempty"`

	// Energy & Sustainability
	EnergyClimate  *models.EnergyClimateData  `json:"energyClimate,omitempty"`
	Sustainability *models.SustainabilityData `json:"sustainability,omitempty"`

	// Risk Assessment
	FloodRisk    *models.FloodRiskData    `json:"floodRisk,omitempty"`
	WaterQuality *models.WaterQualityData `json:"waterQuality,omitempty"`
	Safety       *models.SafetyData       `json:"safety,omitempty"`

	// Mobility & Accessibility
	TrafficData     []models.NDWTrafficData     `json:"trafficData,omitempty"`
	PublicTransport *models.OpenOVTransportData `json:"publicTransport,omitempty"`
	ParkingData     *models.ParkingData         `json:"parkingData,omitempty"`

	// Demographics & Neighborhood
	Population   *models.CBSPopulationData  `json:"population,omitempty"`
	StatLineData *models.CBSStatLineData    `json:"statLineData,omitempty"`
	SquareStats  *models.CBSSquareStatsData `json:"squareStats,omitempty"`
	CBSData      *models.CBSData            `json:"cbsData,omitempty"`

	// Infrastructure
	// Infrastructure
	GreenSpaces     *models.GreenSpacesData     `json:"greenSpaces,omitempty"`
	Education       *models.EducationData       `json:"education,omitempty"`
	BuildingPermits *models.BuildingPermitsData `json:"buildingPermits,omitempty"`
	Facilities      *models.FacilitiesData      `json:"facilities,omitempty"`
	Elevation       *models.AHNHeightData       `json:"elevation,omitempty"`

	// Comprehensive Platforms
	PDOKData            *models.PDOKPlatformData        `json:"pdokData,omitempty"`
	StratopoEnvironment *models.StratopoEnvironmentData `json:"stratopoEnvironment,omitempty"`
	LandUse             *models.LandUseData             `json:"landUse,omitempty"`

	// Aviation
	SchipholFlights *models.SchipholFlightData `json:"schipholFlights,omitempty"`

	// AI Summary
	// AI Summary
	AISummary *models.GeminiSummary `json:"aiSummary,omitempty"`

	// Metadata
	AggregatedAt time.Time         `json:"aggregatedAt"`
	DataSources  []string          `json:"dataSources"`
	Errors       map[string]string `json:"errors,omitempty"`
}

// ProgressEvent represents a progress update during aggregation
type ProgressEvent struct {
	Source        string `json:"source"`
	Status        string `json:"status"` // "success", "error", "skipped"
	Completed     int    `json:"completed"`
	Total         int    `json:"total"`
	LastCompleted string `json:"lastCompleted"`
}

// AggregatePropertyData fetches and combines data from all available sources
func (pa *PropertyAggregator) AggregatePropertyData(ctx context.Context, postcode, houseNumber string) (*ComprehensivePropertyData, error) {
	return pa.AggregatePropertyDataWithOptions(ctx, postcode, houseNumber, false, nil)
}

// AggregatePropertyDataWithOptions fetches and combines data from all available sources with cache bypass option and progress reporting
func (pa *PropertyAggregator) AggregatePropertyDataWithOptions(ctx context.Context, postcode, houseNumber string, bypassCache bool, progressCh chan<- ProgressEvent) (*ComprehensivePropertyData, error) {
	logutil.Debugf("[AGGREGATOR] Starting aggregation for %s %s", postcode, houseNumber)

	// Check cache first (if available and not bypassed)
	if pa.cache != nil && !bypassCache {
		cacheKey := cache.CacheKey{}.AggregatedKey(postcode, houseNumber)
		var cached ComprehensivePropertyData
		if err := pa.cache.Get(cacheKey, &cached); err == nil {
			logutil.Debugf("[AGGREGATOR] Cache hit for %s %s - returning cached data", postcode, houseNumber)
			return &cached, nil
		}
		logutil.Debugf("[AGGREGATOR] Cache miss for %s %s - fetching fresh data", postcode, houseNumber)
	} else if bypassCache {
		logutil.Debugf("[AGGREGATOR] Cache bypass requested for %s %s - fetching fresh data", postcode, houseNumber)
	}

	// Start with BAG data (essential) - this must be done sequentially as other data depends on it
	bagData, err := pa.apiClient.FetchBAGData(ctx, postcode, houseNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch BAG data: %w", err)
	}

	lat := bagData.Coordinates[1]
	lon := bagData.Coordinates[0]

	logutil.Debugf("[AGGREGATOR] Coordinates from BAG: lat=%.6f, lon=%.6f", lat, lon)

	// Extract BAG ID from response
	bagID := bagData.ID
	if bagID == "" {
		logutil.Debugf("[AGGREGATOR] Warning: No BAG ID found in response, some APIs may fail")
	}

	// Lookup neighborhood and region codes dynamically
	// This is also fast and essential for other calls, so we do it synchronously (or could be parallel, but simpler here)
	// Actually, let's keep it sync for simplicity of dependency management
	var neighborhoodCode, regionCode string
	regionCodes, err := pa.apiClient.LookupNeighborhoodCode(ctx, pa.config, lat, lon)
	if err == nil && regionCodes != nil {
		neighborhoodCode = regionCodes.NeighborhoodCode
		regionCode = regionCodes.MunicipalityCode
		logutil.Debugf("[AGGREGATOR] Resolved neighborhood=%s, region=%s", neighborhoodCode, regionCode)
	} else {
		logutil.Debugf("[AGGREGATOR] Warning: Could not resolve neighborhood codes: %v", err)
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

	// Total expected sources (approximate)
	const totalSources = 33
	var completedSources atomic.Int32

	// Progress callback
	reportProgress := func(source, status string) {
		newCompleted := completedSources.Add(1)
		logutil.Debugf("[AGGREGATOR] Progress: %s (%s) - %d/%d", source, status, newCompleted, totalSources)

		if progressCh != nil {
			// Non-blocking send to avoid stalling if channel is full/slow
			select {
			case progressCh <- ProgressEvent{
				Source:        source,
				Status:        status,
				Completed:     int(newCompleted),
				Total:         totalSources,
				LastCompleted: source,
			}:
			default:
				logutil.Debugf("[AGGREGATOR] Progress channel full, dropping event for %s", source)
			}
		}
	}

	// Mutex for thread-safe writes to DataSources and Errors
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Helper to handle panic recovery in goroutines
	safeGo := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logutil.Debugf("[AGGREGATOR] Panic recovered in goroutine: %v", r)
				}
			}()
			fn()
		}()
	}

	// Launch parallel fetchers

	// Group 1: Property Data
	safeGo(func() { pa.fetchPropertyData(ctx, &mu, data, bagID, lat, lon, reportProgress) })

	// Group 2: Environmental Data
	safeGo(func() { pa.fetchEnvironmentalData(ctx, &mu, data, lat, lon, reportProgress) })

	// Group 3: Energy Data
	safeGo(func() { pa.fetchEnergyData(ctx, &mu, data, bagID, reportProgress) })

	// Group 4: Risk Data
	safeGo(func() { pa.fetchRiskData(ctx, &mu, data, lat, lon, neighborhoodCode, reportProgress) })

	// Group 5: Mobility Data
	safeGo(func() { pa.fetchMobilityData(ctx, &mu, data, lat, lon, reportProgress) })

	// Group 6: Demographics Data
	safeGo(func() {
		pa.fetchDemographicsData(ctx, &mu, data, lat, lon, neighborhoodCode, regionCode, reportProgress)
	})

	// Group 7: Infrastructure Data
	safeGo(func() { pa.fetchInfrastructureData(ctx, &mu, data, lat, lon, reportProgress) })

	// Group 8: Platform Data
	safeGo(func() { pa.fetchPlatformData(ctx, &mu, data, lat, lon, reportProgress) })

	// Wait for all data fetching to complete
	wg.Wait()

	// AI Summary (Gemini) - called after all data is collected (as it summarizes the data)
	// We pass the context here too
	if aiSummary, err := pa.apiClient.GenerateLocationSummary(ctx, pa.config, data); err == nil {
		data.AISummary = aiSummary
		if aiSummary.Generated {
			// No need for mutex here as we are back to single thread (after wait)
			data.DataSources = append(data.DataSources, "Gemini AI")
			reportProgress("Gemini AI", "success")
		}
	} else {
		logutil.Debugf("[AGGREGATOR] Gemini AI summary failed: %v", err)
		if data.Errors != nil {
			data.Errors["Gemini AI"] = err.Error()
		}
		reportProgress("Gemini AI", "error")
	}

	// Cache the aggregated result (if caching is available)
	if pa.cache != nil {
		cacheKey := cache.CacheKey{}.AggregatedKey(postcode, houseNumber)
		pa.cache.Set(cacheKey, data, cache.PropertyDataTTL)
	}

	return data, nil
}

// safeAppend adds a source to DataSources in a thread-safe way
func safeAppendSource(mu *sync.Mutex, data *ComprehensivePropertyData, source string) {
	mu.Lock()
	defer mu.Unlock()
	data.DataSources = append(data.DataSources, source)
}

// safeRecordError records an error in a thread-safe way
func safeRecordError(mu *sync.Mutex, data *ComprehensivePropertyData, source, err string) {
	mu.Lock()
	defer mu.Unlock()
	if data.Errors == nil {
		data.Errors = make(map[string]string)
	}
	data.Errors[source] = err
}

func (pa *PropertyAggregator) fetchPropertyData(ctx context.Context, mu *sync.Mutex, data *ComprehensivePropertyData, bagID string, lat, lon float64, onProgress func(string, string)) {
	var wg sync.WaitGroup

	// Helper for this sub-group
	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logutil.Debugf("[AGGREGATOR] Panic in fetchPropertyData: %v", r)
				}
			}()
			fn()
		}()
	}

	// Kadaster Object Info
	run(func() {
		if kadasterInfo, err := pa.apiClient.FetchKadasterObjectInfo(ctx, pa.config, bagID); err == nil {
			data.KadasterInfo = kadasterInfo
			safeAppendSource(mu, data, "Kadaster")
			onProgress("Kadaster", "success")
		} else {
			logutil.Debugf("[AGGREGATOR] Kadaster fetch failed: %v", err)
			safeRecordError(mu, data, "Kadaster", err.Error())
			onProgress("Kadaster", "error")
		}
	})

	// WOZ Data
	run(func() {
		if wozData, err := pa.apiClient.FetchAltumWOZData(ctx, pa.config, bagID); err == nil {
			data.WOZData = wozData
			safeAppendSource(mu, data, "Altum WOZ")
			onProgress("Altum WOZ", "success")
		} else {
			logutil.Debugf("[AGGREGATOR] Altum WOZ fetch failed: %v", err)
			safeRecordError(mu, data, "Altum WOZ", err.Error())
			onProgress("Altum WOZ", "error")
		}
	})

	// Market Valuation
	run(func() {
		if valuation, err := pa.apiClient.FetchPropertyValuePlus(ctx, pa.config, bagID, lat, lon); err == nil {
			data.MarketValuation = valuation
			safeAppendSource(mu, data, "Matrixian")
			onProgress("Matrixian", "success")
		} else {
			logutil.Debugf("[AGGREGATOR] Matrixian fetch failed: %v", err)
			safeRecordError(mu, data, "Matrixian", err.Error())
			onProgress("Matrixian", "error")
		}
	})

	// Transaction History
	run(func() {
		if transactions, err := pa.apiClient.FetchTransactionHistory(ctx, pa.config, bagID); err == nil {
			data.TransactionHistory = transactions
			safeAppendSource(mu, data, "Altum Transactions")
			onProgress("Altum Transactions", "success")
		} else {
			logutil.Debugf("[AGGREGATOR] Altum Transactions fetch failed: %v", err)
			safeRecordError(mu, data, "Altum Transactions", err.Error())
			onProgress("Altum Transactions", "error")
		}
	})

	// Monument Status
	run(func() {
		if monument, err := pa.apiClient.FetchMonumentDataByCoords(ctx, pa.config, lat, lon); err == nil {
			data.MonumentStatus = monument
			safeAppendSource(mu, data, "Monument Register")
			onProgress("Monument Register", "success")
		} else {
			logutil.Debugf("[AGGREGATOR] Monument fetch failed: %v", err)
			safeRecordError(mu, data, "Monument Register", err.Error())
			onProgress("Monument Register", "error")
		}
	})

	wg.Wait()
}

func (pa *PropertyAggregator) fetchEnvironmentalData(ctx context.Context, mu *sync.Mutex, data *ComprehensivePropertyData, lat, lon float64, onProgress func(string, string)) {
	var wg sync.WaitGroup
	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logutil.Debugf("[AGGREGATOR] Panic in fetchEnvironmentalData: %v", r)
				}
			}()
			fn()
		}()
	}

	// Weather
	run(func() {
		weather, weatherErr := pa.apiClient.FetchKNMIWeatherData(ctx, pa.config, lat, lon)
		if weatherErr == nil && weather != nil {
			data.Weather = weather
			safeAppendSource(mu, data, "KNMI Weather")
			onProgress("KNMI Weather", "success")
		} else {
			safeRecordError(mu, data, "KNMI Weather", weatherErr.Error())
			onProgress("KNMI Weather", "error")
		}
	})

	// Solar
	run(func() {
		solar, solarErr := pa.apiClient.FetchKNMISolarData(ctx, pa.config, lat, lon)
		if solarErr == nil && solar != nil {
			data.SolarPotential = solar
			safeAppendSource(mu, data, "KNMI Solar")
			onProgress("KNMI Solar", "success")
		} else {
			safeRecordError(mu, data, "KNMI Solar", solarErr.Error())
			onProgress("KNMI Solar", "error")
		}
	})

	// Soil Data
	run(func() {
		if soil, err := pa.apiClient.FetchWURSoilData(ctx, pa.config, lat, lon); err == nil {
			data.SoilData = soil
			safeAppendSource(mu, data, "WUR Soil")
			onProgress("WUR Soil", "success")
		} else {
			safeRecordError(mu, data, "WUR Soil", err.Error())
			onProgress("WUR Soil", "error")
		}
	})

	// Subsidence
	run(func() {
		if subsidence, err := pa.apiClient.FetchSubsidenceData(ctx, pa.config, lat, lon); err == nil {
			data.Subsidence = subsidence
			safeAppendSource(mu, data, "SkyGeo")
			onProgress("SkyGeo", "success")
		} else {
			safeRecordError(mu, data, "SkyGeo Subsidence", err.Error())
			onProgress("SkyGeo", "error")
		}
	})

	// Soil Quality
	run(func() {
		if soilQuality, err := pa.apiClient.FetchSoilQualityData(ctx, pa.config, lat, lon); err == nil {
			data.SoilQuality = soilQuality
			safeAppendSource(mu, data, "Soil Quality")
			onProgress("Soil Quality", "success")
		} else {
			safeRecordError(mu, data, "Soil Quality", err.Error())
			onProgress("Soil Quality", "error")
		}
	})

	// BRO Soil Map
	run(func() {
		if broSoil, err := pa.apiClient.FetchBROSoilMapData(ctx, pa.config, lat, lon); err == nil {
			data.BROSoilMap = broSoil
			safeAppendSource(mu, data, "BRO")
			onProgress("BRO", "success")
		} else {
			safeRecordError(mu, data, "BRO Soil Map", err.Error())
			onProgress("BRO", "error")
		}
	})

	// Air Quality
	run(func() {
		if airQuality, err := pa.apiClient.FetchAirQualityData(ctx, pa.config, lat, lon); err == nil {
			data.AirQuality = airQuality
			safeAppendSource(mu, data, "Air Quality")
			onProgress("Air Quality", "success")
		} else {
			safeRecordError(mu, data, "Luchtmeetnet Air Quality", err.Error())
			onProgress("Air Quality", "error")
		}
	})

	// Noise Pollution
	run(func() {
		if noise, err := pa.apiClient.FetchNoisePollutionData(ctx, pa.config, lat, lon); err == nil {
			data.NoisePollution = noise
			safeAppendSource(mu, data, "Noise Register")
			onProgress("Noise Register", "success")
		} else {
			safeRecordError(mu, data, "Noise Pollution", err.Error())
			onProgress("Noise Register", "error")
		}
	})

	wg.Wait()
}

func (pa *PropertyAggregator) fetchEnergyData(ctx context.Context, mu *sync.Mutex, data *ComprehensivePropertyData, bagID string, onProgress func(string, string)) {
	var wg sync.WaitGroup
	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logutil.Debugf("[AGGREGATOR] Panic in fetchEnergyData: %v", r)
				}
			}()
			fn()
		}()
	}

	// Energy & Climate
	run(func() {
		if energy, err := pa.apiClient.FetchEnergyClimateData(ctx, pa.config, bagID); err == nil {
			data.EnergyClimate = energy
			safeAppendSource(mu, data, "Altum Energy")
			onProgress("Altum Energy", "success")
		} else {
			safeRecordError(mu, data, "Altum Energy", err.Error())
			onProgress("Altum Energy", "error")
		}
	})

	// Sustainability
	run(func() {
		if sustainability, err := pa.apiClient.FetchSustainabilityData(ctx, pa.config, bagID); err == nil {
			data.Sustainability = sustainability
			safeAppendSource(mu, data, "Altum Sustainability")
			onProgress("Altum Sustainability", "success")
		} else {
			safeRecordError(mu, data, "Altum Sustainability", err.Error())
			onProgress("Altum Sustainability", "error")
		}
	})

	wg.Wait()
}

func (pa *PropertyAggregator) fetchRiskData(ctx context.Context, mu *sync.Mutex, data *ComprehensivePropertyData, lat, lon float64, neighborhoodCode string, onProgress func(string, string)) {
	var wg sync.WaitGroup
	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logutil.Debugf("[AGGREGATOR] Panic in fetchRiskData: %v", r)
				}
			}()
			fn()
		}()
	}

	// Flood Risk
	run(func() {
		if flood, err := pa.apiClient.FetchFloodRiskData(ctx, pa.config, lat, lon); err == nil {
			data.FloodRisk = flood
			safeAppendSource(mu, data, "Flood Risk")
			onProgress("Flood Risk", "success")
		} else {
			safeRecordError(mu, data, "Flood Risk", err.Error())
			onProgress("Flood Risk", "error")
		}
	})

	// Water Quality
	run(func() {
		if water, err := pa.apiClient.FetchWaterQualityData(ctx, pa.config, lat, lon); err == nil {
			data.WaterQuality = water
			safeAppendSource(mu, data, "Digital Delta")
			onProgress("Digital Delta", "success")
		} else {
			safeRecordError(mu, data, "Water Quality", err.Error())
			onProgress("Digital Delta", "error")
		}
	})

	// Safety
	run(func() {
		if neighborhoodCode != "" {
			if safety, err := pa.apiClient.FetchSafetyData(ctx, pa.config, neighborhoodCode); err == nil {
				data.Safety = safety
				safeAppendSource(mu, data, "CBS Safety")
				onProgress("CBS Safety", "success")
			} else {
				safeRecordError(mu, data, "CBS Safety", err.Error())
				onProgress("CBS Safety", "error")
			}
		} else {
			safeRecordError(mu, data, "CBS Safety", "neighborhood code not available")
			onProgress("CBS Safety", "skipped")
		}
	})

	// Schiphol Flights
	run(func() {
		if flights, err := pa.apiClient.FetchSchipholFlightData(ctx, pa.config, lat, lon); err == nil {
			data.SchipholFlights = flights
			safeAppendSource(mu, data, "Schiphol")
			onProgress("Schiphol", "success")
		} else {
			safeRecordError(mu, data, "Schiphol", err.Error())
			onProgress("Schiphol", "error")
		}
	})

	wg.Wait()
}

func (pa *PropertyAggregator) fetchMobilityData(ctx context.Context, mu *sync.Mutex, data *ComprehensivePropertyData, lat, lon float64, onProgress func(string, string)) {
	var wg sync.WaitGroup
	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logutil.Debugf("[AGGREGATOR] Panic in fetchMobilityData: %v", r)
				}
			}()
			fn()
		}()
	}

	// Traffic Data
	run(func() {
		if traffic, err := pa.apiClient.FetchNDWTrafficData(ctx, pa.config, lat, lon, 1000); err == nil {
			data.TrafficData = traffic
			safeAppendSource(mu, data, "NDW Traffic")
			onProgress("NDW Traffic", "success")
		} else {
			safeRecordError(mu, data, "NDW Traffic", err.Error())
			onProgress("NDW Traffic", "error")
		}
	})

	// Public Transport
	run(func() {
		if transport, err := pa.apiClient.FetchOpenOVData(ctx, pa.config, lat, lon); err == nil {
			data.PublicTransport = transport
			safeAppendSource(mu, data, "OpenOV")
			onProgress("OpenOV", "success")
		} else {
			safeRecordError(mu, data, "OpenOV", err.Error())
			onProgress("OpenOV", "error")
		}
	})

	// Parking
	run(func() {
		if parking, err := pa.apiClient.FetchParkingData(ctx, pa.config, lat, lon, 500); err == nil {
			data.ParkingData = parking
			safeAppendSource(mu, data, "Parking")
			onProgress("Parking", "success")
		} else {
			safeRecordError(mu, data, "Parking", err.Error())
			onProgress("Parking", "error")
		}
	})

	wg.Wait()
}

func (pa *PropertyAggregator) fetchDemographicsData(ctx context.Context, mu *sync.Mutex, data *ComprehensivePropertyData, lat, lon float64, neighborhoodCode, regionCode string, onProgress func(string, string)) {
	var wg sync.WaitGroup
	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logutil.Debugf("[AGGREGATOR] Panic in fetchDemographicsData: %v", r)
				}
			}()
			fn()
		}()
	}

	// Population
	run(func() {
		if population, err := pa.apiClient.FetchCBSPopulationData(ctx, pa.config, lat, lon); err == nil {
			data.Population = population
			safeAppendSource(mu, data, "CBS Population")
			onProgress("CBS Population", "success")
		} else {
			safeRecordError(mu, data, "CBS Population", err.Error())
			onProgress("CBS Population", "error")
		}
	})

	// Square Stats
	run(func() {
		if squareStats, err := pa.apiClient.FetchCBSSquareStats(ctx, pa.config, lat, lon); err == nil {
			data.SquareStats = squareStats
			safeAppendSource(mu, data, "CBS Square Stats")
			onProgress("CBS Square Stats", "success")
		} else {
			safeRecordError(mu, data, "CBS Square Stats", err.Error())
			onProgress("CBS Square Stats", "error")
		}
	})

	// StatLine
	run(func() {
		if regionCode != "" {
			if statLine, err := pa.apiClient.FetchCBSStatLineData(ctx, pa.config, regionCode); err == nil {
				data.StatLineData = statLine
				safeAppendSource(mu, data, "CBS StatLine")
				onProgress("CBS StatLine", "success")
			} else {
				safeRecordError(mu, data, "CBS StatLine", err.Error())
				onProgress("CBS StatLine", "error")
			}
		} else {
			safeRecordError(mu, data, "CBS StatLine", "region code not available")
			onProgress("CBS StatLine", "skipped")
		}
	})

	// Legacy CBS Data
	run(func() {
		if neighborhoodCode != "" {
			if cbsData, err := pa.apiClient.FetchCBSData(ctx, pa.config, neighborhoodCode); err == nil {
				data.CBSData = cbsData
				safeAppendSource(mu, data, "CBS")
				onProgress("CBS", "success")
			} else {
				safeRecordError(mu, data, "CBS", err.Error())
				onProgress("CBS", "error")
			}
		} else {
			safeRecordError(mu, data, "CBS", "neighborhood code not available")
			onProgress("CBS", "skipped")
		}
	})

	wg.Wait()
}

func (pa *PropertyAggregator) fetchInfrastructureData(ctx context.Context, mu *sync.Mutex, data *ComprehensivePropertyData, lat, lon float64, onProgress func(string, string)) {
	var wg sync.WaitGroup
	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logutil.Debugf("[AGGREGATOR] Panic in fetchInfrastructureData: %v", r)
				}
			}()
			fn()
		}()
	}

	// Green Spaces
	run(func() {
		if greenSpaces, err := pa.apiClient.FetchGreenSpacesData(ctx, pa.config, lat, lon, 1000); err == nil {
			data.GreenSpaces = greenSpaces
			safeAppendSource(mu, data, "Green Spaces")
			onProgress("Green Spaces", "success")
		} else {
			safeRecordError(mu, data, "Green Spaces", err.Error())
			onProgress("Green Spaces", "error")
		}
	})

	// Education
	run(func() {
		if education, err := pa.apiClient.FetchEducationData(ctx, pa.config, lat, lon); err == nil {
			data.Education = education
			safeAppendSource(mu, data, "Education")
			onProgress("Education", "success")
		} else {
			safeRecordError(mu, data, "Education", err.Error())
			onProgress("Education", "error")
		}
	})

	// Building Permits
	run(func() {
		if permits, err := pa.apiClient.FetchBuildingPermitsData(ctx, pa.config, lat, lon, 1000); err == nil {
			data.BuildingPermits = permits
			safeAppendSource(mu, data, "Building Permits")
			onProgress("Building Permits", "success")
		} else {
			safeRecordError(mu, data, "Building Permits", err.Error())
			onProgress("Building Permits", "error")
		}
	})

	// Facilities
	run(func() {
		if facilities, err := pa.apiClient.FetchFacilitiesData(ctx, pa.config, lat, lon); err == nil {
			data.Facilities = facilities
			safeAppendSource(mu, data, "Facilities")
			onProgress("Facilities", "success")
		} else {
			safeRecordError(mu, data, "Facilities", err.Error())
			onProgress("Facilities", "error")
		}
	})

	// Elevation (AHN)
	run(func() {
		if elevation, err := pa.apiClient.FetchAHNHeightData(ctx, pa.config, lat, lon); err == nil {
			data.Elevation = elevation
			safeAppendSource(mu, data, "AHN")
			onProgress("AHN", "success")
		} else {
			safeRecordError(mu, data, "AHN Elevation", err.Error())
			onProgress("AHN", "error")
		}
	})

	wg.Wait()
}

func (pa *PropertyAggregator) fetchPlatformData(ctx context.Context, mu *sync.Mutex, data *ComprehensivePropertyData, lat, lon float64, onProgress func(string, string)) {
	var wg sync.WaitGroup
	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logutil.Debugf("[AGGREGATOR] Panic in fetchPlatformData: %v", r)
				}
			}()
			fn()
		}()
	}

	// PDOK Platform
	run(func() {
		if pdok, err := pa.apiClient.FetchPDOKPlatformData(ctx, pa.config, lat, lon); err == nil {
			data.PDOKData = pdok
			safeAppendSource(mu, data, "PDOK Platform")
			onProgress("PDOK Platform", "success")
		} else {
			safeRecordError(mu, data, "PDOK Platform", err.Error())
			onProgress("PDOK Platform", "error")
		}
	})

	// Stratopo Environment
	run(func() {
		if stratopo, err := pa.apiClient.FetchStratopoEnvironmentData(ctx, pa.config, lat, lon); err == nil {
			data.StratopoEnvironment = stratopo
			safeAppendSource(mu, data, "Stratopo")
			onProgress("Stratopo", "success")
		} else {
			safeRecordError(mu, data, "Stratopo", err.Error())
			onProgress("Stratopo", "error")
		}
	})

	// Land Use
	run(func() {
		if landUse, err := pa.apiClient.FetchLandUseData(ctx, pa.config, lat, lon); err == nil {
			data.LandUse = landUse
			safeAppendSource(mu, data, "Land Use")
			onProgress("Land Use", "success")
		} else {
			safeRecordError(mu, data, "Land Use", err.Error())
			onProgress("Land Use", "error")
		}
	})

	wg.Wait()
}
