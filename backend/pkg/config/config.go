package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DatabaseURL    string `envconfig:"DATABASE_URL"`
	RedisURL       string `envconfig:"REDIS_URL"`
	FrontendOrigin string `envconfig:"FRONTEND_ORIGIN"`
	AdminSecret    string `envconfig:"ADMIN_SECRET"`

	// Property & Land Data APIs
	BagApiURL                string `envconfig:"BAG_API_URL"`
	KadasterObjectInfoApiURL string `envconfig:"KADASTER_OBJECTINFO_API_URL"`
	KadasterObjectInfoApiKey string `envconfig:"KADASTER_OBJECTINFO_API_KEY"`
	AltumWOZApiURL           string `envconfig:"ALTUM_WOZ_API_URL"`
	AltumWOZApiKey           string `envconfig:"ALTUM_WOZ_API_KEY"`
	MatrixianApiURL          string `envconfig:"MATRIXIAN_API_URL"`
	MatrixianApiKey          string `envconfig:"MATRIXIAN_API_KEY"`
	AltumTransactionApiURL   string `envconfig:"ALTUM_TRANSACTION_API_URL"`
	AltumTransactionApiKey   string `envconfig:"ALTUM_TRANSACTION_API_KEY"`

	// Weather & Climate Data
	KNMIWeatherApiURL string `envconfig:"KNMI_WEATHER_API_URL"`
	KNMIWeatherApiKey string `envconfig:"KNMI_WEATHER_API_KEY"`
	WeerliveApiURL    string `envconfig:"WEERLIVE_API_URL"`
	WeerliveApiKey    string `envconfig:"WEERLIVE_API_KEY"`
	KNMISolarApiURL   string `envconfig:"KNMI_SOLAR_API_URL"`
	KNMISolarApiKey   string `envconfig:"KNMI_SOLAR_API_KEY"`

	// Environmental & Soil Data
	WURSoilApiURL          string `envconfig:"WUR_SOIL_API_URL"`
	SkyGeoSubsidenceApiURL string `envconfig:"SKYGEO_SUBSIDENCE_API_URL"`
	SoilQualityApiURL      string `envconfig:"SOIL_QUALITY_API_URL"`
	BROSoilMapApiURL       string `envconfig:"BRO_SOIL_MAP_API_URL"`

	// Energy & Sustainability APIs
	AltumEnergyApiURL         string `envconfig:"ALTUM_ENERGY_API_URL"`
	AltumEnergyApiKey         string `envconfig:"ALTUM_ENERGY_API_KEY"`
	AltumSustainabilityApiURL string `envconfig:"ALTUM_SUSTAINABILITY_API_URL"`
	AltumSustainabilityApiKey string `envconfig:"ALTUM_SUSTAINABILITY_API_KEY"`
	EnergieLabelApiKey        string `envconfig:"ENERGIE_LABEL_API_KEY"`
	EnergieLabelApiURL        string `envconfig:"ENERGIE_LABEL_API_URL"`

	// Traffic & Mobility
	NDWTrafficApiURL string `envconfig:"NDW_TRAFFIC_API_URL"`
	OpenOVApiURL     string `envconfig:"OPENOV_API_URL"`
	ParkingApiURL    string `envconfig:"PARKING_API_URL"`

	// Population & Demographics
	CBSPopulationApiURL  string `envconfig:"CBS_POPULATION_API_URL"`
	CBSStatLineApiURL    string `envconfig:"CBS_STATLINE_API_URL"`
	CBSSquareStatsApiURL string `envconfig:"CBS_SQUARE_STATS_API_URL"`
	CBSApiURL            string `envconfig:"CBS_API_URL"`

	// Environmental Quality
	LuchtmeetnetApiURL   string `envconfig:"LUCHTMEETNET_API_URL"`
	NoisePollutionApiURL string `envconfig:"NOISE_POLLUTION_API_URL"`
	GeluidregisterApiURL string `envconfig:"GELUIDREGISTER_API_URL"`

	// Water & Flooding
	FloodRiskApiURL    string `envconfig:"FLOOD_RISK_API_URL"`
	DigitalDeltaApiURL string `envconfig:"DIGITAL_DELTA_API_URL"`

	// Safety & Aviation
	SafetyExperienceApiURL string `envconfig:"SAFETY_EXPERIENCE_API_URL"`
	SchipholApiURL         string `envconfig:"SCHIPHOL_API_URL"`
	SchipholApiKey         string `envconfig:"SCHIPHOL_API_KEY"`
	SchipholAppID          string `envconfig:"SCHIPHOL_APP_ID"`

	// Infrastructure & Facilities
	GreenSpacesApiURL     string `envconfig:"GREEN_SPACES_API_URL"`
	EducationApiURL       string `envconfig:"EDUCATION_API_URL"`
	BuildingPermitsApiURL string `envconfig:"BUILDING_PERMITS_API_URL"`
	FacilitiesApiURL      string `envconfig:"FACILITIES_API_URL"`

	// Elevation & Topography
	AHNHeightModelApiURL string `envconfig:"AHN_HEIGHT_MODEL_API_URL"`

	// Comprehensive Platforms
	PDOKApiURL     string `envconfig:"PDOK_API_URL"`
	StratopoApiURL string `envconfig:"STRATOPO_API_URL"`
	StratopoApiKey string `envconfig:"STRATOPO_API_KEY"`
	LandUseApiURL  string `envconfig:"LAND_USE_API_URL"`

	// Legacy/Existing
	ZoningApiURL     string `envconfig:"ZONING_API_URL"`
	BodemloketApiURL string `envconfig:"BODEMLOKET_API_URL"`
	MonumentenApiURL string `envconfig:"MONUMENTEN_API_URL"`

	// AI Summary
	GeminiApiKey string `envconfig:"GEMINI_API_KEY"`
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	// Validate required configuration (Fail-Fast)
	if cfg.BagApiURL == "" {
		return nil, fmt.Errorf("required environment variable BAG_API_URL is not set")
	}

	return &cfg, nil
}
