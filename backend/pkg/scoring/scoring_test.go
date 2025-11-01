package scoring

import (
	"math"
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func TestCalculateScore_Wonen(t *testing.T) {
	agg := &models.AggregatedData{
		BAGData:  models.BAGData{Address: "Teststraat 10, 1234AB Testdorp"},
		PDOKData: models.PDOKData{ZoningInfo: "Wonen"},
	}
	score := CalculateScore(agg)
	if !floatEquals(score.Viability, 8.6) {
		t.Errorf("Expected Viability 8.6, got %v", score.Viability)
	}
	if !floatEquals(score.ESG, 8.0) {
		t.Errorf("Expected ESG 8.0, got %v", score.ESG)
	}
	if !floatEquals(score.Investment, 6.55) {
		t.Errorf("Expected Investment 6.55, got %v", score.Investment)
	}
}

func TestCalculateScore_Industrie(t *testing.T) {
	agg := &models.AggregatedData{
		BAGData:  models.BAGData{Address: "Industrieweg 1, 5678CD Industriedorp"},
		PDOKData: models.PDOKData{ZoningInfo: "Industrie"},
	}
	score := CalculateScore(agg)
	if !floatEquals(score.Viability, 4.4) {
		t.Errorf("Expected Viability 4.4, got %v", score.Viability)
	}
	if !floatEquals(score.ESG, 8.0) {
		t.Errorf("Expected ESG 8.0, got %v", score.ESG)
	}
	if !floatEquals(score.Investment, 5.3) {
		t.Errorf("Expected Investment 5.3, got %v", score.Investment)
	}
}

func TestCalculateScore_GroenESG(t *testing.T) {
	agg := &models.AggregatedData{
		BAGData:  models.BAGData{Address: "Groenlaan 5, 9999GG Groendorp"},
		PDOKData: models.PDOKData{ZoningInfo: "Groen"},
	}
	score := CalculateScore(agg)
	if !floatEquals(score.Viability, 7.9) {
		t.Errorf("Expected Viability 7.9, got %v", score.Viability)
	}
	if !floatEquals(score.ESG, 8.0) {
		t.Errorf("Expected ESG 8.0, got %v", score.ESG)
	}
	if !floatEquals(score.Investment, 5.3) {
		t.Errorf("Expected Investment 5.3, got %v", score.Investment)
	}
}
