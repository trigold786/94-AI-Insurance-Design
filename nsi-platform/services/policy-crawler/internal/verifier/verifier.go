package verifier

import (
	"math"
	"strconv"
	"strings"
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

func AggregateConfidence(sources []SourceResult) float64 {
	if len(sources) == 0 {
		return 0
	}

	var weightedSum, totalWeight float64
	for _, s := range sources {
		baseConf := SourceConfidence(s.SourceLevel)
		adjusted := baseConf * s.MatchRate * s.Weight
		weightedSum += adjusted * s.Weight
		totalWeight += s.Weight
	}

	if totalWeight == 0 {
		return 0
	}

	score := weightedSum / totalWeight

	if score > 1.0 {
		score = 1.0
	}
	return score
}

func DecideStatus(confidence float64) string {
	if confidence >= 0.9 {
		return "verified"
	}
	if confidence >= 0.7 {
		return "pending_review"
	}
	return "unverified"
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
