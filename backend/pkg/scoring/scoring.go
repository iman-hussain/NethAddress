package scoring

import (
	"strings"

	"github.com/iman-hussain/AddressIQ/backend/pkg/models"
)

func CalculateScore(data *models.AggregatedData) *models.PropertyScore {
	score := &models.PropertyScore{}
	// Weighted ESG score
	energyScore := scoreEnergy(data)
	noiseScore := scoreNoise(data)
	soilScore := scoreSoil(data)
	esg := (energyScore * 0.4) + (noiseScore * 0.3) + (soilScore * 0.3)
	score.ESG = esg

	// Weighted Investment score
	inv := scoreInvestment(data)
	cbs := scoreCBS(data)
	monument := scoreMonument(data)
	score.Investment = (inv * 0.5) + (cbs * 0.3) + (monument * 0.2)

	// Weighted Viability score
	zoning := scoreZoning(data)
	asbestos := scoreAsbestos(data)
	score.Viability = (zoning * 0.7) + (asbestos * 0.3)

	return score
}

func scoreZoning(data *models.AggregatedData) float64 {
	info := data.PDOKData.ZoningInfo
	switch strings.ToLower(info) {
	case "wonen", "residential":
		return 8.0
	case "gemengd-1":
		return 6.0
	case "industrie", "industrial":
		return 2.0
	case "groen":
		return 7.0
	default:
		return 4.0
	}
}

func scoreEnergy(data *models.AggregatedData) float64 {
	// Example: energy label scoring
	if data.EnergyJSON != nil && *data.EnergyJSON != "" {
		// Parse label from JSON (mock)
		label := "A" // TODO: parse real label
		switch label {
		case "A++", "A+", "A":
			return 10
		case "B":
			return 8
		case "C":
			return 6
		default:
			return 2
		}
	}
	return 5
}

func scoreNoise(data *models.AggregatedData) float64 {
	// Example: noise dB scoring
	db := 50.0 // TODO: parse real dB from NoiseJSON
	if db > 70 {
		return 1
	}
	return 10
}

func scoreSoil(data *models.AggregatedData) float64 {
	// Example: soil contamination scoring
	contaminated := false // TODO: parse from SoilJSON
	if contaminated {
		return 1
	}
	return 10
}

func scoreInvestment(data *models.AggregatedData) float64 {
	if strings.Contains(strings.ToLower(data.BAGData.Address), "straat") {
		return 7.5
	}
	return 5.0
}

func scoreCBS(data *models.AggregatedData) float64 {
	// Example: higher income, higher score
	income := 30000.0 // TODO: parse from CBSJSON
	if income > 40000 {
		return 10
	} else if income > 30000 {
		return 7
	}
	return 4
}

func scoreMonument(data *models.AggregatedData) float64 {
	// Example: monument status scoring
	isMonument := false // TODO: parse from MonumentJSON
	if isMonument {
		return 2
	}
	return 8
}

func scoreAsbestos(data *models.AggregatedData) float64 {
	hasAsbestos := false // TODO: parse from AsbestosJSON
	if hasAsbestos {
		return 1
	}
	return 10
}
