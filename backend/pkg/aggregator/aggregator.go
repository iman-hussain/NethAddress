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

// GetCachedData retrieves data from cache if available (used for quick checks)
func (pa *PropertyAggregator) GetCachedData(postcode, houseNumber string) (*ComprehensivePropertyData, bool) {
	if pa.cache == nil {
		return nil, false
	}
	cacheKey := cache.CacheKey{}.AggregatedKey(postcode, houseNumber)
	var data ComprehensivePropertyData
	if err := pa.cache.Get(cacheKey, &data); err == nil {
		return &data, true
	}
	return nil, false
}

// ComprehensivePropertyData represents all collected property data
type ComprehensivePropertyData struct {
	// Core Property Data
	Address     string     `json:"address"`
	Coordinates [2]float64 `json:"coordinates"`
	BAGID       string     `json:"bagId"`
	GeoJSON     string     `json:"geojson,omitempty"` // Raw GeoJSON for map display

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
	Source        string      `json:"source"`
	Status        string      `json:"status"` // "success", "error", "skipped"
	Completed     int         `json:"completed"`
	Total         int         `json:"total"`
	LastCompleted string      `json:"lastCompleted"`
	Data          interface{} `json:"data,omitempty"` // Partial data payload
}

// AggregatePropertyData fetches and combines data from all available sources
func (pa *PropertyAggregator) AggregatePropertyData(ctx context.Context, postcode, houseNumber string) (*ComprehensivePropertyData, error) {
	return pa.AggregatePropertyDataWithOptions(ctx, postcode, houseNumber, false, nil, nil)
}

// AggregatePropertyDataWithOptions fetches and combines data from all available sources with cache bypass option and progress reporting
func (pa *PropertyAggregator) AggregatePropertyDataWithOptions(ctx context.Context, postcode, houseNumber string, bypassCache bool, progressCh chan<- ProgressEvent, userKeys map[string]string) (*ComprehensivePropertyData, error) {
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

	// Create a shallow copy of config to apply user provided keys (if any)
	reqConfig := *pa.config
	if userKeys != nil {
		reqConfig.ApplyUserLocalKeys(userKeys)
	}
	cfg := &reqConfig

	// Start with BAG data (essential) - this must be done sequentially as other data depends on it
	bagData, err := pa.apiClient.FetchBAGData(ctx, postcode, houseNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch BAG data: %w", err)
	}

	lat := bagData.Coordinates[1]
	lon := bagData.Coordinates[0]

	// Extract BAG ID from response
	bagID := bagData.ID
	if bagID == "" {
		logutil.Debugf("[AGGREGATOR] Warning: No BAG ID found in response, some APIs may fail")
	}

	// Dynamic neighborhood/region lookup
	var neighborhoodCode, regionCode string
	regionCodes, err := pa.apiClient.LookupNeighborhoodCode(ctx, cfg, lat, lon)
	if err == nil && regionCodes != nil {
		neighborhoodCode = regionCodes.NeighborhoodCode
		regionCode = regionCodes.MunicipalityCode
	} else {
		if bagData.MunicipalityCode != "" {
			regionCode = bagData.MunicipalityCode
		}
	}

	data := &ComprehensivePropertyData{
		Address:      bagData.Address,
		Coordinates:  bagData.Coordinates,
		BAGID:        bagID,
		GeoJSON:      bagData.GeoJSON, // Populate GeoJSON!
		AggregatedAt: time.Now(),
		DataSources:  []string{"BAG"},
		Errors:       make(map[string]string),
	}

	// Dynamic progress tracking
	totalSources := pa.countEnabledSources(cfg)
	var completedSources atomic.Int32
	// Account for BAG being done
	completedSources.Store(1)

	// Progress callback
	reportProgress := func(source, status string, data interface{}) {
		newCompleted := completedSources.Add(1)
		// logutil.Debugf("[AGGREGATOR] Progress: %s (%s) - %d/%d", source, status, newCompleted, totalSources)

		if progressCh != nil {
			select {
			case progressCh <- ProgressEvent{
				Source:        source,
				Status:        status,
				Completed:     int(newCompleted),
				Total:         totalSources,
				LastCompleted: source,
				Data:          data,
			}:
			default:
				// Buffer full, skip update (client eventually gets complete)
			}
		}
	}

	// Concurrency Control
	// Data Protection
	var mu sync.Mutex
	// Limit concurrently active groups to prevent exploding goroutines
	// Flattened concurrency: We use a shared semaphore for ALL individual API calls
	const maxConcurrency = 8
	sem := make(chan struct{}, maxConcurrency)

	// Runner helper to execute tasks with concurrency limit
	// Runner helper to execute tasks with concurrency limit
	runTask := func(phaseWg *sync.WaitGroup, fn func()) {
		phaseWg.Add(1)
		go func() {
			defer phaseWg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release
			defer func() {
				if r := recover(); r != nil {
					logutil.Debugf("[AGGREGATOR] Panic recovered: %v", r)
				}
			}()
			fn()
		}()
	}

	// Launch parallel fetchers - passed runTask to allow them to schedule sub-tasks on the main semaphore
	// Note: runTask is passed to child functions, but we must use it for top-level calls too OR just call them, since they don't do work themselves, just spawn.
	// Actually, the top level calls (fetchPropertyData etc) are just orchestrators now. They accept 'runTask'.
	// They DO NOT block. We can just call them sequentially on the main thread, and they will schedule tasks via runTask.
	// OR we can schedule them via runTask too, but that would use a semaphore slot for the orchestrator which is fast.
	// Let's just run them directly since they are non-blocking now!

	// Launch fetchers in prioritized phases

	// Helper to create a runner bound to a specific WaitGroup
	makeRunner := func(wg *sync.WaitGroup) func(func()) {
		return func(fn func()) {
			runTask(wg, fn)
		}
	}

	// Execute Phases 1-3 with a global timeout for the AI Summary compatibility
	phasesDone := make(chan struct{})
	go func() {
		defer close(phasesDone)

		// Phase 1: High Priority / Fast / Visual (Weather, Environment, Transport, Population)
		logutil.Debugf("[AGGREGATOR] Starting Phase 1: Environment & Transport")

		contextHit := false
		if pa.cache != nil { // Context check (even if bypassCache? No, respect bypass)
			if !bypassCache {
				ctxKey := cache.CacheKey{}.ContextKey(postcode)
				var cachedCtx ComprehensivePropertyData
				if err := pa.cache.Get(ctxKey, &cachedCtx); err == nil {
					logutil.Debugf("[AGGREGATOR] Context cache hit for %s", postcode)
					// Copy context fields
					mu.Lock()
					data.Weather = cachedCtx.Weather
					data.SolarPotential = cachedCtx.SolarPotential
					data.AirQuality = cachedCtx.AirQuality
					data.NoisePollution = cachedCtx.NoisePollution
					data.Population = cachedCtx.Population
					data.SquareStats = cachedCtx.SquareStats
					data.TrafficData = cachedCtx.TrafficData
					data.PublicTransport = cachedCtx.PublicTransport
					data.GreenSpaces = cachedCtx.GreenSpaces
					data.Education = cachedCtx.Education
					data.Facilities = cachedCtx.Facilities
					data.BROSoilMap = cachedCtx.BROSoilMap
					data.LandUse = cachedCtx.LandUse
					data.PDOKData = cachedCtx.PDOKData
					mu.Unlock()

					// Report progress for each cached data source so frontend updates cards
					if data.Weather != nil {
						reportProgress("KNMI Weather", "success", data.Weather)
					}
					if data.SolarPotential != nil {
						reportProgress("KNMI Solar", "success", data.SolarPotential)
					}
					if data.AirQuality != nil {
						reportProgress("Luchtmeetnet Air Quality", "success", data.AirQuality)
					}
					if data.NoisePollution != nil {
						reportProgress("Noise Pollution", "success", data.NoisePollution)
					}
					if data.Population != nil {
						reportProgress("CBS Population", "success", data.Population)
					}
					if data.SquareStats != nil {
						reportProgress("CBS Square Statistics", "success", data.SquareStats)
					}
					if len(data.TrafficData) > 0 {
						reportProgress("NDW Traffic", "success", data.TrafficData)
					}
					if data.PublicTransport != nil {
						reportProgress("openOV Public Transport", "success", data.PublicTransport)
					}
					if data.GreenSpaces != nil {
						reportProgress("Green Spaces", "success", data.GreenSpaces)
					}
					if data.Education != nil {
						reportProgress("Education Facilities", "success", data.Education)
					}
					if data.Facilities != nil {
						reportProgress("Facilities & Amenities", "success", data.Facilities)
					}
					if data.BROSoilMap != nil {
						reportProgress("BRO Soil Map", "success", data.BROSoilMap)
					}
					if data.LandUse != nil {
						reportProgress("Land Use & Zoning", "success", data.LandUse)
					}
					if data.PDOKData != nil {
						reportProgress("PDOK Platform", "success", data.PDOKData)
					}

					contextHit = true
				}
			}
		}

		if !contextHit {
			var wg1 sync.WaitGroup
			runner1 := makeRunner(&wg1)

			runner1(func() { pa.fetchEnvironmentalData(ctx, cfg, &mu, data, lat, lon, reportProgress, runner1) })
			runner1(func() { pa.fetchMobilityData(ctx, cfg, &mu, data, lat, lon, reportProgress, runner1) })
			runner1(func() {
				pa.fetchDemographicsData(ctx, cfg, &mu, data, lat, lon, neighborhoodCode, regionCode, reportProgress, runner1)
			})
			runner1(func() { pa.fetchInfrastructureData(ctx, cfg, &mu, data, lat, lon, reportProgress, runner1) })
			wg1.Wait()

			// Save to context cache (in background)
			if pa.cache != nil {
				go func() {
					ctxKey := cache.CacheKey{}.ContextKey(postcode)
					// Create copy with only context fields
					mu.Lock()
					ctxData := ComprehensivePropertyData{
						Weather:         data.Weather,
						SolarPotential:  data.SolarPotential,
						AirQuality:      data.AirQuality,
						NoisePollution:  data.NoisePollution,
						Population:      data.Population,
						SquareStats:     data.SquareStats,
						TrafficData:     data.TrafficData,
						PublicTransport: data.PublicTransport,
						GreenSpaces:     data.GreenSpaces,
						Education:       data.Education,
						Facilities:      data.Facilities,
						BROSoilMap:      data.BROSoilMap,
						LandUse:         data.LandUse,
						PDOKData:        data.PDOKData,
						AggregatedAt:    time.Now(),
					}
					mu.Unlock()
					if err := pa.cache.Set(ctxKey, ctxData, cache.PropertyDataTTL); err != nil {
						logutil.Warnf("Failed to cache context data: %v", err)
					}
				}()
			}
		}

		// Phase 2: Property Specifics & Risk
		logutil.Debugf("[AGGREGATOR] Starting Phase 2: Property & Risk")
		var wg2 sync.WaitGroup
		runner2 := makeRunner(&wg2)

		runner2(func() { pa.fetchPropertyData(ctx, cfg, &mu, data, bagID, lat, lon, reportProgress, runner2) })
		runner2(func() { pa.fetchRiskData(ctx, cfg, &mu, data, lat, lon, neighborhoodCode, reportProgress, runner2) })
		wg2.Wait()

		// Phase 3: Energy, Platforms, Supplemental
		logutil.Debugf("[AGGREGATOR] Starting Phase 3: Energy & Platforms")
		var wg3 sync.WaitGroup
		runner3 := makeRunner(&wg3)

		runner3(func() { pa.fetchEnergyData(ctx, cfg, &mu, data, bagID, reportProgress, runner3) })
		runner3(func() { pa.fetchPlatformData(ctx, cfg, &mu, data, lat, lon, reportProgress, runner3) })
		wg3.Wait()
	}()

	// Wait for data collection to finish OR 30s timeout
	select {
	case <-phasesDone:
		logutil.Debugf("[AGGREGATOR] All data phases completed in time")
	case <-time.After(30 * time.Second):
		logutil.Warnf("[AGGREGATOR] Data collection timed out (30s); proceeding to AI summary with partial data")
	}

	// AI Summary (Sequential) - Cache per postcode since it's based only on area data
	if aiSummary, err := pa.getOrGenerateAISummary(ctx, cfg, data, postcode); err == nil {
		data.AISummary = aiSummary
		if aiSummary.Generated {
			data.DataSources = append(data.DataSources, "Gemini AI")
			reportProgress("Gemini AI", "success", aiSummary)
		}
	} else {
		if data.Errors != nil {
			data.Errors["Gemini AI"] = err.Error()
		}
		reportProgress("Gemini AI", "error", nil)
	}

	// Cache the aggregated result
	if pa.cache != nil {
		cacheKey := cache.CacheKey{}.AggregatedKey(postcode, houseNumber)
		pa.cache.Set(cacheKey, data, cache.PropertyDataTTL)
	}

	return data, nil
}

// getOrGenerateAISummary retrieves AI summary from cache or generates it
// AI summaries are cached per postcode since they're based only on area-level data
func (pa *PropertyAggregator) getOrGenerateAISummary(ctx context.Context, cfg *config.Config, data *ComprehensivePropertyData, postcode string) (*models.GeminiSummary, error) {
	if pa.cache != nil {
		cacheKey := cache.CacheKey{}.AISummaryKey(postcode)
		var cachedSummary models.GeminiSummary
		if err := pa.cache.Get(cacheKey, &cachedSummary); err == nil {
			logutil.Debugf("[AGGREGATOR] AI summary cache hit for postcode %s", postcode)
			return &cachedSummary, nil
		}
	}

	// Generate summary based on area-level data only
	aiSummary, err := pa.apiClient.GenerateLocationSummary(ctx, cfg, data)
	if err != nil {
		return aiSummary, err
	}

	// Cache the summary per postcode
	if pa.cache != nil && aiSummary != nil && aiSummary.Generated {
		cacheKey := cache.CacheKey{}.AISummaryKey(postcode)
		if err := pa.cache.Set(cacheKey, aiSummary, cache.PropertyDataTTL); err != nil {
			logutil.Warnf("[AGGREGATOR] Failed to cache AI summary: %v", err)
		}
	}

	return aiSummary, nil
}

// safeAppendSource adds a source to DataSources in a thread-safe way
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

func (pa *PropertyAggregator) fetchPropertyData(ctx context.Context, cfg *config.Config, mu *sync.Mutex, data *ComprehensivePropertyData, bagID string, lat, lon float64, onProgress func(string, string, interface{}), runTask func(func())) {
	// Kadaster Object Info
	runTask(func() {
		if kadasterInfo, err := pa.apiClient.FetchKadasterObjectInfo(ctx, cfg, bagID); err == nil {
			data.KadasterInfo = kadasterInfo
			safeAppendSource(mu, data, "Kadaster Object Info")
			onProgress("Kadaster Object Info", "success", kadasterInfo)
		} else {
			logutil.Debugf("[AGGREGATOR] Kadaster fetch failed: %v", err)
			safeRecordError(mu, data, "Kadaster Object Info", err.Error())
			onProgress("Kadaster Object Info", "error", nil)
		}
	})

	// WOZ Data
	runTask(func() {
		if wozData, err := pa.apiClient.FetchAltumWOZData(ctx, cfg, bagID); err == nil {
			data.WOZData = wozData
			safeAppendSource(mu, data, "Altum WOZ")
			onProgress("Altum WOZ", "success", wozData)
		} else {
			logutil.Debugf("[AGGREGATOR] Altum WOZ fetch failed: %v", err)
			safeRecordError(mu, data, "Altum WOZ", err.Error())
			onProgress("Altum WOZ", "error", nil)
		}
	})

	// Market Valuation
	runTask(func() {
		if valuation, err := pa.apiClient.FetchPropertyValuePlus(ctx, cfg, bagID, lat, lon); err == nil {
			data.MarketValuation = valuation
			safeAppendSource(mu, data, "Matrixian Property Value+")
			onProgress("Matrixian Property Value+", "success", valuation)
		} else {
			logutil.Debugf("[AGGREGATOR] Matrixian fetch failed: %v", err)
			safeRecordError(mu, data, "Matrixian Property Value+", err.Error())
			onProgress("Matrixian Property Value+", "error", nil)
		}
	})

	// Transaction History
	runTask(func() {
		if transactions, err := pa.apiClient.FetchTransactionHistory(ctx, cfg, bagID); err == nil {
			data.TransactionHistory = transactions
			safeAppendSource(mu, data, "Altum Transactions")
			onProgress("Altum Transactions", "success", transactions)
		} else {
			logutil.Debugf("[AGGREGATOR] Altum Transactions fetch failed: %v", err)
			safeRecordError(mu, data, "Altum Transactions", err.Error())
			onProgress("Altum Transactions", "error", nil)
		}
	})

	// Monument Status
	runTask(func() {
		if monument, err := pa.apiClient.FetchMonumentDataByCoords(ctx, cfg, lat, lon); err == nil {
			data.MonumentStatus = monument
			safeAppendSource(mu, data, "Monument Status")
			onProgress("Monument Status", "success", monument)
		} else {
			logutil.Debugf("[AGGREGATOR] Monument fetch failed: %v", err)
			safeRecordError(mu, data, "Monument Status", err.Error())
			onProgress("Monument Status", "error", nil)
		}
	})
}

func (pa *PropertyAggregator) fetchEnvironmentalData(ctx context.Context, cfg *config.Config, mu *sync.Mutex, data *ComprehensivePropertyData, lat, lon float64, onProgress func(string, string, interface{}), runTask func(func())) {
	// Weather
	runTask(func() {
		weather, weatherErr := pa.apiClient.FetchKNMIWeatherData(ctx, cfg, lat, lon)
		if weatherErr == nil && weather != nil {
			data.Weather = weather
			safeAppendSource(mu, data, "KNMI Weather")
			onProgress("KNMI Weather", "success", weather)
		} else {
			safeRecordError(mu, data, "KNMI Weather", weatherErr.Error())
			onProgress("KNMI Weather", "error", nil)
		}
	})

	// Solar
	runTask(func() {
		solar, solarErr := pa.apiClient.FetchKNMISolarData(ctx, cfg, lat, lon)
		if solarErr == nil && solar != nil {
			data.SolarPotential = solar
			safeAppendSource(mu, data, "KNMI Solar")
			onProgress("KNMI Solar", "success", solar)
		} else {
			safeRecordError(mu, data, "KNMI Solar", solarErr.Error())
			onProgress("KNMI Solar", "error", nil)
		}
	})

	// Soil Data
	runTask(func() {
		if soil, err := pa.apiClient.FetchWURSoilData(ctx, cfg, lat, lon); err == nil {
			data.SoilData = soil
			safeAppendSource(mu, data, "WUR Soil Physicals")
			onProgress("WUR Soil Physicals", "success", soil)
		} else {
			safeRecordError(mu, data, "WUR Soil Physicals", err.Error())
			onProgress("WUR Soil Physicals", "error", nil)
		}
	})

	// Subsidence
	runTask(func() {
		if subsidence, err := pa.apiClient.FetchSubsidenceData(ctx, cfg, lat, lon); err == nil {
			data.Subsidence = subsidence
			safeAppendSource(mu, data, "SkyGeo Subsidence")
			onProgress("SkyGeo Subsidence", "success", subsidence)
		} else {
			safeRecordError(mu, data, "SkyGeo Subsidence", err.Error())
			onProgress("SkyGeo Subsidence", "error", nil)
		}
	})

	// Soil Quality
	runTask(func() {
		if soilQuality, err := pa.apiClient.FetchSoilQualityData(ctx, cfg, lat, lon); err == nil {
			data.SoilQuality = soilQuality
			safeAppendSource(mu, data, "Soil Quality")
			onProgress("Soil Quality", "success", soilQuality)
		} else {
			safeRecordError(mu, data, "Soil Quality", err.Error())
			onProgress("Soil Quality", "error", nil)
		}
	})

	// BRO Soil Map
	runTask(func() {
		if broSoil, err := pa.apiClient.FetchBROSoilMapData(ctx, cfg, lat, lon); err == nil {
			data.BROSoilMap = broSoil
			safeAppendSource(mu, data, "BRO Soil Map")
			onProgress("BRO Soil Map", "success", broSoil)
		} else {
			safeRecordError(mu, data, "BRO Soil Map", err.Error())
			onProgress("BRO Soil Map", "error", nil)
		}
	})

	// Air Quality
	runTask(func() {
		if airQuality, err := pa.apiClient.FetchAirQualityData(ctx, cfg, lat, lon); err == nil {
			data.AirQuality = airQuality
			safeAppendSource(mu, data, "Luchtmeetnet Air Quality")
			onProgress("Luchtmeetnet Air Quality", "success", airQuality)
		} else {
			safeRecordError(mu, data, "Luchtmeetnet Air Quality", err.Error())
			onProgress("Luchtmeetnet Air Quality", "error", nil)
		}
	})

	// Noise Pollution
	runTask(func() {
		if noise, err := pa.apiClient.FetchNoisePollutionData(ctx, cfg, lat, lon); err == nil {
			data.NoisePollution = noise
			safeAppendSource(mu, data, "Noise Pollution")
			onProgress("Noise Pollution", "success", noise)
		} else {
			safeRecordError(mu, data, "Noise Pollution", err.Error())
			onProgress("Noise Pollution", "error", nil)
		}
	})
}

func (pa *PropertyAggregator) fetchEnergyData(ctx context.Context, cfg *config.Config, mu *sync.Mutex, data *ComprehensivePropertyData, bagID string, onProgress func(string, string, interface{}), runTask func(func())) {
	// Energy & Climate
	runTask(func() {
		if energy, err := pa.apiClient.FetchEnergyClimateData(ctx, cfg, bagID); err == nil {
			data.EnergyClimate = energy
			safeAppendSource(mu, data, "Altum Energy & Climate")
			onProgress("Altum Energy & Climate", "success", energy)
		} else {
			safeRecordError(mu, data, "Altum Energy & Climate", err.Error())
			onProgress("Altum Energy & Climate", "error", nil)
		}
	})

	// Sustainability
	runTask(func() {
		if sustainability, err := pa.apiClient.FetchSustainabilityData(ctx, cfg, bagID); err == nil {
			data.Sustainability = sustainability
			safeAppendSource(mu, data, "Altum Sustainability")
			onProgress("Altum Sustainability", "success", sustainability)
		} else {
			safeRecordError(mu, data, "Altum Sustainability", err.Error())
			onProgress("Altum Sustainability", "error", nil)
		}
	})
}

func (pa *PropertyAggregator) fetchRiskData(ctx context.Context, cfg *config.Config, mu *sync.Mutex, data *ComprehensivePropertyData, lat, lon float64, neighborhoodCode string, onProgress func(string, string, interface{}), runTask func(func())) {
	// Flood Risk
	runTask(func() {
		if flood, err := pa.apiClient.FetchFloodRiskData(ctx, cfg, lat, lon); err == nil {
			data.FloodRisk = flood
			safeAppendSource(mu, data, "Flood Risk")
			onProgress("Flood Risk", "success", flood)
		} else {
			safeRecordError(mu, data, "Flood Risk", err.Error())
			onProgress("Flood Risk", "error", nil)
		}
	})

	// Water Quality
	runTask(func() {
		if water, err := pa.apiClient.FetchWaterQualityData(ctx, cfg, lat, lon); err == nil {
			data.WaterQuality = water
			safeAppendSource(mu, data, "Digital Delta Water Quality")
			onProgress("Digital Delta Water Quality", "success", water)
		} else {
			safeRecordError(mu, data, "Digital Delta Water Quality", err.Error())
			onProgress("Digital Delta Water Quality", "error", nil)
		}
	})

	// Safety
	runTask(func() {
		if neighborhoodCode != "" {
			if safety, err := pa.apiClient.FetchSafetyData(ctx, cfg, neighborhoodCode); err == nil {
				data.Safety = safety
				safeAppendSource(mu, data, "CBS Safety Experience")
				onProgress("CBS Safety Experience", "success", safety)
			} else {
				safeRecordError(mu, data, "CBS Safety Experience", err.Error())
				onProgress("CBS Safety Experience", "error", nil)
			}
		} else {
			safeRecordError(mu, data, "CBS Safety Experience", "neighborhood code not available")
			onProgress("CBS Safety Experience", "skipped", nil)
		}
	})

	// Schiphol Flights
	runTask(func() {
		if flights, err := pa.apiClient.FetchSchipholFlightData(ctx, cfg, lat, lon); err == nil {
			data.SchipholFlights = flights
			safeAppendSource(mu, data, "Schiphol Flight Noise")
			onProgress("Schiphol Flight Noise", "success", flights)
		} else {
			safeRecordError(mu, data, "Schiphol Flight Noise", err.Error())
			onProgress("Schiphol Flight Noise", "error", nil)
		}
	})
}

func (pa *PropertyAggregator) fetchMobilityData(ctx context.Context, cfg *config.Config, mu *sync.Mutex, data *ComprehensivePropertyData, lat, lon float64, onProgress func(string, string, interface{}), runTask func(func())) {
	// Traffic Data
	runTask(func() {
		if traffic, err := pa.apiClient.FetchNDWTrafficData(ctx, cfg, lat, lon, 1000); err == nil {
			data.TrafficData = traffic
			safeAppendSource(mu, data, "NDW Traffic")
			onProgress("NDW Traffic", "success", traffic)
		} else {
			safeRecordError(mu, data, "NDW Traffic", err.Error())
			onProgress("NDW Traffic", "error", nil)
		}
	})

	// Public Transport
	runTask(func() {
		if transport, err := pa.apiClient.FetchOpenOVData(ctx, cfg, lat, lon); err == nil {
			data.PublicTransport = transport
			safeAppendSource(mu, data, "openOV Public Transport")
			onProgress("openOV Public Transport", "success", transport)
		} else {
			safeRecordError(mu, data, "openOV Public Transport", err.Error())
			onProgress("openOV Public Transport", "error", nil)
		}
	})

	// Parking
	runTask(func() {
		if parking, err := pa.apiClient.FetchParkingData(ctx, cfg, lat, lon, 500); err == nil {
			data.ParkingData = parking
			safeAppendSource(mu, data, "Parking Availability")
			onProgress("Parking Availability", "success", parking)
		} else {
			safeRecordError(mu, data, "Parking Availability", err.Error())
			onProgress("Parking Availability", "error", nil)
		}
	})
}

func (pa *PropertyAggregator) fetchDemographicsData(ctx context.Context, cfg *config.Config, mu *sync.Mutex, data *ComprehensivePropertyData, lat, lon float64, neighborhoodCode, regionCode string, onProgress func(string, string, interface{}), runTask func(func())) {
	// Population
	runTask(func() {
		if population, err := pa.apiClient.FetchCBSPopulationData(ctx, cfg, lat, lon); err == nil {
			data.Population = population
			safeAppendSource(mu, data, "CBS Population")
			onProgress("CBS Population", "success", population)
		} else {
			safeRecordError(mu, data, "CBS Population", err.Error())
			onProgress("CBS Population", "error", nil)
		}
	})

	// Square Stats
	runTask(func() {
		if squareStats, err := pa.apiClient.FetchCBSSquareStats(ctx, cfg, lat, lon); err == nil {
			data.SquareStats = squareStats
			safeAppendSource(mu, data, "CBS Square Statistics")
			onProgress("CBS Square Statistics", "success", squareStats)
		} else {
			safeRecordError(mu, data, "CBS Square Statistics", err.Error())
			onProgress("CBS Square Statistics", "error", nil)
		}
	})

	// StatLine
	runTask(func() {
		if regionCode != "" {
			if statLine, err := pa.apiClient.FetchCBSStatLineData(ctx, cfg, regionCode); err == nil {
				data.StatLineData = statLine
				safeAppendSource(mu, data, "CBS StatLine")
				onProgress("CBS StatLine", "success", statLine)
			} else {
				safeRecordError(mu, data, "CBS StatLine", err.Error())
				onProgress("CBS StatLine", "error", nil)
			}
		} else {
			safeRecordError(mu, data, "CBS StatLine", "region code not available")
			onProgress("CBS StatLine", "skipped", nil)
		}
	})

	// Legacy CBS Data
	runTask(func() {
		if neighborhoodCode != "" {
			if cbsData, err := pa.apiClient.FetchCBSData(ctx, cfg, neighborhoodCode); err == nil {
				data.CBSData = cbsData
				safeAppendSource(mu, data, "CBS")
				onProgress("CBS", "success", cbsData)
			} else {
				safeRecordError(mu, data, "CBS", err.Error())
				onProgress("CBS", "error", nil)
			}
		} else {
			safeRecordError(mu, data, "CBS", "neighborhood code not available")
			onProgress("CBS", "skipped", nil)
		}
	})
}

func (pa *PropertyAggregator) fetchInfrastructureData(ctx context.Context, cfg *config.Config, mu *sync.Mutex, data *ComprehensivePropertyData, lat, lon float64, onProgress func(string, string, interface{}), runTask func(func())) {
	// Green Spaces
	runTask(func() {
		if greenSpaces, err := pa.apiClient.FetchGreenSpacesData(ctx, cfg, lat, lon, 1000); err == nil {
			data.GreenSpaces = greenSpaces
			safeAppendSource(mu, data, "Green Spaces")
			onProgress("Green Spaces", "success", greenSpaces)
		} else {
			safeRecordError(mu, data, "Green Spaces", err.Error())
			onProgress("Green Spaces", "error", nil)
		}
	})

	// Education
	runTask(func() {
		if education, err := pa.apiClient.FetchEducationData(ctx, cfg, lat, lon); err == nil {
			data.Education = education
			safeAppendSource(mu, data, "Education Facilities")
			onProgress("Education Facilities", "success", education)
		} else {
			safeRecordError(mu, data, "Education Facilities", err.Error())
			onProgress("Education Facilities", "error", nil)
		}
	})

	// Building Permits
	runTask(func() {
		if permits, err := pa.apiClient.FetchBuildingPermitsData(ctx, cfg, lat, lon, 1000); err == nil {
			data.BuildingPermits = permits
			safeAppendSource(mu, data, "Building Permits")
			onProgress("Building Permits", "success", permits)
		} else {
			safeRecordError(mu, data, "Building Permits", err.Error())
			onProgress("Building Permits", "error", nil)
		}
	})

	// Facilities
	runTask(func() {
		if facilities, err := pa.apiClient.FetchFacilitiesData(ctx, cfg, lat, lon); err == nil {
			data.Facilities = facilities
			safeAppendSource(mu, data, "Facilities & Amenities")
			onProgress("Facilities & Amenities", "success", facilities)
		} else {
			safeRecordError(mu, data, "Facilities & Amenities", err.Error())
			onProgress("Facilities & Amenities", "error", nil)
		}
	})

	// Elevation (AHN)
	runTask(func() {
		if elevation, err := pa.apiClient.FetchAHNHeightData(ctx, pa.config, lat, lon); err == nil {
			data.Elevation = elevation
			safeAppendSource(mu, data, "AHN Height Model")
			onProgress("AHN Height Model", "success", elevation)
		} else {
			safeRecordError(mu, data, "AHN Height Model", err.Error())
			onProgress("AHN Height Model", "error", nil)
		}
	})
}

func (pa *PropertyAggregator) fetchPlatformData(ctx context.Context, cfg *config.Config, mu *sync.Mutex, data *ComprehensivePropertyData, lat, lon float64, onProgress func(string, string, interface{}), runTask func(func())) {
	// PDOK
	runTask(func() {
		if pdok, err := pa.apiClient.FetchPDOKPlatformData(ctx, cfg, lat, lon); err == nil {
			data.PDOKData = pdok
			safeAppendSource(mu, data, "PDOK Platform")
			onProgress("PDOK Platform", "success", pdok)
		} else {
			safeRecordError(mu, data, "PDOK Platform", err.Error())
			onProgress("PDOK Platform", "error", nil)
		}
	})

	// Stratopo
	runTask(func() {
		if stratopo, err := pa.apiClient.FetchStratopoEnvironmentData(ctx, cfg, lat, lon); err == nil {
			data.StratopoEnvironment = stratopo
			safeAppendSource(mu, data, "Stratopo Environment")
			onProgress("Stratopo Environment", "success", stratopo)
		} else {
			safeRecordError(mu, data, "Stratopo Environment", err.Error())
			onProgress("Stratopo Environment", "error", nil)
		}
	})

	// Land Use
	runTask(func() {
		if landUse, err := pa.apiClient.FetchLandUseData(ctx, cfg, lat, lon); err == nil {
			data.LandUse = landUse
			safeAppendSource(mu, data, "Land Use & Zoning")
			onProgress("Land Use & Zoning", "success", landUse)
		} else {
			safeRecordError(mu, data, "Land Use & Zoning", err.Error())
			onProgress("Land Use & Zoning", "error", nil)
		}
	})
}
