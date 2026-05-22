package verifier

import (
	"testing"
)

func TestSourceConfidenceHIGH(t *testing.T) {
	score := SourceConfidence("HIGH")
	if !almostEqual(score, 0.9, 0.01) {
		t.Errorf("expected 0.9 for HIGH, got %.2f", score)
	}
}

func TestSourceConfidenceMEDIUM(t *testing.T) {
	score := SourceConfidence("MEDIUM")
	if !almostEqual(score, 0.7, 0.01) {
		t.Errorf("expected 0.7 for MEDIUM, got %.2f", score)
	}
}

func TestSourceConfidenceLOW(t *testing.T) {
	score := SourceConfidence("LOW")
	if !almostEqual(score, 0.5, 0.01) {
		t.Errorf("expected 0.5 for LOW, got %.2f", score)
	}
}

func TestSourceConfidenceUnknown(t *testing.T) {
	score := SourceConfidence("UNKNOWN")
	if !almostEqual(score, 0.3, 0.01) {
		t.Errorf("expected 0.3 for UNKNOWN, got %.2f", score)
	}
}

func TestAggregateConfidenceSingleSource(t *testing.T) {
	sources := []SourceResult{
		{SourceLevel: "HIGH", Weight: 1.0, MatchRate: 1.0},
	}
	score := AggregateConfidence(sources)
	if !almostEqual(score, 0.9, 0.01) {
		t.Errorf("expected 0.9, got %.2f", score)
	}
}

func TestAggregateConfidenceMultipleHIGH(t *testing.T) {
	sources := []SourceResult{
		{SourceLevel: "HIGH", Weight: 1.0, MatchRate: 1.0},
		{SourceLevel: "HIGH", Weight: 0.8, MatchRate: 0.95},
	}
	score := AggregateConfidence(sources)
	if score < 0.8 {
		t.Errorf("expected high confidence for multiple HIGH sources, got %.2f", score)
	}
	// Score should be between the individual confidences
	if score > 0.95 {
		t.Errorf("expected reasonable aggregate, got %.2f", score)
	}
}

func TestAggregateConfidenceMixedSources(t *testing.T) {
	sources := []SourceResult{
		{SourceLevel: "HIGH", Weight: 1.0, MatchRate: 1.0},
		{SourceLevel: "LOW", Weight: 0.5, MatchRate: 0.6},
	}
	score := AggregateConfidence(sources)
	if score < 0.3 || score > 0.95 {
		t.Errorf("expected moderate score for mixed sources, got %.2f", score)
	}
}

func TestAggregateConfidenceEmpty(t *testing.T) {
	score := AggregateConfidence([]SourceResult{})
	if !almostEqual(score, 0, 0.01) {
		t.Errorf("expected 0 for no sources, got %.2f", score)
	}
}

func TestDecideStatusVerified(t *testing.T) {
	status := DecideStatus(0.95)
	if status != "verified" {
		t.Errorf("expected 'verified', got '%s'", status)
	}
}

func TestDecideStatusPendingReview(t *testing.T) {
	status := DecideStatus(0.80)
	if status != "pending_review" {
		t.Errorf("expected 'pending_review', got '%s'", status)
	}
}

func TestDecideStatusUnverified(t *testing.T) {
	status := DecideStatus(0.50)
	if status != "unverified" {
		t.Errorf("expected 'unverified', got '%s'", status)
	}
}

func TestDecideStatusBoundaryVerified(t *testing.T) {
	status := DecideStatus(0.90)
	if status != "verified" {
		t.Errorf("expected 'verified' at 0.90, got '%s'", status)
	}
}

func TestDecideStatusBoundaryPending(t *testing.T) {
	status := DecideStatus(0.70)
	if status != "pending_review" {
		t.Errorf("expected 'pending_review' at 0.70, got '%s'", status)
	}
}

func TestFieldLevelDiffExactMatch(t *testing.T) {
	diff := FieldLevelDiff("1000", "1000")
	if diff != 0 {
		t.Errorf("expected 0 diff for exact match, got %.2f", diff)
	}
}

func TestFieldLevelDiffNumericClose(t *testing.T) {
	diff := FieldLevelDiff("1000", "1050")
	if !almostEqual(diff, 0.05, 0.01) {
		t.Errorf("expected ~0.05 diff for 1000 vs 1050, got %.2f", diff)
	}
}

func TestFieldLevelDiffNumericHalf(t *testing.T) {
	diff := FieldLevelDiff("1000", "500")
	if !almostEqual(diff, 0.5, 0.01) {
		t.Errorf("expected ~0.5 diff for 1000 vs 500, got %.2f", diff)
	}
}

func TestFieldLevelDiffNumericTenX(t *testing.T) {
	diff := FieldLevelDiff("1000", "100")
	if !almostEqual(diff, 0.9, 0.01) {
		t.Errorf("expected ~0.9 diff for 1000 vs 100, got %.2f", diff)
	}
}

func TestFieldLevelDiffNonNumericSame(t *testing.T) {
	diff := FieldLevelDiff("pension", "pension")
	if diff != 0 {
		t.Errorf("expected 0 for identical strings, got %.2f", diff)
	}
}

func TestFieldLevelDiffNonNumericDifferent(t *testing.T) {
	diff := FieldLevelDiff("pension", "medical")
	if !almostEqual(diff, 1.0, 0.01) {
		t.Errorf("expected 1.0 for different strings, got %.2f", diff)
	}
}

func TestFieldLevelDiffEmpty(t *testing.T) {
	diff := FieldLevelDiff("", "")
	if !almostEqual(diff, 0, 0.01) {
		t.Errorf("expected 0 for both empty, got %.2f", diff)
	}
}

func TestFieldLevelDiffOneEmpty(t *testing.T) {
	diff := FieldLevelDiff("1000", "")
	if !almostEqual(diff, 1.0, 0.01) {
		t.Errorf("expected 1.0 when one is empty, got %.2f", diff)
	}
}

func almostEqual(a, b, eps float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < eps
}
