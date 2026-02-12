package scoring

import (
	"fmt"
	"math"

	"github.com/iman-hussain/nethaddress/backend/pkg/aggregator"
)

// PropertyScores contains all calculated scores for a property
type PropertyScores struct {
	ESGScore         float64        `json:"esgScore"`         // 0-100
	ProfitScore      float64        `json:"profitScore"`      // 0-100
	OpportunityScore float64        `json:"opportunityScore"` // 0-100
	OverallScore     float64        `json:"overallScore"`     // 0-100
	RiskLevel        string         `json:"riskLevel"`        // Low, Medium, High, Very High
	Breakdown        ScoreBreakdown `json:"breakdown"`
	Recommendations  []string       `json:"recommendations"`
}

// ScoreBreakdown provides detailed scoring components
type ScoreBreakdown struct {
	ESG         ESGBreakdown         `json:"esg"`
	Profit      ProfitBreakdown      `json:"profit"`
	Opportunity OpportunityBreakdown `json:"opportunity"`
}

// ESGBreakdown details environmental, social, and governance factors
type ESGBreakdown struct {
	EnergyEfficiency  float64 `json:"energyEfficiency"`  // 0-100
	EnvironmentalRisk float64 `json:"environmentalRisk"` // 0-100 (higher is better)
	SocialLivability  float64 `json:"socialLivability"`  // 0-100
	Sustainability    float64 `json:"sustainability"`    // 0-100
	FloodRisk         float64 `json:"floodRisk"`         // 0-100 (higher is safer)
	AirQuality        float64 `json:"airQuality"`        // 0-100
	NoiseLevel        float64 `json:"noiseLevel"`        // 0-100 (higher is quieter)
	GreenSpaceAccess  float64 `json:"greenSpaceAccess"`  // 0-100
}

// ProfitBreakdown details investment and profit potential
type ProfitBreakdown struct {
	CurrentValue      float64 `json:"currentValue"`      // EUR
	MarketValue       float64 `json:"marketValue"`       // EUR
	PriceAppreciation float64 `json:"priceAppreciation"` // 0-100
	RentalYield       float64 `json:"rentalYield"`       // percentage
	MarketDemand      float64 `json:"marketDemand"`      // 0-100
	LiquidityScore    float64 `json:"liquidityScore"`    // 0-100
	CapitalGrowth     float64 `json:"capitalGrowth"`     // 0-100
}

// OpportunityBreakdown details development and improvement opportunities
type OpportunityBreakdown struct {
	DevelopmentPotential float64 `json:"developmentPotential"` // 0-100
	RenovationROI        float64 `json:"renovationROI"`        // 0-100
	EnergyUpgradeROI     float64 `json:"energyUpgradeROI"`     // 0-100
	NeighborhoodGrowth   float64 `json:"neighborhoodGrowth"`   // 0-100
	Accessibility        float64 `json:"accessibility"`        // 0-100
	AmenitiesScore       float64 `json:"amenitiesScore"`       // 0-100
	FutureDevelopment    float64 `json:"futureDevelopment"`    // 0-100
}

// EnhancedScoringEngine calculates comprehensive property scores
type EnhancedScoringEngine struct{}

// NewEnhancedScoringEngine creates a new enhanced scoring engine
func NewEnhancedScoringEngine() *EnhancedScoringEngine {
	return &EnhancedScoringEngine{}
}

// CalculateComprehensiveScores computes all scores for comprehensive property data
func (se *EnhancedScoringEngine) CalculateComprehensiveScores(data *aggregator.ComprehensivePropertyData) *PropertyScores {
	scores := &PropertyScores{
		Breakdown:       ScoreBreakdown{},
		Recommendations: []string{},
	}

	// Calculate ESG Score
	scores.ESGScore, scores.Breakdown.ESG = se.calculateESGScore(data)

	// Calculate Profit Score
	scores.ProfitScore, scores.Breakdown.Profit = se.calculateProfitScore(data)

	// Calculate Opportunity Score
	scores.OpportunityScore, scores.Breakdown.Opportunity = se.calculateOpportunityScore(data)

	// Calculate Overall Score (weighted average)
	scores.OverallScore = (scores.ESGScore*0.3 + scores.ProfitScore*0.4 + scores.OpportunityScore*0.3)

	// Determine Risk Level
	scores.RiskLevel = se.calculateRiskLevel(data)

	// Generate Recommendations
	scores.Recommendations = se.generateRecommendations(data, scores)

	return scores
}

func (se *EnhancedScoringEngine) calculateESGScore(data *aggregator.ComprehensivePropertyData) (float64, ESGBreakdown) {
	breakdown := ESGBreakdown{}

	// Energy Efficiency (from energy label)
	if data.EnergyClimate != nil {
		breakdown.EnergyEfficiency = se.energyLabelToScore(data.EnergyClimate.EnergyLabel)
	} else {
		breakdown.EnergyEfficiency = 50.0 // Neutral if unknown
	}

	// Environmental Risk (inverse of contamination, subsidence, etc.)
	envRisk := 100.0
	if data.SoilQuality != nil && data.SoilQuality.ContaminationLevel == "Severe" {
		envRisk -= 40
	} else if data.SoilQuality != nil && data.SoilQuality.ContaminationLevel == "Moderate" {
		envRisk -= 20
	}
	if data.Subsidence != nil && data.Subsidence.StabilityRating == "High risk" {
		envRisk -= 30
	}
	breakdown.EnvironmentalRisk = math.Max(0, envRisk)

	// Social Livability (safety + amenities + education)
	livability := 50.0
	if data.Safety != nil {
		livability = data.Safety.SafetyScore * 0.4
	}
	if data.Facilities != nil {
		livability += data.Facilities.AmenitiesScore * 0.3
	}
	if data.Education != nil {
		livability += data.Education.AverageQuality * 10 * 0.3
	}
	breakdown.SocialLivability = math.Min(100, livability)

	// Sustainability (solar potential + energy efficiency)
	sustainability := breakdown.EnergyEfficiency * 0.6
	if data.SolarPotential != nil && data.SolarPotential.SolarRadiation > 0 {
		// Normalize solar radiation (good values 400-600 W/m²)
		solarScore := math.Min(100, (data.SolarPotential.SolarRadiation/600.0)*100)
		sustainability += solarScore * 0.4
	}
	breakdown.Sustainability = sustainability

	// Flood Risk (inverse - higher score is safer)
	if data.FloodRisk != nil {
		switch data.FloodRisk.RiskLevel {
		case "Low":
			breakdown.FloodRisk = 90
		case "Medium":
			breakdown.FloodRisk = 60
		case "High":
			breakdown.FloodRisk = 30
		case "Very High":
			breakdown.FloodRisk = 10
		default:
			breakdown.FloodRisk = 70
		}
	} else {
		breakdown.FloodRisk = 70 // Assume moderate if unknown
	}

	// Air Quality (AQI to score - lower AQI is better)
	if data.AirQuality != nil {
		// AQI: 0-50 Good, 51-100 Moderate, 101+ Unhealthy
		aqi := float64(data.AirQuality.AQI)
		breakdown.AirQuality = math.Max(0, 100-(aqi*0.8))
	} else {
		breakdown.AirQuality = 70
	}

	// Noise Level (inverse - higher score is quieter)
	if data.NoisePollution != nil {
		// Noise below 50 dB is good, above 65 is bad
		noise := data.NoisePollution.TotalNoise
		if noise < 50 {
			breakdown.NoiseLevel = 100
		} else if noise < 55 {
			breakdown.NoiseLevel = 80
		} else if noise < 60 {
			breakdown.NoiseLevel = 60
		} else if noise < 65 {
			breakdown.NoiseLevel = 40
		} else {
			breakdown.NoiseLevel = 20
		}
	} else {
		breakdown.NoiseLevel = 70
	}

	// Green Space Access
	if data.GreenSpaces != nil {
		breakdown.GreenSpaceAccess = data.GreenSpaces.GreenPercentage
		// Boost if park is nearby
		if data.GreenSpaces.ParkDistance < 500 {
			breakdown.GreenSpaceAccess = math.Min(100, breakdown.GreenSpaceAccess+20)
		}
	} else {
		breakdown.GreenSpaceAccess = 50
	}

	// Calculate overall ESG score (weighted average)
	esgScore := (breakdown.EnergyEfficiency*0.20 +
		breakdown.EnvironmentalRisk*0.15 +
		breakdown.SocialLivability*0.15 +
		breakdown.Sustainability*0.15 +
		breakdown.FloodRisk*0.10 +
		breakdown.AirQuality*0.10 +
		breakdown.NoiseLevel*0.10 +
		breakdown.GreenSpaceAccess*0.05)

	return esgScore, breakdown
}

func (se *EnhancedScoringEngine) calculateProfitScore(data *aggregator.ComprehensivePropertyData) (float64, ProfitBreakdown) {
	breakdown := ProfitBreakdown{}

	// Current Value (WOZ or Kadaster)
	if data.WOZData != nil {
		breakdown.CurrentValue = data.WOZData.WOZValue
	} else if data.KadasterInfo != nil {
		breakdown.CurrentValue = data.KadasterInfo.WOZValue
	}

	// Market Value (Matrixian)
	if data.MarketValuation != nil {
		breakdown.MarketValue = data.MarketValuation.MarketValue
	} else {
		breakdown.MarketValue = breakdown.CurrentValue
	}

	// Price Appreciation (from transaction history)
	if data.TransactionHistory != nil && len(data.TransactionHistory.Transactions) > 0 {
		// Calculate average annual appreciation
		firstPrice := data.TransactionHistory.Transactions[len(data.TransactionHistory.Transactions)-1].PurchasePrice
		currentPrice := breakdown.MarketValue
		if firstPrice > 0 && currentPrice > firstPrice {
			appreciation := ((currentPrice - firstPrice) / firstPrice) * 100
			breakdown.PriceAppreciation = math.Min(100, appreciation*2) // Scale to 0-100
		} else {
			breakdown.PriceAppreciation = 50
		}
	} else {
		breakdown.PriceAppreciation = 50
	}

	// Rental Yield (estimate based on location and type)
	if breakdown.MarketValue > 0 {
		// Netherlands average rental yield is 3-5%
		breakdown.RentalYield = 4.0 // Default estimate
	}

	// Market Demand (based on demographics and building activity)
	demand := 50.0
	if data.Population != nil && data.Population.TotalPopulation > 10000 {
		demand += 20
	}
	if data.BuildingPermits != nil && data.BuildingPermits.GrowthTrend == "Increasing" {
		demand += 20
	}
	if data.StatLineData != nil && data.StatLineData.EmploymentRate > 75 {
		demand += 10
	}
	breakdown.MarketDemand = math.Min(100, demand)

	// Liquidity Score (based on location desirability)
	liquidity := 50.0
	if data.PublicTransport != nil && len(data.PublicTransport.NearestStops) > 2 {
		liquidity += 15
	}
	if data.Facilities != nil && data.Facilities.AmenitiesScore > 70 {
		liquidity += 20
	}
	if data.StatLineData != nil && data.StatLineData.AverageIncome > 40000 {
		liquidity += 15
	}
	breakdown.LiquidityScore = math.Min(100, liquidity)

	// Capital Growth (neighborhood trends)
	growth := 50.0
	if data.BuildingPermits != nil {
		switch data.BuildingPermits.GrowthTrend {
		case "Increasing":
			growth = 80
		case "Stable":
			growth = 60
		case "Decreasing":
			growth = 30
		}
	}
	breakdown.CapitalGrowth = growth

	// Calculate overall profit score
	profitScore := ((breakdown.PriceAppreciation)*0.25 +
		(breakdown.MarketDemand)*0.25 +
		(breakdown.LiquidityScore)*0.20 +
		(breakdown.CapitalGrowth)*0.20 +
		(breakdown.RentalYield*10)*0.10) // Scale rental yield to 0-100

	return profitScore, breakdown
}

func (se *EnhancedScoringEngine) calculateOpportunityScore(data *aggregator.ComprehensivePropertyData) (float64, OpportunityBreakdown) {
	breakdown := OpportunityBreakdown{}

	// Development Potential (zoning and building rights)
	development := 50.0
	if data.LandUse != nil {
		if data.LandUse.BuildingRights != nil {
			if data.LandUse.BuildingRights.CanExpand {
				development += 25
			}
			if data.LandUse.BuildingRights.CanSubdivide {
				development += 25
			}
		}
	}
	breakdown.DevelopmentPotential = math.Min(100, development)

	// Renovation ROI (based on current condition and energy label)
	renovation := 50.0
	if data.EnergyClimate != nil {
		label := data.EnergyClimate.EnergyLabel
		if label == "E" || label == "F" || label == "G" {
			renovation = 85 // High ROI for poor energy labels
		} else if label == "D" || label == "C" {
			renovation = 65
		} else {
			renovation = 30 // Low ROI if already efficient
		}
	}
	breakdown.RenovationROI = renovation

	// Energy Upgrade ROI (from sustainability data)
	if data.Sustainability != nil {
		if data.Sustainability.PaybackPeriod > 0 && data.Sustainability.PaybackPeriod < 10 {
			breakdown.EnergyUpgradeROI = 100 - (data.Sustainability.PaybackPeriod * 10)
		} else if data.Sustainability.PaybackPeriod >= 10 {
			breakdown.EnergyUpgradeROI = 30
		} else {
			breakdown.EnergyUpgradeROI = 50
		}
	} else {
		breakdown.EnergyUpgradeROI = 50
	}

	// Neighborhood Growth
	growth := 50.0
	if data.BuildingPermits != nil {
		growth = float64(data.BuildingPermits.NewConstruction) / 10.0 // Scale based on permits
		if data.BuildingPermits.GrowthTrend == "Increasing" {
			growth += 30
		}
	}
	if data.StatLineData != nil {
		if data.StatLineData.Population > 50000 {
			growth += 10
		}
	}
	breakdown.NeighborhoodGrowth = math.Min(100, growth)

	// Accessibility
	accessibility := 50.0
	if data.PublicTransport != nil {
		stopCount := len(data.PublicTransport.NearestStops)
		accessibility = math.Min(100, 50+float64(stopCount)*10)
	}
	if len(data.TrafficData) > 0 {
		// Check for good traffic flow
		avgSpeed := 0.0
		for _, traffic := range data.TrafficData {
			avgSpeed += traffic.AverageSpeed
		}
		avgSpeed /= float64(len(data.TrafficData))
		if avgSpeed > 40 {
			accessibility = math.Min(100, accessibility+10)
		}
	}
	breakdown.Accessibility = accessibility

	// Amenities Score
	if data.Facilities != nil {
		breakdown.AmenitiesScore = data.Facilities.AmenitiesScore
	} else {
		breakdown.AmenitiesScore = 50
	}

	// Future Development (from land use plans)
	futureDev := 50.0
	if data.LandUse != nil && len(data.LandUse.FuturePlans) > 0 {
		for _, plan := range data.LandUse.FuturePlans {
			if plan.Status == "Approved" && plan.Impact == "Positive" {
				futureDev += 15
			}
		}
	}
	breakdown.FutureDevelopment = math.Min(100, futureDev)

	// Calculate overall opportunity score
	opportunityScore := (breakdown.DevelopmentPotential*0.20 +
		breakdown.RenovationROI*0.15 +
		breakdown.EnergyUpgradeROI*0.15 +
		breakdown.NeighborhoodGrowth*0.15 +
		breakdown.Accessibility*0.15 +
		breakdown.AmenitiesScore*0.10 +
		breakdown.FutureDevelopment*0.10)

	return opportunityScore, breakdown
}

func (se *EnhancedScoringEngine) calculateRiskLevel(data *aggregator.ComprehensivePropertyData) string {
	riskPoints := 0

	// Flood risk
	if data.FloodRisk != nil {
		switch data.FloodRisk.RiskLevel {
		case "High", "Very High":
			riskPoints += 3
		case "Medium":
			riskPoints += 1
		}
	}

	// Subsidence
	if data.Subsidence != nil && data.Subsidence.StabilityRating == "High risk" {
		riskPoints += 2
	}

	// Soil contamination
	if data.SoilQuality != nil {
		switch data.SoilQuality.ContaminationLevel {
		case "Severe":
			riskPoints += 3
		case "Moderate":
			riskPoints += 1
		}
	}

	// Safety
	if data.Safety != nil && data.Safety.SafetyScore < 40 {
		riskPoints += 2
	}

	// Market risk (declining area)
	if data.BuildingPermits != nil && data.BuildingPermits.GrowthTrend == "Decreasing" {
		riskPoints += 1
	}

	// Categorize risk
	if riskPoints >= 6 {
		return "Very High"
	} else if riskPoints >= 4 {
		return "High"
	} else if riskPoints >= 2 {
		return "Medium"
	}
	return "Low"
}

func (se *EnhancedScoringEngine) generateRecommendations(data *aggregator.ComprehensivePropertyData, scores *PropertyScores) []string {
	recommendations := []string{}

	// Energy recommendations
	if scores.Breakdown.ESG.EnergyEfficiency < 60 {
		recommendations = append(recommendations, "Consider energy efficiency improvements (insulation, double glazing, solar panels)")
	}

	// Sustainability recommendations
	if data.Sustainability != nil && data.Sustainability.TotalCostSavings > 1000 {
		recommendations = append(recommendations, fmt.Sprintf("Energy upgrades could save €%.0f/year", data.Sustainability.TotalCostSavings))
	}

	// Flood risk
	if data.FloodRisk != nil && (data.FloodRisk.RiskLevel == "High" || data.FloodRisk.RiskLevel == "Very High") {
		recommendations = append(recommendations, "High flood risk - ensure comprehensive insurance coverage")
	}

	// Development opportunities
	if scores.Breakdown.Opportunity.DevelopmentPotential > 70 {
		recommendations = append(recommendations, "Property has significant development potential - check zoning regulations")
	}

	// Market timing
	if scores.ProfitScore > 75 {
		recommendations = append(recommendations, "Strong market conditions - good time for investment or sale")
	} else if scores.ProfitScore < 40 {
		recommendations = append(recommendations, "Weak market indicators - consider holding or substantial improvements")
	}

	// Accessibility
	if scores.Breakdown.Opportunity.Accessibility < 50 {
		recommendations = append(recommendations, "Limited accessibility may affect resale value")
	}

	// Capital improvements
	if scores.Breakdown.Opportunity.RenovationROI > 70 {
		recommendations = append(recommendations, "High ROI potential for renovations - prioritize kitchen and bathroom upgrades")
	}

	return recommendations
}

func (se *EnhancedScoringEngine) energyLabelToScore(label string) float64 {
	switch label {
	case "A++++", "A+++", "A++", "A+":
		return 95
	case "A":
		return 85
	case "B":
		return 75
	case "C":
		return 60
	case "D":
		return 45
	case "E":
		return 30
	case "F":
		return 20
	case "G":
		return 10
	default:
		return 50
	}
}
