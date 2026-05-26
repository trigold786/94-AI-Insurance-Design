package embeddings

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type VectorSearcher struct {
	db       *sql.DB
	provider EmbeddingProvider
}

func NewVectorSearcher(db *sql.DB, provider EmbeddingProvider) *VectorSearcher {
	return &VectorSearcher{db: db, provider: provider}
}

func (s *VectorSearcher) GetEmbedding(claimID string) []float64 {
	if s.db == nil {
		return nil
	}
	var embStr string
	err := s.db.QueryRow(`SELECT embedding::text FROM policy_claims WHERE claim_id = $1`, claimID).Scan(&embStr)
	if err != nil {
		log.Printf("[pgvector] GetEmbedding error for %s: %v", claimID, err)
		return nil
	}
	return parseVectorText(embStr)
}

func (s *VectorSearcher) SearchSimilar(emb []float64, threshold float64, limit int, filter *SearchFilter) []SimilarResult {
	if s.db == nil || len(emb) == 0 {
		return nil
	}

	embStr := formatVector(emb)
	regionCode := ""
	policyType := ""
	if filter != nil {
		regionCode = filter.RegionCode
		policyType = filter.PolicyType
	}

	query := `SELECT claim_id, policy_id, policy_type, region_code,
	       COALESCE(source_name, ''), COALESCE(policy_url, ''), COALESCE(status, 'pending_review'),
	       1 - (embedding <=> $1::vector) AS score
	FROM policy_claims
	WHERE embedding IS NOT NULL
	  AND ($2::text = '' OR region_code = $2)
	  AND ($3::text = '' OR policy_type = $3)
	  AND 1 - (embedding <=> $1::vector) >= $4
	ORDER BY embedding <=> $1::vector
	LIMIT $5`

	rows, err := s.db.Query(query, embStr, regionCode, policyType, threshold, limit)
	if err != nil {
		log.Printf("[pgvector] SearchSimilar error: %v", err)
		return nil
	}
	defer rows.Close()

	return scanSimilarResults(rows)
}

func (s *VectorSearcher) SearchByText(ctx context.Context, queryText string, threshold float64, limit int, filter *SearchFilter) ([]SimilarResult, error) {
	vecs, err := s.provider.Embed(ctx, []string{queryText})
	if err != nil {
		results := s.KeywordSearch(queryText, limit, filter)
		return results, nil
	}
	if len(vecs) == 0 {
		results := s.KeywordSearch(queryText, limit, filter)
		return results, nil
	}
	results := s.SearchSimilar(vecs[0], threshold, limit, filter)
	return results, nil
}

func (s *VectorSearcher) KeywordSearch(query string, limit int, filter *SearchFilter) []SimilarResult {
	if s.db == nil || query == "" {
		return nil
	}

	regionCode := ""
	policyType := ""
	if filter != nil {
		regionCode = filter.RegionCode
		policyType = filter.PolicyType
	}

	q := `SELECT claim_id, policy_id, policy_type, region_code,
	       COALESCE(source_name, ''), COALESCE(policy_url, ''), COALESCE(status, 'pending_review'),
	       confidence_score AS score
	FROM policy_claims
	WHERE (policy_id ILIKE '%' || $1 || '%' OR subsidy_calc_method ILIKE '%' || $1 || '%')
	  AND ($2::text = '' OR region_code = $2)
	  AND ($3::text = '' OR policy_type = $3)
	ORDER BY confidence_score DESC
	LIMIT $4`

	rows, err := s.db.Query(q, query, regionCode, policyType, limit)
	if err != nil {
		log.Printf("[pgvector] KeywordSearch error: %v", err)
		return nil
	}
	defer rows.Close()

	return scanSimilarResults(rows)
}

func scanSimilarResults(rows *sql.Rows) []SimilarResult {
	var results []SimilarResult
	for rows.Next() {
		var r SimilarResult
		if err := rows.Scan(&r.ClaimID, &r.PolicyID, &r.PolicyType, &r.RegionCode,
			&r.SourceName, &r.PolicyURL, &r.Status, &r.Score); err != nil {
			break
		}
		results = append(results, r)
	}
	return results
}

func formatVector(v []float64) string {
	parts := make([]string, len(v))
	for i, x := range v {
		parts[i] = fmt.Sprintf("%g", x)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func parseVectorText(s string) []float64 {
	s = strings.Trim(s, "[]")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]float64, len(parts))
	for i, p := range parts {
		fmt.Sscanf(strings.TrimSpace(p), "%f", &result[i])
	}
	return result
}
