package models

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
