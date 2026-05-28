package config

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ProviderConfig struct {
	ID           int       `json:"id"`
	ProviderName string   `json:"provider_name"`
	APIKey       string   `json:"api_key"`
	Endpoint     string   `json:"endpoint"`
	ModelName    string    `json:"model_name"`
	MaxTokens    int       `json:"max_tokens"`
	IsPrimary    bool      `json:"is_primary"`
	IsEnabled    bool      `json:"is_enabled"`
	Priority     int       `json:"priority"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ConfigStore struct {
	db *sql.DB
}

func NewConfigStore(db *sql.DB) (*ConfigStore, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &ConfigStore{db: db}, nil
}

func (s *ConfigStore) ListProviders(ctx context.Context) ([]ProviderConfig, error) {
	query := `SELECT id, provider_name, api_key, endpoint, model_name, max_tokens,
		is_primary, is_enabled, priority, created_at, updated_at
		FROM llm_providers ORDER BY priority ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query providers: %w", err)
	}
	defer rows.Close()

	var providers []ProviderConfig
	for rows.Next() {
		var p ProviderConfig
		if err := rows.Scan(&p.ID, &p.ProviderName, &p.APIKey, &p.Endpoint,
			&p.ModelName, &p.MaxTokens, &p.IsPrimary, &p.IsEnabled,
			&p.Priority, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan provider: %w", err)
		}
		providers = append(providers, p)
	}
	if providers == nil {
		providers = []ProviderConfig{}
	}
	return providers, nil
}

func (s *ConfigStore) GetEnabledProviders(ctx context.Context) ([]ProviderConfig, error) {
	query := `SELECT id, provider_name, api_key, endpoint, model_name, max_tokens,
		is_primary, is_enabled, priority, created_at, updated_at
		FROM llm_providers WHERE is_enabled = true ORDER BY priority ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query enabled providers: %w", err)
	}
	defer rows.Close()

	var providers []ProviderConfig
	for rows.Next() {
		var p ProviderConfig
		if err := rows.Scan(&p.ID, &p.ProviderName, &p.APIKey, &p.Endpoint,
			&p.ModelName, &p.MaxTokens, &p.IsPrimary, &p.IsEnabled,
			&p.Priority, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan provider: %w", err)
		}
		providers = append(providers, p)
	}
	if providers == nil {
		providers = []ProviderConfig{}
	}
	return providers, nil
}

func (s *ConfigStore) SaveProvider(ctx context.Context, p *ProviderConfig) error {
	query := `INSERT INTO llm_providers (provider_name, api_key, endpoint, model_name, max_tokens, is_primary, is_enabled, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (provider_name) DO UPDATE SET
			api_key = EXCLUDED.api_key,
			endpoint = EXCLUDED.endpoint,
			model_name = EXCLUDED.model_name,
			max_tokens = EXCLUDED.max_tokens,
			is_primary = EXCLUDED.is_primary,
			is_enabled = EXCLUDED.is_enabled,
			priority = EXCLUDED.priority,
			updated_at = NOW()`

	_, err := s.db.ExecContext(ctx, query,
		p.ProviderName, p.APIKey, p.Endpoint, p.ModelName,
		p.MaxTokens, p.IsPrimary, p.IsEnabled, p.Priority)
	if err != nil {
		return fmt.Errorf("save provider: %w", err)
	}
	return nil
}
