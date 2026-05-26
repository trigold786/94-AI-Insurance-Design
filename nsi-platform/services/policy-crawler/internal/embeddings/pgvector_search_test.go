package embeddings

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestVectorSearcher_GetEmbedding_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	p := &HashProvider{}
	s := NewVectorSearcher(db, p)

	mock.ExpectQuery(`SELECT embedding::text FROM policy_claims WHERE claim_id = \$1`).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"embedding"}))

	result := s.GetEmbedding("nonexistent")
	if result != nil {
		t.Fatal("expected nil for nonexistent")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestVectorSearcher_SearchSimilar_WithFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	p := &HashProvider{}
	s := NewVectorSearcher(db, p)

	emb := make([]float64, 1536)
	emb[0] = 1.0

	rows := sqlmock.NewRows([]string{
		"claim_id", "policy_id", "policy_type", "region_code",
		"source_name", "policy_url", "status", "score",
	}).AddRow("c1", "p1", "subsidy", "110000", "gov", "", "verified", 0.9)

	mock.ExpectQuery(`SELECT claim_id, policy_id, policy_type, region_code`).
		WithArgs(sqlmock.AnyArg(), "110000", "subsidy", 0.5, 20).
		WillReturnRows(rows)

	results := s.SearchSimilar(emb, 0.5, 20, &SearchFilter{RegionCode: "110000", PolicyType: "subsidy"})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ClaimID != "c1" {
		t.Fatalf("expected c1, got %s", results[0].ClaimID)
	}
}

func TestVectorSearcher_SearchSimilar_EmptyDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	p := &HashProvider{}
	s := NewVectorSearcher(db, p)

	emb := make([]float64, 1536)

	mock.ExpectQuery(`SELECT claim_id, policy_id, policy_type, region_code`).
		WithArgs(sqlmock.AnyArg(), "", "", 0.0, 10).
		WillReturnRows(sqlmock.NewRows([]string{
			"claim_id", "policy_id", "policy_type", "region_code",
			"source_name", "policy_url", "status", "score",
		}))

	results := s.SearchSimilar(emb, 0.0, 10, nil)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestVectorSearcher_KeywordSearch_NilDB(t *testing.T) {
	p := &HashProvider{}
	s := NewVectorSearcher(nil, p)
	if s.KeywordSearch("test", 10, nil) != nil {
		t.Fatal("expected nil for nil db")
	}
}

func TestVectorSearcher_KeywordSearch_EmptyQuery(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	p := &HashProvider{}
	s := NewVectorSearcher(db, p)
	if s.KeywordSearch("", 10, nil) != nil {
		t.Fatal("expected nil for empty query")
	}
}

func TestVectorSearcher_SearchByText_UsesHashProvider(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	p := &HashProvider{}
	s := NewVectorSearcher(db, p)

	rows := sqlmock.NewRows([]string{
		"claim_id", "policy_id", "policy_type", "region_code",
		"source_name", "policy_url", "status", "score",
	}).AddRow("c1", "p1", "subsidy", "110000", "gov", "", "verified", 0.7)

	mock.ExpectQuery(`SELECT claim_id, policy_id, policy_type, region_code`).
		WithArgs(sqlmock.AnyArg(), "", "", 0.0, 10).
		WillReturnRows(rows)

	results, err := s.SearchByText(context.Background(), "补贴政策", 0.0, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestVectorSearcher_NilDB(t *testing.T) {
	p := &HashProvider{}
	s := NewVectorSearcher(nil, p)
	if s.GetEmbedding("c1") != nil {
		t.Fatal("expected nil for nil db")
	}
	if s.SearchSimilar([]float64{1}, 0, 10, nil) != nil {
		t.Fatal("expected nil for nil db")
	}
	if s.KeywordSearch("test", 10, nil) != nil {
		t.Fatal("expected nil for nil db")
	}
}

func TestFormatVector(t *testing.T) {
	v := []float64{1.0, 2.5, 0.0}
	s := formatVector(v)
	if s != "[1,2.5,0]" {
		t.Fatalf("unexpected: %s", s)
	}
}

func TestParseVectorText(t *testing.T) {
	result := parseVectorText("[1,2.5,0]")
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
	if result[0] != 1.0 || result[1] != 2.5 || result[2] != 0.0 {
		t.Fatalf("unexpected: %v", result)
	}
}
