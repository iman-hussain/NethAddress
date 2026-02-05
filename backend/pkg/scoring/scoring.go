package scoring

import (
	"encoding/json"
	"strings"

	"github.com/iman-hussain/AddressIQ/backend/pkg/apiclient"
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
	if data.EnergyJSON == nil || *data.EnergyJSON == "" {
		return 5.0 // Default score when no data available
	}

	var energy models.EnergyClimateData
	if err := json.Unmarshal([]byte(*data.EnergyJSON), &energy); err != nil {
		return 5.0
	}

	label := strings.ToUpper(energy.EnergyLabel)
	switch label {
	case "A++++", "A+++", "A++", "A+", "A":
		return 10.0
	case "B":
		return 8.0
	case "C":
		return 6.0
	case "D":
		return 4.0
	case "E":
		return 3.0
	case "F":
		return 2.0
	case "G":
		return 1.0
	default:
		return 5.0
	}
}

func scoreNoise(data *models.AggregatedData) float64 {
	if data.NoiseJSON == nil || *data.NoiseJSON == "" {
		return 5.0 // Default score when no data available
	}

	var noise models.NoisePollutionData
	if err := json.Unmarshal([]byte(*data.NoiseJSON), &noise); err != nil {
		return 5.0
	}

	db := noise.TotalNoise
	switch {
	case db <= 40:
		return 10.0
	case db <= 50:
		return 8.0
	case db <= 55:
		return 6.0
	case db <= 65:
		return 4.0
	case db <= 70:
		return 2.0
	default:
		return 1.0
	}
}

func scoreSoil(data *models.AggregatedData) float64 {
	if data.SoilJSON == nil || *data.SoilJSON == "" {
		return 5.0 // Default score when no data available
	}

	var soil models.SoilQualityData
	if err := json.Unmarshal([]byte(*data.SoilJSON), &soil); err != nil {
		return 5.0
	}

	level := strings.ToLower(soil.ContaminationLevel)
	switch level {
	case "clean":
		return 10.0
	case "light":
		return 7.0
	case "moderate":
		return 4.0
	case "severe":
		return 1.0
	default:
		if soil.RestrictedUse {
			return 2.0
		}
		return 5.0
	}
}

func scoreInvestment(data *models.AggregatedData) float64 {
	if strings.Contains(strings.ToLower(data.BAGData.Address), "straat") {
		return 7.5
	}
	return 5.0
}

func scoreCBS(data *models.AggregatedData) float64 {
	if data.CBSJSON == nil || *data.CBSJSON == "" {
		return 5.0 // Default score when no data available
	}

	var cbs models.CBSData
	if err := json.Unmarshal([]byte(*data.CBSJSON), &cbs); err != nil {
		return 5.0
	}

	income := cbs.AvgIncome
	switch {
	case income >= 50000:
		return 10.0
	case income >= 40000:
		return 8.0
	case income >= 35000:
		return 7.0
	case income >= 30000:
		return 6.0
	case income >= 25000:
		return 5.0
	case income >= 20000:
		return 4.0
	default:
		return 3.0
	}
}

func scoreMonument(data *models.AggregatedData) float64 {
	if data.MonumentJSON == nil || *data.MonumentJSON == "" {
		return 8.0 // Default: assume not a monument
	}

	var monument models.MonumentData
	if err := json.Unmarshal([]byte(*data.MonumentJSON), &monument); err != nil {
		return 8.0
	}

	// Monuments have more restrictions on modifications, impacting investment flexibility
	if monument.IsMonument {
		return 2.0
	}
	return 8.0
}

func scoreAsbestos(data *models.AggregatedData) float64 {
	if data.AsbestosJSON == nil || *data.AsbestosJSON == "" {
		return 10.0 // Default: assume no asbestos
	}

	var asbestos apiclient.AsbestosData
	if err := json.Unmarshal([]byte(*data.AsbestosJSON), &asbestos); err != nil {
		return 10.0
	}

	// Properties with asbestos reports have significant remediation concerns
	if asbestos.HasAsbestosReport {
		return 1.0
	}
	return 10.0
}
