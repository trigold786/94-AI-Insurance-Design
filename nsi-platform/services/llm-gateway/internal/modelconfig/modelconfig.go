package modelconfig

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type ModelConfig struct {
	ID                int             `json:"id"`
	FunctionKey       string          `json:"function_key"`
	Provider          string          `json:"provider"`
	ModelID           string          `json:"model_id"`
	APIEndpoint       string          `json:"api_endpoint"`
	APIKey            string          `json:"api_key"`
	ExtraParams       json.RawMessage `json:"extra_params"`
	MaxTokens         int             `json:"max_tokens"`
	Enabled           bool            `json:"enabled"`
	BackupProvider    string          `json:"backup_provider"`
	BackupModelID     string          `json:"backup_model_id"`
	BackupAPIEndpoint string          `json:"backup_api_endpoint"`
	BackupAPIKey      string          `json:"backup_api_key"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &Store{db: db}, nil
}

func (s *Store) ListAll(ctx context.Context) ([]ModelConfig, error) {
	query := `SELECT id, function_key, provider, model_id, api_endpoint, api_key,
		extra_params, max_tokens, enabled,
		backup_provider, backup_model_id, backup_api_endpoint, backup_api_key,
		created_at, updated_at
		FROM model_configs ORDER BY function_key`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query model_configs: %w", err)
	}
	defer rows.Close()

	var configs []ModelConfig
	for rows.Next() {
		var c ModelConfig
		if err := rows.Scan(&c.ID, &c.FunctionKey, &c.Provider, &c.ModelID,
			&c.APIEndpoint, &c.APIKey, &c.ExtraParams, &c.MaxTokens, &c.Enabled,
			&c.BackupProvider, &c.BackupModelID, &c.BackupAPIEndpoint, &c.BackupAPIKey,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan model_config: %w", err)
		}
		configs = append(configs, c)
	}
	if configs == nil {
		configs = []ModelConfig{}
	}
	return configs, nil
}

func (s *Store) GetByKey(ctx context.Context, functionKey string) (*ModelConfig, error) {
	query := `SELECT id, function_key, provider, model_id, api_endpoint, api_key,
		extra_params, max_tokens, enabled,
		backup_provider, backup_model_id, backup_api_endpoint, backup_api_key,
		created_at, updated_at
		FROM model_configs WHERE function_key = $1`

	var c ModelConfig
	err := s.db.QueryRowContext(ctx, query, functionKey).Scan(
		&c.ID, &c.FunctionKey, &c.Provider, &c.ModelID,
		&c.APIEndpoint, &c.APIKey, &c.ExtraParams, &c.MaxTokens, &c.Enabled,
		&c.BackupProvider, &c.BackupModelID, &c.BackupAPIEndpoint, &c.BackupAPIKey,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("model config not found: %s", functionKey)
	}
	if err != nil {
		return nil, fmt.Errorf("query model_config: %w", err)
	}
	return &c, nil
}

func (s *Store) Save(ctx context.Context, cfg *ModelConfig) error {
	query := `INSERT INTO model_configs (function_key, provider, model_id, api_endpoint, api_key,
		extra_params, max_tokens, enabled,
		backup_provider, backup_model_id, backup_api_endpoint, backup_api_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (function_key) DO UPDATE SET
			provider = EXCLUDED.provider,
			model_id = EXCLUDED.model_id,
			api_endpoint = EXCLUDED.api_endpoint,
			api_key = EXCLUDED.api_key,
			extra_params = EXCLUDED.extra_params,
			max_tokens = EXCLUDED.max_tokens,
			enabled = EXCLUDED.enabled,
			backup_provider = EXCLUDED.backup_provider,
			backup_model_id = EXCLUDED.backup_model_id,
			backup_api_endpoint = EXCLUDED.backup_api_endpoint,
			backup_api_key = EXCLUDED.backup_api_key,
			updated_at = NOW()`

	if cfg.ExtraParams == nil {
		cfg.ExtraParams = json.RawMessage(`{}`)
	}

	_, err := s.db.ExecContext(ctx, query,
		cfg.FunctionKey, cfg.Provider, cfg.ModelID, cfg.APIEndpoint, cfg.APIKey,
		cfg.ExtraParams, cfg.MaxTokens, cfg.Enabled,
		cfg.BackupProvider, cfg.BackupModelID, cfg.BackupAPIEndpoint, cfg.BackupAPIKey,
	)
	if err != nil {
		return fmt.Errorf("save model_config: %w", err)
	}
	return nil
}

func MaskKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func (c *ModelConfig) ToMasked() *ModelConfig {
	cp := *c
	cp.APIKey = MaskKey(c.APIKey)
	if c.BackupAPIKey != "" {
		cp.BackupAPIKey = MaskKey(c.BackupAPIKey)
	}
	return &cp
}
