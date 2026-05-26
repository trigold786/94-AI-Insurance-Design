package embeddings

import "context"

type HashProvider struct{}

func (p *HashProvider) Embed(_ context.Context, texts []string) ([][]float64, error) {
	result := make([][]float64, len(texts))
	for i, text := range texts {
		raw := FromText(text)
		padded := make([]float64, 1536)
		copy(padded, raw)
		Normalize(padded)
		result[i] = padded
	}
	return result, nil
}

func (p *HashProvider) Dimensions() int { return 1536 }

func (p *HashProvider) ModelName() string { return "hash-bow" }
