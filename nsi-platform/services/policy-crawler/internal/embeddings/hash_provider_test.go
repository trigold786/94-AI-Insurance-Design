package embeddings

import (
	"math"
	"testing"
)

func TestHashProvider_Dimensions(t *testing.T) {
	p := &HashProvider{}
	if p.Dimensions() != 1536 {
		t.Fatalf("expected 1536 dimensions, got %d", p.Dimensions())
	}
}

func TestHashProvider_ModelName(t *testing.T) {
	p := &HashProvider{}
	if p.ModelName() != "hash-bow" {
		t.Fatalf("expected hash-bow, got %s", p.ModelName())
	}
}

func TestHashProvider_Embed_Single(t *testing.T) {
	p := &HashProvider{}
	vecs, err := p.Embed(nil, []string{"养老保险 补贴"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 1 {
		t.Fatalf("expected 1 vector, got %d", len(vecs))
	}
	v := vecs[0]
	if len(v) != 1536 {
		t.Fatalf("expected 1536 dims, got %d", len(v))
	}
	nonZero := 0
	for i := 0; i < 256; i++ {
		if v[i] != 0 {
			nonZero++
		}
	}
	if nonZero == 0 {
		t.Fatal("first 256 dims should have non-zero values from hash")
	}
	for i := 256; i < 1536; i++ {
		if v[i] != 0 {
			t.Fatalf("dims 256-1535 should be zero, but dim %d = %f", i, v[i])
		}
	}
}

func TestHashProvider_Embed_Normalized(t *testing.T) {
	p := &HashProvider{}
	vecs, _ := p.Embed(nil, []string{"test input text here"})
	v := vecs[0]
	var norm float64
	for _, x := range v {
		norm += x * x
	}
	norm = math.Sqrt(norm)
	if math.Abs(norm-1.0) > 1e-10 {
		t.Fatalf("expected L2 norm 1.0, got %f", norm)
	}
}

func TestHashProvider_Embed_Multiple(t *testing.T) {
	p := &HashProvider{}
	vecs, err := p.Embed(nil, []string{"text one", "text two"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vecs))
	}
}
