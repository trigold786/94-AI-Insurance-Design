package embeddings

import (
	"context"
	"net/http"
)

type EmbeddingProvider interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
	Dimensions() int
	ModelName() string
}

type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	model      string
	dimensions int
	client     HTTPDoer
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewProviderFromConfig(apiKey, baseURL, model string, dimensions int) EmbeddingProvider {
	if apiKey != "" {
		return &OpenAIProvider{
			apiKey:     apiKey,
			baseURL:    baseURL,
			model:      model,
			dimensions: dimensions,
		}
	}
	return &HashProvider{}
}
