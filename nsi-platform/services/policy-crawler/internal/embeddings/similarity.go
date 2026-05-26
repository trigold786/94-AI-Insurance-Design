package embeddings

import "math"

func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("cosine similarity: vectors must have same length")
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
