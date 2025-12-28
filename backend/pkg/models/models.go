package models

import (
	"encoding/json"
	"time"
)

// BAGData represents data from the BAG API
// Address: full address string
// Coordinates: [latitude, longitude]
// GeoJSON: raw GeoJSON as a string
// ...rockstar style, keep it simple

type BAGData struct {
	Address            string     `json:"address"`
	Coordinates        [2]float64 `json:"coordinates"`
	GeoJSON            string     `json:"geojson"`
	ID                 string     `json:"id,omitempty"`
	NummeraanduidingID string     `json:"nummeraanduiding_id,omitempty"`
	VerblijfsobjectID  string     `json:"verblijfsobject_id,omitempty"`
	PandID             string     `json:"pand_id,omitempty"`
	Municipality       string     `json:"municipality,omitempty"`
	MunicipalityCode   string     `json:"municipality_code,omitempty"`
	Province           string     `json:"province,omitempty"`
	ProvinceCode       string     `json:"province_code,omitempty"`
}

// PDOKData represents data from PDOK
// ZoningInfo: zoning information string
// Restrictions: list of restrictions

type PDOKData struct {
	ZoningInfo   string   `json:"zoning_info"`
	Restrictions []string `json:"restrictions"`
}

// PropertyScore holds calculated scores
// Viability, Investment, ESG: all float64

type PropertyScore struct {
	Viability  float64 `json:"viability"`
	Investment float64 `json:"investment"`
	ESG        float64 `json:"esg"`
}

// AggregatedData wraps all the above
type AggregatedData struct {
	BAGData          BAGData       `json:"bag_data"`
	PDOKData         PDOKData      `json:"pdok_data"`
	PropertyScore    PropertyScore `json:"property_score"`
	BAGJSON          *string       `json:"bag_data_json"`
	PDOKJSON         *string       `json:"pdok_data_json"`
	ScoresJSON       *string       `json:"scores_json"`
	Address          string        `json:"address"`
	Municipality     string        `json:"municipality"`
	RawJSON          string        `json:"raw_json"`
	ID               string        `json:"id"`
	CBSJSON          *string       `json:"cbs_data_json"`
	NeighborhoodCode string        `json:"neighborhood_code"`
	MonumentJSON     *string       `json:"monument_data_json"`
	AsbestosJSON     *string       `json:"asbestos_data_json"`
	EnergyJSON       *string       `json:"energy_data_json"`
}

// BagResponse mirrors the PDOK Locatieserver free endpoint JSON payload.
type BagResponse struct {
	Response struct {
		Docs []BagDocument `json:"docs"`
	} `json:"response"`
}

// BagDocument contains fields from BAG API
type BagDocument struct {
	ID                 string  `json:"id"`
	NummeraanduidingID string  `json:"nummeraanduiding_id"`
	VerblijfsobjectID  string  `json:"verblijfsobject_id"`
	PandID             string  `json:"pand_id"`
	Weergavenaam       string  `json:"weergavenaam"`
	Straatnaam         string  `json:"straatnaam"`
	Huisnummer         float64 `json:"huisnummer"`
	HuisNLT            string  `json:"huis_nlt"`
	Huisletter         string  `json:"huisletter"`
	Huisnummertoevoeg  string  `json:"huisnummertoevoeging"`
	Postcode           string  `json:"postcode"`
	WoonplaatsNaam     string  `json:"woonplaatsnaam"`
	Gemeentenaam       string  `json:"gemeentenaam"`
	Gemeentecode       string  `json:"gemeentecode"`
	Provincienaam      string  `json:"provincienaam"`
	Provinciecode      string  `json:"provinciecode"`
	CentroidLL         string  `json:"centroide_ll"`
	GeometriePolygoon  string  `json:"geometrie_polygoon"`
}

// KNMIWeatherData represents weather data
type KNMIWeatherData struct {
	StationName        string           `json:"stationName"`
	Temperature        float64          `json:"temperature"`      // Celsius
	WindSpeed          float64          `json:"windSpeed"`        // m/s
	WindDirection      int              `json:"windDirection"`    // degrees
	Rainfall           float64          `json:"rainfall"`         // mm last 24h
	Precipitation      float64          `json:"precipitation"`    // mm current
	RainfallForecast   []float64        `json:"rainfallForecast"` // mm next hours
	Humidity           float64          `json:"humidity"`         // percentage
	Pressure           float64          `json:"pressure"`         // hPa
	SunshineConfig     string           `json:"sunshine"`         // percentage or description
	Code               string           `json:"code"`             // weather code description
	Timestamp          string           `json:"timestamp"`
	LastUpdated        time.Time        `json:"lastUpdated"`
	HistoricalRainfall []HistoricalData `json:"historicalRainfall"`
}

// KNMISolarData represents solar potential
type KNMISolarData struct {
	SolarRadiation      float64          `json:"solarRadiation"` // W/m2 (GlobalRadiation)
	SunshineHours       float64          `json:"sunshineHours"`  // hours (SunshineDuration)
	UVIndex             float64          `json:"uvIndex"`
	Date                string           `json:"date"`
	Historical          []HistoricalData `json:"historical"`
	PotentialGeneration float64          `json:"potentialGeneration"` // kWh/year estimate
}

// HistoricalData represents historical climate data point
type HistoricalData struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// GreenSpacesData represents parks and green areas
type GreenSpacesData struct {
	TotalGreenArea  float64      `json:"totalGreenArea"`  // m² within radius
	GreenPercentage float64      `json:"greenPercentage"` // percentage of area
	NearestPark     string       `json:"nearestPark"`
	ParkDistance    float64      `json:"parkDistance"`    // meters
	TreeCanopyCover float64      `json:"treeCanopyCover"` // percentage
	GreenSpaces     []GreenSpace `json:"greenSpaces"`
}

// GreenSpace represents a park or green area
type GreenSpace struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`       // Park, Forest, Garden, etc.
	Area       float64  `json:"area"`       // m²
	Distance   float64  `json:"distance"`   // meters
	Facilities []string `json:"facilities"` // Playground, Sports, etc.
	Lat        float64  `json:"lat"`
	Lon        float64  `json:"lon"`
}

// BgtGreenResponse represents PDOK BGT begroeidterreindeel API response
type BgtGreenResponse struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Properties struct {
			FysiekVoorkomen string `json:"fysiekVoorkomen"` // e.g., "groenvoorziening", "bos"
			Naam            string `json:"naam"`
			OpenbareRuimte  string `json:"openbareRuimte"`
		} `json:"properties"`
		Geometry struct {
			Type        string          `json:"type"`
			Coordinates json.RawMessage `json:"coordinates"` // using json.RawMessage as definition
		} `json:"geometry"`
	} `json:"features"`
	NumberReturned int `json:"numberReturned"`
}

// Natura2000Response represents PDOK Natura2000 API response
type Natura2000Response struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Properties struct {
			Naam        string  `json:"naam"`
			Oppervlakte float64 `json:"oppervlakte"`
			Status      string  `json:"status"`
		} `json:"properties"`
	} `json:"features"`
	NumberReturned int `json:"numberReturned"`
}

// EducationData represents schools and education facilities
type EducationData struct {
	NearestPrimarySchool   *School  `json:"nearestPrimarySchool"`
	NearestSecondarySchool *School  `json:"nearestSecondarySchool"`
	AllSchools             []School `json:"allSchools"`
	AverageQuality         float64  `json:"averageQuality"` // 0-10 rating
}

// School represents an educational facility
type School struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`         // Primary, Secondary, Special
	Distance     float64 `json:"distance"`     // meters
	QualityScore float64 `json:"qualityScore"` // 0-10
	Students     int     `json:"students"`
	Address      string  `json:"address"`
	Denomination string  `json:"denomination"` // Public, Catholic, etc.
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
}

// OverpassResponse represents OSM Overpass API response
type OverpassResponse struct {
	Elements []struct {
		Type string  `json:"type"`
		ID   int64   `json:"id"`
		Lat  float64 `json:"lat"`
		Lon  float64 `json:"lon"`
		Tags struct {
			Name        string `json:"name"`
			Amenity     string `json:"amenity"`
			ISCEDLevel  string `json:"isced:level"`
			Operator    string `json:"operator"`
			Religion    string `json:"religion"`
			AddrStreet  string `json:"addr:street"`
			AddrHouseNo string `json:"addr:housenumber"`
			AddrCity    string `json:"addr:city"`
		} `json:"tags"`
	} `json:"elements"`
}

// BuildingPermitsData represents recent construction activity
type BuildingPermitsData struct {
	TotalPermits    int              `json:"totalPermits"`
	NewConstruction int              `json:"newConstruction"`
	Renovations     int              `json:"renovations"`
	Permits         []BuildingPermit `json:"permits"`
	GrowthTrend     string           `json:"growthTrend"` // Increasing, Stable, Decreasing
}

// BuildingPermit represents a single permit
type BuildingPermit struct {
	PermitID     string  `json:"permitId"`
	Type         string  `json:"type"` // New, Renovation, Extension, Demolition
	Address      string  `json:"address"`
	Distance     float64 `json:"distance"` // meters
	IssueDate    string  `json:"issueDate"`
	ProjectValue float64 `json:"projectValue"` // EUR
	Status       string  `json:"status"`       // Approved, In Progress, Completed
}

// FacilitiesData represents nearby amenities
type FacilitiesData struct {
	TopFacilities  []Facility     `json:"topFacilities"`
	AmenitiesScore float64        `json:"amenitiesScore"` // 0-100
	CategoryCounts map[string]int `json:"categoryCounts"`
}

// Facility represents a single amenity
type Facility struct {
	Name      string  `json:"name"`
	Category  string  `json:"category"`  // Retail, Healthcare, Leisure, etc.
	Type      string  `json:"type"`      // Supermarket, Hospital, Gym, etc.
	Distance  float64 `json:"distance"`  // meters
	WalkTime  int     `json:"walkTime"`  // minutes
	DriveTime int     `json:"driveTime"` // minutes
	Rating    float64 `json:"rating"`    // 0-5 stars
	Address   string  `json:"address"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
}

// OverpassFacilitiesResponse for OSM amenities query
type OverpassFacilitiesResponse struct {
	Elements []struct {
		Type string  `json:"type"`
		ID   int64   `json:"id"`
		Lat  float64 `json:"lat"`
		Lon  float64 `json:"lon"`
		Tags struct {
			Name       string `json:"name"`
			Amenity    string `json:"amenity"`
			Shop       string `json:"shop"`
			Leisure    string `json:"leisure"`
			Healthcare string `json:"healthcare"`
		} `json:"tags"`
	} `json:"elements"`
}

// AHNHeightData represents elevation and terrain data
type AHNHeightData struct {
	Elevation     float64   `json:"elevation"`     // meters above NAP
	TerrainSlope  float64   `json:"terrainSlope"`  // degrees
	FloodRisk     string    `json:"floodRisk"`     // Low, Medium, High based on elevation
	ViewPotential string    `json:"viewPotential"` // Poor, Fair, Good, Excellent
	Surrounding   []float64 `json:"surrounding"`   // Elevations of nearby points
}

// OpenElevationResponse represents Open-Elevation API response
type OpenElevationResponse struct {
	Results []struct {
		Elevation float64 `json:"elevation"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

// WeerliveWeatherData represents weather data from Weerlive API
type WeerliveWeatherData struct {
	Temperature   float64            `json:"temp"`
	WeatherDesc   string             `json:"samenv"`
	WindSpeed     float64            `json:"windsnelheid"`
	WindDirection string             `json:"windrichting"`
	Pressure      float64            `json:"luchtdruk"`
	Humidity      float64            `json:"lv"`
	Visibility    float64            `json:"zicht"`
	Forecast      []WeerliveForecast `json:"verwachting"`
}

// WeerliveForecast represents daily forecast
type WeerliveForecast struct {
	Day           string  `json:"dag"`
	MinTemp       float64 `json:"mintemp"`
	MaxTemp       float64 `json:"maxtemp"`
	Precipitation float64 `json:"neerslag"`
	WindForce     float64 `json:"windkracht"`
}

// NDWTrafficData represents real-time traffic data
type NDWTrafficData struct {
	LocationID      string  `json:"locationId"`
	Intensity       int     `json:"intensity"`       // vehicles/hour
	AverageSpeed    float64 `json:"averageSpeed"`    // km/h
	CongestionLevel string  `json:"congestionLevel"` // Free, Light, Moderate, Heavy, Jammed
	LastUpdated     string  `json:"lastUpdated"`
	Coordinates     struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coordinates"`
}

// OpenOVTransportData represents public transport accessibility
type OpenOVTransportData struct {
	NearestStops []PublicTransportStop `json:"nearestStops"`
	Connections  []Connection          `json:"connections"`
}

// PublicTransportStop represents a PT stop
type PublicTransportStop struct {
	StopID      string  `json:"stopId"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`     // Bus, Tram, Metro, Train
	Distance    float64 `json:"distance"` // meters
	Coordinates struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coordinates"`
	Lines []string `json:"lines"`
}

// Connection represents a PT connection
type Connection struct {
	Line      string `json:"line"`
	Direction string `json:"direction"`
	Departure string `json:"departure"`
	Delay     int    `json:"delay"` // minutes
	Platform  string `json:"platform"`
}

// OverpassTransportResponse for OSM public transport query
type OverpassTransportResponse struct {
	Elements []struct {
		Type string  `json:"type"`
		ID   int64   `json:"id"`
		Lat  float64 `json:"lat"`
		Lon  float64 `json:"lon"`
		Tags struct {
			Name            string `json:"name"`
			Highway         string `json:"highway"`          // bus_stop
			Railway         string `json:"railway"`          // station, tram_stop, halt
			PublicTransport string `json:"public_transport"` // stop_position, platform
			Network         string `json:"network"`
			Operator        string `json:"operator"`
			Ref             string `json:"ref"`
		} `json:"tags"`
	} `json:"elements"`
}

// ParkingData represents parking availability data
type ParkingData struct {
	TotalSpaces     int           `json:"totalSpaces"`
	AvailableSpaces int           `json:"availableSpaces"`
	OccupancyRate   float64       `json:"occupancyRate"` // percentage
	ParkingZones    []ParkingZone `json:"parkingZones"`
	LastUpdated     string        `json:"lastUpdated"`
}

// ParkingZone represents a parking area
type ParkingZone struct {
	ZoneID      string  `json:"zoneId"`
	Name        string  `json:"name"`
	Type        string  `json:"type"` // Street, Garage, Private
	Capacity    int     `json:"capacity"`
	Available   int     `json:"available"`
	Distance    float64 `json:"distance"`   // meters
	HourlyRate  float64 `json:"hourlyRate"` // EUR
	Coordinates struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coordinates"`
}

// FloodRiskData represents flood risk assessment
type FloodRiskData struct {
	RiskLevel        string  `json:"riskLevel"`        // Low, Medium, High, Very High
	FloodProbability float64 `json:"floodProbability"` // percentage per year
	WaterDepth       float64 `json:"waterDepth"`       // meters in worst-case scenario
	NearestDike      float64 `json:"nearestDike"`      // meters
	DikeQuality      string  `json:"dikeQuality"`      // Excellent, Good, Fair, Poor
	FloodZone        string  `json:"floodZone"`        // Zone classification
}

// FloodRiskResponse represents PDOK flood risk API response (INSPIRE format)
type FloodRiskResponse struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Properties struct {
			// INSPIRE format fields
			QualitativeValue string `json:"qualitative_value"` // e.g., "Area of Potential Significant Flood Risk"
			Description      string `json:"description"`       // e.g., "Rijn type B - beschermd langs hoofdwatersysteem"
			LocalID          string `json:"local_id"`
		} `json:"properties"`
	} `json:"features"`
	NumberReturned int `json:"numberReturned"`
}

// WaterQualityData represents water quality and levels
type WaterQualityData struct {
	WaterLevel   float64            `json:"waterLevel"`   // meters above NAP
	WaterQuality string             `json:"waterQuality"` // Excellent, Good, Fair, Poor
	Parameters   map[string]float64 `json:"parameters"`   // pH, dissolved oxygen, etc.
	NearestWater string             `json:"nearestWater"` // Name of nearest water body
	Distance     float64            `json:"distance"`     // meters
	LastMeasured string             `json:"lastMeasured"`
}

// SafetyData represents safety perception and crime statistics
type SafetyData struct {
	SafetyScore        float64        `json:"safetyScore"`        // 0-100
	SafetyPerception   string         `json:"safetyPerception"`   // Very Safe, Safe, Moderate, Unsafe
	CrimeRate          float64        `json:"crimeRate"`          // per 1000 residents
	CrimeTypes         map[string]int `json:"crimeTypes"`         // Burglary, theft, etc.
	PoliceResponse     float64        `json:"policeResponse"`     // minutes average
	YearOverYearChange float64        `json:"yearOverYearChange"` // percentage change
}

// SchipholFlightData represents aviation noise data
type SchipholFlightData struct {
	DailyFlights int          `json:"dailyFlights"`
	NoiseLevel   float64      `json:"noiseLevel"` // dB(A) average
	PeakHours    []string     `json:"peakHours"`
	FlightPaths  []FlightPath `json:"flightPaths"`
	NightFlights int          `json:"nightFlights"` // 23:00-07:00
	NoiseContour string       `json:"noiseContour"` // Ke zone (35, 40, 45, etc.)
}

// FlightPath represents a flight route
type FlightPath struct {
	RouteID       string  `json:"routeId"`
	Altitude      float64 `json:"altitude"` // meters
	Distance      float64 `json:"distance"` // meters from property
	FlightsPerDay int     `json:"flightsPerDay"`
}

// WURSoilData represents soil physical properties
type WURSoilData struct {
	SoilType      string  `json:"soilType"`
	Composition   string  `json:"composition"`
	Permeability  float64 `json:"permeability"`
	OrganicMatter float64 `json:"organicMatter"`
	PH            float64 `json:"ph"`
	Suitability   string  `json:"suitability"`
}

// SubsidenceData represents land subsidence data
type SubsidenceData struct {
	SubsidenceRate  float64 `json:"subsidenceRate"`  // mm/year
	TotalSubsidence float64 `json:"totalSubsidence"` // mm since baseline
	StabilityRating string  `json:"stabilityRating"` // Low, Medium, High risk
	MeasurementDate string  `json:"measurementDate"`
	GroundMovement  float64 `json:"groundMovement"`
}

// SoilQualityData represents soil contamination and quality
type SoilQualityData struct {
	ContaminationLevel string   `json:"contaminationLevel"` // Clean, Light, Moderate, Severe
	Contaminants       []string `json:"contaminants"`
	QualityZone        string   `json:"qualityZone"`
	RestrictedUse      bool     `json:"restrictedUse"`
	LastTested         string   `json:"lastTested"`
}

// BROSoilMapData represents soil types and foundation quality
type BROSoilMapData struct {
	SoilType          string  `json:"soilType"`
	PeatComposition   float64 `json:"peatComposition"` // percentage
	Profile           string  `json:"profile"`
	FoundationQuality string  `json:"foundationQuality"` // Excellent, Good, Fair, Poor
	GroundwaterDepth  float64 `json:"groundwaterDepth"`  // meters
}

// PDOKPlatformData represents comprehensive geodata from PDOK
type PDOKPlatformData struct {
	CadastralData  *CadastralInfo  `json:"cadastralData"`
	AddressData    *AddressInfo    `json:"addressData"`
	TopographyData *TopographyInfo `json:"topographyData"`
	BoundariesData *BoundariesInfo `json:"boundariesData"`
}

// CadastralInfo represents cadastral parcel information
type CadastralInfo struct {
	ParcelID     string  `json:"parcelId"`
	Municipality string  `json:"municipality"`
	Section      string  `json:"section"`
	ParcelNumber string  `json:"parcelNumber"`
	Area         float64 `json:"area"` // m²
	LandUse      string  `json:"landUse"`
}

// AddressInfo represents comprehensive address data
type AddressInfo struct {
	BAGID        string  `json:"bagId"`
	FullAddress  string  `json:"fullAddress"`
	Municipality string  `json:"municipality"`
	Province     string  `json:"province"`
	PostalCode   string  `json:"postalCode"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

// TopographyInfo represents topographic data
type TopographyInfo struct {
	LandType        string   `json:"landType"`
	TerrainFeatures []string `json:"terrainFeatures"`
	WaterBodies     []string `json:"waterBodies"`
}

// BoundariesInfo represents administrative boundaries
type BoundariesInfo struct {
	Municipality     string `json:"municipality"`
	Province         string `json:"province"`
	Neighborhood     string `json:"neighborhood"`
	NeighborhoodCode string `json:"neighborhoodCode"`
	District         string `json:"district"`
}

// StratopoEnvironmentData represents comprehensive environmental assessment
type StratopoEnvironmentData struct {
	EnvironmentScore   float64                `json:"environmentScore"` // 0-100
	TotalVariables     int                    `json:"totalVariables"`
	PollutionIndex     float64                `json:"pollutionIndex"`
	UrbanizationLevel  string                 `json:"urbanizationLevel"` // Rural, Suburban, Urban, Metropolitan
	EnvironmentFactors map[string]interface{} `json:"environmentFactors"`
	ESGRating          string                 `json:"esgRating"` // A+ to E
	Recommendations    []string               `json:"recommendations"`
}

// LandUseData represents land use and zoning information
type LandUseData struct {
	PrimaryUse      string            `json:"primaryUse"` // Residential, Commercial, Industrial, etc.
	ZoningCode      string            `json:"zoningCode"`
	ZoningDetails   string            `json:"zoningDetails"`
	Restrictions    []string          `json:"restrictions"`
	AllowedUses     []string          `json:"allowedUses"`
	BuildingRights  *BuildingRights   `json:"buildingRights"`
	ProtectedStatus string            `json:"protectedStatus"` // None, Monument, Conservation Area
	FuturePlans     []DevelopmentPlan `json:"futurePlans"`
}

// BuildingRights represents development rights
type BuildingRights struct {
	MaxHeight      float64 `json:"maxHeight"`      // meters
	MaxBuildArea   float64 `json:"maxBuildArea"`   // m²
	FloorAreaRatio float64 `json:"floorAreaRatio"` // FSI
	GroundCoverage float64 `json:"groundCoverage"` // percentage
	CanSubdivide   bool    `json:"canSubdivide"`
	CanExpand      bool    `json:"canExpand"`
}

// DevelopmentPlan represents future development
type DevelopmentPlan struct {
	PlanName     string `json:"planName"`
	Type         string `json:"type"`
	Status       string `json:"status"` // Proposed, Approved, In Progress
	ExpectedDate string `json:"expectedDate"`
	Impact       string `json:"impact"` // Positive, Neutral, Negative
}

// MonumentData represents heritage status
type MonumentData struct {
	IsMonument bool   `json:"isMonument"`
	Type       string `json:"type"`
	Date       string `json:"date"`
	Name       string `json:"name,omitempty"`
	Number     string `json:"number,omitempty"`
}

// MonumentResponse represents the PDOK RCE INSPIRE monuments API response
type MonumentResponse struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Properties struct {
			// INSPIRE format fields
			Text                string `json:"text"`                // Monument name
			LegalFoundationDate string `json:"legalfoundationdate"` // Date of designation
			CICitation          string `json:"ci_citation"`         // Link to monumentenregister
		} `json:"properties"`
	} `json:"features"`
	NumberReturned int `json:"numberReturned"`
}

// MatrixianPropertyValue represents comprehensive market valuation data
type MatrixianPropertyValue struct {
	MarketValue          float64                `json:"marketValue"`
	ValuationDate        string                 `json:"valuationDate"`
	Confidence           float64                `json:"confidence"`
	ComparableProperties []ComparableProperty   `json:"comparableProperties"`
	Features             map[string]interface{} `json:"features"`
	PricePerSqm          float64                `json:"pricePerSqm"`
}

// ComparableProperty represents similar properties in the area
type ComparableProperty struct {
	Address      string  `json:"address"`
	Distance     float64 `json:"distance"`
	SalePrice    float64 `json:"salePrice"`
	SaleDate     string  `json:"saleDate"`
	SurfaceArea  float64 `json:"surfaceArea"`
	PropertyType string  `json:"propertyType"`
}

// AltumWOZData represents WOZ tax value data from Altum AI
type AltumWOZData struct {
	WOZValue     float64 `json:"wozValue"`
	ValueYear    int     `json:"valueYear"`
	BuildingType string  `json:"buildingType"`
	BuildYear    int     `json:"buildYear"`
	SurfaceArea  float64 `json:"surfaceArea"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

// TransactionData represents historical property transaction
type TransactionData struct {
	TransactionID string  `json:"transactionId"`
	Date          string  `json:"date"`
	PurchasePrice float64 `json:"purchasePrice"`
	PropertyType  string  `json:"propertyType"`
	SurfaceArea   float64 `json:"surfaceArea"`
	BAGObjectID   string  `json:"bagObjectId"`
}

// TransactionHistory contains list of transactions for a property
type TransactionHistory struct {
	Transactions []TransactionData `json:"transactions"`
	TotalCount   int               `json:"totalCount"`
}

// CBSData represents aggregated CBS data
type CBSData struct {
	AvgIncome         float64
	PopulationDensity float64
	AvgWOZValue       float64
}

// NeighborhoodRegionCodes contains CBS region codes for a location
type NeighborhoodRegionCodes struct {
	NeighborhoodCode string `json:"neighborhoodCode"` // Buurt code (BU...)
	DistrictCode     string `json:"districtCode"`     // Wijk code (WK...)
	MunicipalityCode string `json:"municipalityCode"` // Gemeente code (GM...)
	NeighborhoodName string `json:"neighborhoodName"`
	DistrictName     string `json:"districtName"`
	MunicipalityName string `json:"municipalityName"`
}

// CBSPopulationData represents neighbourhood-based population data from CBS buurten
type CBSPopulationData struct {
	TotalPopulation   int                    `json:"totalPopulation"`
	AgeDistribution   map[string]int         `json:"ageDistribution"`
	Households        int                    `json:"households"`
	AverageHHSize     float64                `json:"averageHouseholdSize"`
	Demographics      PopulationDemographics `json:"demographics"`
	NeighbourhoodName string                 `json:"neighbourhoodName"` // Buurtnaam
	MunicipalityName  string                 `json:"municipalityName"`  // Gemeentenaam
	DensityPerKm2     int                    `json:"densityPerKm2"`     // Population density per km²
	NeighbourhoodCode string                 `json:"neighbourhoodCode"` // Buurtcode
}

// PopulationDemographics represents demographic breakdown
type PopulationDemographics struct {
	Age0to14  int `json:"age0to14"`
	Age15to24 int `json:"age15to24"`
	Age25to44 int `json:"age25to44"`
	Age45to64 int `json:"age45to64"`
	Age65Plus int `json:"age65plus"`
}

// CBSBuurtenResponse represents the PDOK CBS wijken-en-buurten OGC API response
type CBSBuurtenResponse struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Properties struct {
			Buurtcode                    string   `json:"buurtcode"`
			Buurtnaam                    string   `json:"buurtnaam"`
			Wijkcode                     string   `json:"wijkcode"`
			Gemeentecode                 string   `json:"gemeentecode"`
			Gemeentenaam                 string   `json:"gemeentenaam"`
			AantalInwoners               *int     `json:"aantal_inwoners"`
			AantalHuishoudens            *int     `json:"aantal_huishoudens"`
			GemiddeldeHuishoudensgrootte *float64 `json:"gemiddelde_huishoudsgrootte"`
			Bevolkingsdichtheid          *int     `json:"bevolkingsdichtheid_inwoners_per_km2"`
			// Age distribution fields (percentage)
			Perc0Tot15Jaar  *int `json:"percentage_personen_0_tot_15_jaar"`
			Perc15Tot25Jaar *int `json:"percentage_personen_15_tot_25_jaar"`
			Perc25Tot45Jaar *int `json:"percentage_personen_25_tot_45_jaar"`
			Perc45Tot65Jaar *int `json:"percentage_personen_45_tot_65_jaar"`
			Perc65Plus      *int `json:"percentage_personen_65_jaar_en_ouder"`
		} `json:"properties"`
	} `json:"features"`
	NumberReturned int `json:"numberReturned"`
}

// CBSStatLineData represents comprehensive socioeconomic data
type CBSStatLineData struct {
	RegionCode     string  `json:"regionCode"`
	RegionName     string  `json:"regionName"`
	Population     int     `json:"population"`
	AverageIncome  float64 `json:"averageIncome"`  // EUR per household
	EmploymentRate float64 `json:"employmentRate"` // percentage
	EducationLevel string  `json:"educationLevel"` // Low, Medium, High
	HousingStock   int     `json:"housingStock"`
	AverageWOZ     float64 `json:"averageWOZ"`
	Year           int     `json:"year"`
}

// CBSSquareStatsData represents hyperlocal neighbourhood statistics
type CBSSquareStatsData struct {
	GridID            string  `json:"gridId"`
	Population        int     `json:"population"`
	Households        int     `json:"households"`
	AverageWOZ        float64 `json:"averageWOZ"`
	AverageIncome     float64 `json:"averageIncome"`
	HousingDensity    int     `json:"housingDensity"`    // addresses per km²
	NeighbourhoodName string  `json:"neighbourhoodName"` // Buurtnaam
	MunicipalityName  string  `json:"municipalityName"`  // Gemeentenaam
}

// CBSSquareResponse for parsing CBS grid statistics
type CBSSquareResponse struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Properties struct {
			Buurtcode                  string `json:"buurtcode"`
			Buurtnaam                  string `json:"buurtnaam"`
			Gemeentenaam               string `json:"gemeentenaam"`
			AantalInwoners             *int   `json:"aantal_inwoners"`
			AantalHuishoudens          *int   `json:"aantal_huishoudens"`
			GemiddeldeWozWaardeWoning  *int   `json:"gemiddelde_woningwaarde"`
			Omgevingsadressendichtheid *int   `json:"omgevingsadressendichtheid"`
			// Income data
			GemHuishoudinkomen *int `json:"gemiddeld_gestandaardiseerd_inkomen_van_huishoudens"`
		} `json:"properties"`
	} `json:"features"`
}

// EnergyClimateData represents energy labels and climate risk
type EnergyClimateData struct {
	EnergyLabel      string  `json:"energyLabel"`      // A++++ to G
	ClimateRisk      string  `json:"climateRisk"`      // Low, Medium, High
	EfficiencyScore  float64 `json:"efficiencyScore"`  // 0-100
	AnnualEnergyCost float64 `json:"annualEnergyCost"` // EUR
	CO2Emissions     float64 `json:"co2Emissions"`     // kg/year
	HeatLoss         float64 `json:"heatLoss"`         // W/m²K
}

// SustainabilityData represents sustainability measures and CO2 savings potential
type SustainabilityData struct {
	CurrentRating       string                  `json:"currentRating"`
	PotentialRating     string                  `json:"potentialRating"`
	RecommendedMeasures []SustainabilityMeasure `json:"recommendedMeasures"`
	TotalCO2Savings     float64                 `json:"totalCO2Savings"`  // kg/year
	TotalCostSavings    float64                 `json:"totalCostSavings"` // EUR/year
	InvestmentCost      float64                 `json:"investmentCost"`   // EUR
	PaybackPeriod       float64                 `json:"paybackPeriod"`    // years
}

// SustainabilityMeasure represents a single energy improvement measure
type SustainabilityMeasure struct {
	Type         string  `json:"type"`
	Description  string  `json:"description"`
	CO2Savings   float64 `json:"co2Savings"`  // kg/year
	CostSavings  float64 `json:"costSavings"` // EUR/year
	Investment   float64 `json:"investment"`  // EUR
	PaybackYears float64 `json:"paybackYears"`
	Priority     int     `json:"priority"` // 1 = high, 3 = low
}

// AirQualityData represents comprehensive air quality measurements
type AirQualityData struct {
	StationID    string           `json:"stationId"`
	StationName  string           `json:"stationName"`
	Measurements []AirMeasurement `json:"measurements"`
	AQI          int              `json:"aqi"`      // Air Quality Index 0-500
	Category     string           `json:"category"` // Good, Moderate, Unhealthy, etc.
	LastUpdated  string           `json:"lastUpdated"`
}

// AirMeasurement represents a single pollutant measurement
type AirMeasurement struct {
	Parameter string  `json:"parameter"` // NO2, PM10, PM2.5, O3, etc.
	Value     float64 `json:"value"`     // µg/m³
	Unit      string  `json:"unit"`
}

// NoisePollutionData represents noise levels from various sources
type NoisePollutionData struct {
	TotalNoise    float64       `json:"totalNoise"`    // dB(A)
	RoadNoise     float64       `json:"roadNoise"`     // dB(A)
	RailNoise     float64       `json:"railNoise"`     // dB(A)
	IndustryNoise float64       `json:"industryNoise"` // dB(A)
	AircraftNoise float64       `json:"aircraftNoise"` // dB(A)
	NoiseCategory string        `json:"noiseCategory"` // Quiet, Moderate, Loud, Very Loud
	ExceedsLimit  bool          `json:"exceedsLimit"`  // Above 55 dB(A) limit
	Sources       []NoiseSource `json:"sources"`
}

// NoiseSource represents a specific noise contributor
type NoiseSource struct {
	Type       string  `json:"type"`
	Distance   float64 `json:"distance"`   // meters
	NoiseLevel float64 `json:"noiseLevel"` // dB(A)
}

// GeminiSummary represents the AI-generated location summary
type GeminiSummary struct {
	Summary   string `json:"summary"`
	Generated bool   `json:"generated"`
	Error     string `json:"error,omitempty"`
}

// KadasterObjectInfo represents comprehensive property data from Kadaster Objectinformatie API
type KadasterObjectInfo struct {
	OwnerName          string  `json:"ownerName"`
	CadastralReference string  `json:"cadastralReference"`
	WOZValue           float64 `json:"wozValue"`
	EnergyLabel        string  `json:"energyLabel"`
	MunicipalTaxes     float64 `json:"municipalTaxes"`
	SurfaceArea        float64 `json:"surfaceArea"`
	PlotSize           float64 `json:"plotSize"`
	BuildingType       string  `json:"buildingType"`
	BuildYear          int     `json:"buildYear"`
}
