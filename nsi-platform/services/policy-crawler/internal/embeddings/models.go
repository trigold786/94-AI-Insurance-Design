package embeddings

type EmbeddedClaim struct {
	ClaimID    string
	PolicyID   string
	PolicyType string
	RegionCode string
	Embedding  []float64
	SourceName string
	PolicyURL  string
	Status     string
}

type SimilarResult struct {
	ClaimID    string  `json:"claim_id"`
	PolicyID   string  `json:"policy_id"`
	PolicyType string  `json:"policy_type"`
	RegionCode string  `json:"region_code"`
	Score      float64 `json:"score"`
	SourceName string  `json:"source_name"`
	PolicyURL  string  `json:"policy_url,omitempty"`
	Status     string  `json:"status"`
}

type SearchFilter struct {
	RegionCode string
	PolicyType string
}
