package embeddings

import (
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

const Dim = 256

func tokenize(text string) []string {
	var tokens []string
	var buf strings.Builder
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
		} else {
			if buf.Len() > 0 {
				tokens = append(tokens, buf.String())
				buf.Reset()
			}
		}
	}
	if buf.Len() > 0 {
		tokens = append(tokens, buf.String())
	}
	return tokens
}

func hashPos(token string, dim int) int {
	h := fnv.New64a()
	h.Write([]byte(token))
	return int(h.Sum64() % uint64(dim))
}

func Normalize(v []float64) {
	var sum float64
	for _, x := range v {
		sum += x * x
	}
	norm := float64(math.Sqrt(sum))
	if norm > 0 {
		for i := range v {
			v[i] /= norm
		}
	}
}

func FromText(text string) []float64 {
	vec := make([]float64, Dim)
	tokens := tokenize(text)
	for _, t := range tokens {
		pos := hashPos(t, Dim)
		vec[pos]++
	}
	Normalize(vec)
	return vec
}
