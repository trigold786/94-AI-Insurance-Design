package embeddings

import (
	"math"
	"testing"
)

func TestCosineSimilarity_Identical(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{1, 0, 0}
	s := CosineSimilarity(a, b)
	if s != 1.0 {
		t.Fatalf("identical vectors should have similarity 1.0, got %f", s)
	}
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{0, 1, 0}
	s := CosineSimilarity(a, b)
	if s != 0.0 {
		t.Fatalf("orthogonal vectors should have similarity 0.0, got %f", s)
	}
}

func TestCosineSimilarity_Partial(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{0.5, 0.5, 0}
	s := CosineSimilarity(a, b)
	expected := 1.0 / math.Sqrt(2)
	if math.Abs(s-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, s)
	}
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	a := []float64{0, 0, 0}
	b := []float64{1, 0, 0}
	s := CosineSimilarity(a, b)
	if s != 0.0 {
		t.Fatalf("zero vector should return 0.0, got %f", s)
	}
}

func TestCosineSimilarity_BothZero(t *testing.T) {
	a := []float64{0, 0, 0}
	b := []float64{0, 0, 0}
	s := CosineSimilarity(a, b)
	if s != 0.0 {
		t.Fatalf("both zero should return 0.0, got %f", s)
	}
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should panic on different lengths")
		}
	}()
	CosineSimilarity([]float64{1, 0}, []float64{1, 0, 0})
}
