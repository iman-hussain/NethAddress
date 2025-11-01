package scoring

import (
	"testing"

	"github.com/iman-hussain/AddressIQ/backend/pkg/aggregator"
	"github.com/iman-hussain/AddressIQ/backend/pkg/apiclient"
)

func TestCalculateComprehensiveScores_Basic(t *testing.T) {
	engine := NewEnhancedScoringEngine()

	// Create minimal test data
	data := &aggregator.ComprehensivePropertyData{
		EnergyClimate: &apiclient.EnergyClimateData{
			EnergyLabel: "B",
		},
		FloodRisk: &apiclient.FloodRiskData{
			RiskLevel: "Low",
		},
	}

	scores := engine.CalculateComprehensiveScores(data)

	// Test that scores are within valid range (0-100)
	if scores.ESGScore < 0 || scores.ESGScore > 100 {
		t.Errorf("ESG score out of range: %f", scores.ESGScore)
	}
	if scores.ProfitScore < 0 || scores.ProfitScore > 100 {
		t.Errorf("Profit score out of range: %f", scores.ProfitScore)
	}
	if scores.OpportunityScore < 0 || scores.OpportunityScore > 100 {
		t.Errorf("Opportunity score out of range: %f", scores.OpportunityScore)
	}
	if scores.OverallScore < 0 || scores.OverallScore > 100 {
		t.Errorf("Overall score out of range: %f", scores.OverallScore)
	}

	// Test risk level is valid
	validRiskLevels := map[string]bool{
		"Low": true, "Medium": true, "High": true, "Very High": true,
	}
	if !validRiskLevels[scores.RiskLevel] {
		t.Errorf("Invalid risk level: %s", scores.RiskLevel)
	}
}

func TestEnergyLabelToScore(t *testing.T) {
	engine := NewEnhancedScoringEngine()

	tests := []struct {
		label       string
		expectedMin float64
		expectedMax float64
	}{
		{"A++++", 90, 100},
		{"A", 80, 90},
		{"B", 70, 80},
		{"C", 55, 65},
		{"D", 40, 50},
		{"E", 25, 35},
		{"F", 15, 25},
		{"G", 5, 15},
		{"Unknown", 45, 55},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			score := engine.energyLabelToScore(tt.label)
			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("Label %s: expected score between %f and %f, got %f",
					tt.label, tt.expectedMin, tt.expectedMax, score)
			}
		})
	}
}

func TestCalculateRiskLevel(t *testing.T) {
	engine := NewEnhancedScoringEngine()

	tests := []struct {
		name     string
		data     *aggregator.ComprehensivePropertyData
		expected string
	}{
		{
			name: "Low Risk",
			data: &aggregator.ComprehensivePropertyData{
				FloodRisk: &apiclient.FloodRiskData{
					RiskLevel: "Low",
				},
				Safety: &apiclient.SafetyData{
					SafetyScore: 80,
				},
			},
			expected: "Low",
		},
		{
			name: "High Risk",
			data: &aggregator.ComprehensivePropertyData{
				FloodRisk: &apiclient.FloodRiskData{
					RiskLevel: "Very High",
				},
				SoilQuality: &apiclient.SoilQualityData{
					ContaminationLevel: "Severe",
				},
				Subsidence: &apiclient.SubsidenceData{
					StabilityRating: "High risk",
				},
			},
			expected: "Very High",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			riskLevel := engine.calculateRiskLevel(tt.data)
			if riskLevel != tt.expected {
				t.Errorf("Expected risk level '%s', got '%s'", tt.expected, riskLevel)
			}
		})
	}
}
