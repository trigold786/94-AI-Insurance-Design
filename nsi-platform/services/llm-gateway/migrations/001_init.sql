BEGIN;

CREATE TABLE IF NOT EXISTS llm_providers (
  id SERIAL PRIMARY KEY,
  provider_name VARCHAR(50) NOT NULL UNIQUE,
  api_key TEXT NOT NULL,
  endpoint TEXT NOT NULL,
  model_name VARCHAR(100) NOT NULL,
  max_tokens INT DEFAULT 4096,
  is_primary BOOLEAN DEFAULT false,
  is_enabled BOOLEAN DEFAULT true,
  priority INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS llm_usage_logs (
  id BIGSERIAL PRIMARY KEY,
  provider_name VARCHAR(50),
  model VARCHAR(100),
  caller VARCHAR(50),
  tokens_in INT,
  tokens_out INT,
  latency_ms INT,
  status VARCHAR(20),
  error_message TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_usage_logs_created_at ON llm_usage_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_usage_logs_provider ON llm_usage_logs(provider_name);

COMMIT;
