package scoring

import (
	"math"
	"testing"

	"github.com/iman-hussain/nethaddress/backend/pkg/models"
)

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

// Helper to create string pointers for JSON test data
func strPtr(s string) *string {
	return &s
}

func TestCalculateScore_Wonen(t *testing.T) {
	agg := &models.AggregatedData{
		BAGData:  models.BAGData{Address: "Teststraat 10, 1234AB Testdorp"},
		PDOKData: models.PDOKData{ZoningInfo: "Wonen"},
	}
	score := CalculateScore(agg)
	// Viability = 0.7*8 (wonen) + 0.3*10 (no asbestos) = 8.6
	if !floatEquals(score.Viability, 8.6) {
		t.Errorf("Expected Viability 8.6, got %v", score.Viability)
	}
	// ESG = 0.4*5 (default energy) + 0.3*5 (default noise) + 0.3*5 (default soil) = 5.0
	if !floatEquals(score.ESG, 5.0) {
		t.Errorf("Expected ESG 5.0, got %v", score.ESG)
	}
	// Investment = 0.5*7.5 (straat) + 0.3*5 (default cbs) + 0.2*8 (default monument) = 6.85
	if !floatEquals(score.Investment, 6.85) {
		t.Errorf("Expected Investment 6.85, got %v", score.Investment)
	}
}

func TestCalculateScore_Industrie(t *testing.T) {
	agg := &models.AggregatedData{
		BAGData:  models.BAGData{Address: "Industrieweg 1, 5678CD Industriedorp"},
		PDOKData: models.PDOKData{ZoningInfo: "Industrie"},
	}
	score := CalculateScore(agg)
	// Viability = 0.7*2 (industrie) + 0.3*10 (no asbestos) = 4.4
	if !floatEquals(score.Viability, 4.4) {
		t.Errorf("Expected Viability 4.4, got %v", score.Viability)
	}
	// ESG = 0.4*5 + 0.3*5 + 0.3*5 = 5.0
	if !floatEquals(score.ESG, 5.0) {
		t.Errorf("Expected ESG 5.0, got %v", score.ESG)
	}
	// Investment = 0.5*5 (no straat) + 0.3*5 (default cbs) + 0.2*8 (default monument) = 5.6
	if !floatEquals(score.Investment, 5.6) {
		t.Errorf("Expected Investment 5.6, got %v", score.Investment)
	}
}

func TestCalculateScore_GroenESG(t *testing.T) {
	agg := &models.AggregatedData{
		BAGData:  models.BAGData{Address: "Groenlaan 5, 9999GG Groendorp"},
		PDOKData: models.PDOKData{ZoningInfo: "Groen"},
	}
	score := CalculateScore(agg)
	// Viability = 0.7*7 (groen) + 0.3*10 (no asbestos) = 7.9
	if !floatEquals(score.Viability, 7.9) {
		t.Errorf("Expected Viability 7.9, got %v", score.Viability)
	}
	// ESG = 0.4*5 + 0.3*5 + 0.3*5 = 5.0
	if !floatEquals(score.ESG, 5.0) {
		t.Errorf("Expected ESG 5.0, got %v", score.ESG)
	}
	// Investment = 0.5*5 (no straat) + 0.3*5 (default cbs) + 0.2*8 (default monument) = 5.6
	if !floatEquals(score.Investment, 5.6) {
		t.Errorf("Expected Investment 5.6, got %v", score.Investment)
	}
}

func TestCalculateScore_WithEnergyData(t *testing.T) {
	energyJSON := `{"energyLabel": "A", "efficiencyScore": 85.0}`
	agg := &models.AggregatedData{
		BAGData:    models.BAGData{Address: "Teststraat 10, 1234AB Testdorp"},
		PDOKData:   models.PDOKData{ZoningInfo: "Wonen"},
		EnergyJSON: strPtr(energyJSON),
	}
	score := CalculateScore(agg)
	// ESG = 0.4*10 (A label) + 0.3*5 (default noise) + 0.3*5 (default soil) = 7.0
	if !floatEquals(score.ESG, 7.0) {
		t.Errorf("Expected ESG 7.0, got %v", score.ESG)
	}
}

func TestCalculateScore_WithNoiseData(t *testing.T) {
	noiseJSON := `{"totalNoise": 45.0, "noiseCategory": "Quiet"}`
	agg := &models.AggregatedData{
		BAGData:   models.BAGData{Address: "Teststraat 10, 1234AB Testdorp"},
		PDOKData:  models.PDOKData{ZoningInfo: "Wonen"},
		NoiseJSON: strPtr(noiseJSON),
	}
	score := CalculateScore(agg)
	// ESG = 0.4*5 (default energy) + 0.3*8 (45dB) + 0.3*5 (default soil) = 5.9
	if !floatEquals(score.ESG, 5.9) {
		t.Errorf("Expected ESG 5.9, got %v", score.ESG)
	}
}

func TestCalculateScore_WithSoilData(t *testing.T) {
	soilJSON := `{"contaminationLevel": "Clean", "restrictedUse": false}`
	agg := &models.AggregatedData{
		BAGData:  models.BAGData{Address: "Teststraat 10, 1234AB Testdorp"},
		PDOKData: models.PDOKData{ZoningInfo: "Wonen"},
		SoilJSON: strPtr(soilJSON),
	}
	score := CalculateScore(agg)
	// ESG = 0.4*5 (default energy) + 0.3*5 (default noise) + 0.3*10 (clean soil) = 6.5
	if !floatEquals(score.ESG, 6.5) {
		t.Errorf("Expected ESG 6.5, got %v", score.ESG)
	}
}

func TestCalculateScore_WithCBSData(t *testing.T) {
	cbsJSON := `{"AvgIncome": 45000.0, "PopulationDensity": 1500.0}`
	agg := &models.AggregatedData{
		BAGData:  models.BAGData{Address: "Teststraat 10, 1234AB Testdorp"},
		PDOKData: models.PDOKData{ZoningInfo: "Wonen"},
		CBSJSON:  strPtr(cbsJSON),
	}
	score := CalculateScore(agg)
	// Investment = 0.5*7.5 (straat) + 0.3*8 (45k income) + 0.2*8 (default monument) = 7.75
	if !floatEquals(score.Investment, 7.75) {
		t.Errorf("Expected Investment 7.75, got %v", score.Investment)
	}
}

func TestCalculateScore_WithMonumentData(t *testing.T) {
	monumentJSON := `{"isMonument": true, "type": "Rijksmonument"}`
	agg := &models.AggregatedData{
		BAGData:      models.BAGData{Address: "Teststraat 10, 1234AB Testdorp"},
		PDOKData:     models.PDOKData{ZoningInfo: "Wonen"},
		MonumentJSON: strPtr(monumentJSON),
	}
	score := CalculateScore(agg)
	// Investment = 0.5*7.5 (straat) + 0.3*5 (default cbs) + 0.2*2 (monument) = 5.65
	if !floatEquals(score.Investment, 5.65) {
		t.Errorf("Expected Investment 5.65, got %v", score.Investment)
	}
}

func TestCalculateScore_WithAsbestosData(t *testing.T) {
	asbestosJSON := `{"hasAsbestosReport": true}`
	agg := &models.AggregatedData{
		BAGData:      models.BAGData{Address: "Teststraat 10, 1234AB Testdorp"},
		PDOKData:     models.PDOKData{ZoningInfo: "Wonen"},
		AsbestosJSON: strPtr(asbestosJSON),
	}
	score := CalculateScore(agg)
	// Viability = 0.7*8 (wonen) + 0.3*1 (asbestos present) = 5.9
	if !floatEquals(score.Viability, 5.9) {
		t.Errorf("Expected Viability 5.9, got %v", score.Viability)
	}
}
