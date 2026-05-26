package verifier

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type SourceResult struct {
	SourceLevel string
	Weight      float64
	MatchRate   float64
}

func SourceConfidence(level string) float64 {
	switch strings.ToUpper(level) {
	case "HIGH":
		return 0.9
	case "MEDIUM":
		return 0.7
	case "LOW":
		return 0.5
	default:
		return 0.3
	}
}

type ConfidenceConfig struct {
	WSource    float64
	WMatch     float64
	WConflict  float64
	WFreshness float64
	WExpert    float64
}

func DefaultConfidenceConfig() ConfidenceConfig {
	return ConfidenceConfig{
		WSource:    0.3,
		WMatch:     0.2,
		WConflict:  0.2,
		WFreshness: 0.15,
		WExpert:    0.15,
	}
}

func sourceAuthorityScore(level string) float64 {
	switch strings.ToUpper(level) {
	case "HIGH":
		return 1.0
	case "MEDIUM":
		return 0.7
	case "LOW":
		return 0.4
	default:
		return 0.3
	}
}

func freshnessScore(fetchedAt string) float64 {
	if fetchedAt == "" {
		return 0.2
	}
	t, err := time.Parse(time.RFC3339, fetchedAt)
	if err != nil {
		return 0.2
	}
	days := time.Since(t).Hours() / 24
	switch {
	case days <= 30:
		return 1.0
	case days <= 90:
		return 0.8
	case days <= 365:
		return 0.5
	default:
		return 0.2
	}
}

func expertVerificationScore(verifiedBy string) float64 {
	switch strings.ToLower(verifiedBy) {
	case "expert":
		return 1.0
	case "auto":
		return 0.5
	default:
		return 0.0
	}
}

func CalculateConfidence(claim *models.PolicyClaim, config ConfidenceConfig) float64 {
	if claim == nil {
		return 0
	}

	sSource := sourceAuthorityScore(claim.SourceLevel)
	if sSource < 0 {
		sSource = 0
	}

	sMatch := claim.MatchRate
	if sMatch < 0 {
		sMatch = 0
	}
	if sMatch > 1 {
		sMatch = 1
	}

	sConflict := claim.ConflictScore
	if sConflict == 0 {
		sConflict = 1.0
	}
	if sConflict < 0 {
		sConflict = 0
	}
	if sConflict > 1 {
		sConflict = 1
	}

	sFreshness := freshnessScore(claim.FetchedAt)
	sExpert := expertVerificationScore(claim.VerifiedBy)

	score := config.WSource*sSource +
		config.WMatch*sMatch +
		config.WConflict*sConflict +
		config.WFreshness*sFreshness +
		config.WExpert*sExpert

	if score > 1.0 {
		score = 1.0
	}
	if score < 0 {
		score = 0
	}
	return score
}

func AggregateConfidence(sources []SourceResult) float64 {
	if len(sources) == 0 {
		return 0
	}

	bestLevel := "LOW"
	for _, s := range sources {
		if s.SourceLevel == "HIGH" {
			bestLevel = "HIGH"
			break
		}
		if s.SourceLevel == "MEDIUM" {
			bestLevel = "MEDIUM"
		}
	}

	var weightedMatch, totalWeight float64
	for _, s := range sources {
		weightedMatch += s.MatchRate * s.Weight
		totalWeight += s.Weight
	}
	avgMatchRate := 0.0
	if totalWeight > 0 {
		avgMatchRate = weightedMatch / totalWeight
	}

	claim := &models.PolicyClaim{
		SourceLevel:   bestLevel,
		MatchRate:     avgMatchRate,
		ConflictScore: 1.0,
	}

	return CalculateConfidence(claim, DefaultConfidenceConfig())
}

type StatusThresholds struct {
	Verified float64
	Pending  float64
}

func DefaultStatusThresholds() StatusThresholds {
	return StatusThresholds{
		Verified: 0.85,
		Pending:  0.6,
	}
}

func DecideStatusWithConfig(confidence float64, thresholds StatusThresholds) string {
	if confidence >= thresholds.Verified {
		return "verified"
	}
	if confidence >= thresholds.Pending {
		return "pending_review"
	}
	return "unverified"
}

func DecideStatus(confidence float64) string {
	return DecideStatusWithConfig(confidence, DefaultStatusThresholds())
}

func FieldLevelDiff(valA, valB string) float64 {
	if valA == valB {
		return 0
	}
	if valA == "" || valB == "" {
		return 1.0
	}

	numA, errA := strconv.ParseFloat(valA, 64)
	numB, errB := strconv.ParseFloat(valB, 64)

	if errA != nil || errB != nil {
		return 1.0
	}

	if numA == 0 && numB == 0 {
		return 0
	}

	if numA == 0 || numB == 0 {
		return 1.0
	}

	diff := math.Abs(numA-numB) / math.Max(math.Abs(numA), math.Abs(numB))
	if diff > 1.0 {
		return 1.0
	}
	return diff
}
