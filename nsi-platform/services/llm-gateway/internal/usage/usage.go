package usage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type UsageLog struct {
	ID           int64     `json:"id"`
	ProviderName string    `json:"provider_name"`
	Model        string    `json:"model"`
	Caller       string    `json:"caller"`
	TokensIn     int       `json:"tokens_in"`
	TokensOut    int       `json:"tokens_out"`
	LatencyMs    int       `json:"latency_ms"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type UsageStore struct {
	db *sql.DB
}

func NewUsageStore(db *sql.DB) (*UsageStore, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &UsageStore{db: db}, nil
}

func (s *UsageStore) Record(ctx context.Context, log *UsageLog) error {
	query := `INSERT INTO llm_usage_logs (provider_name, model, caller, tokens_in, tokens_out, latency_ms, status, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := s.db.ExecContext(ctx, query,
		log.ProviderName, log.Model, log.Caller,
		log.TokensIn, log.TokensOut, log.LatencyMs,
		log.Status, log.ErrorMessage)
	if err != nil {
		return fmt.Errorf("record usage: %w", err)
	}
	return nil
}

type DailyUsage struct {
	Date          string  `json:"date"`
	ProviderName  string  `json:"provider_name"`
	TotalCalls    int     `json:"total_calls"`
	TotalTokensIn  int    `json:"total_tokens_in"`
	TotalTokensOut int    `json:"total_tokens_out"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
}

func (s *UsageStore) GetDailyUsage(ctx context.Context, days int) ([]DailyUsage, error) {
	query := `SELECT DATE(created_at) as date, provider_name,
		COUNT(*) as total_calls,
		COALESCE(SUM(tokens_in), 0) as total_tokens_in,
		COALESCE(SUM(tokens_out), 0) as total_tokens_out,
		COALESCE(AVG(latency_ms), 0) as avg_latency_ms
		FROM llm_usage_logs
		WHERE created_at >= NOW() - ($1 || ' days')::INTERVAL
		GROUP BY DATE(created_at), provider_name
		ORDER BY date DESC, provider_name`

	rows, err := s.db.QueryContext(ctx, query, fmt.Sprintf("%d", days))
	if err != nil {
		return nil, fmt.Errorf("query daily usage: %w", err)
	}
	defer rows.Close()

	var result []DailyUsage
	for rows.Next() {
		var u DailyUsage
		if err := rows.Scan(&u.Date, &u.ProviderName, &u.TotalCalls,
			&u.TotalTokensIn, &u.TotalTokensOut, &u.AvgLatencyMs); err != nil {
			return nil, fmt.Errorf("scan daily usage: %w", err)
		}
		result = append(result, u)
	}
	if result == nil {
		result = []DailyUsage{}
	}
	return result, nil
}
