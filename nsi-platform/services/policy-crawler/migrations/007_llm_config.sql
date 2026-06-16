-- 007_llm_config: LLM provider config + extract tracking
-- depends on: 004_crawl_tables

BEGIN;

CREATE TABLE IF NOT EXISTS llm_configs (
    id         BIGSERIAL PRIMARY KEY,
    provider   TEXT NOT NULL DEFAULT 'deepseek',
    api_key    TEXT NOT NULL DEFAULT '',
    endpoint   TEXT NOT NULL DEFAULT 'https://api.deepseek.com/v1/chat/completions',
    model_name TEXT NOT NULL DEFAULT 'deepseek-v4-flash',
    max_tokens INT NOT NULL DEFAULT 4096,
    enabled    BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO llm_configs (provider, api_key, endpoint, model_name, enabled)
SELECT 'deepseek', '', 'https://api.deepseek.com/v1/chat/completions', 'deepseek-v4-flash', false
WHERE NOT EXISTS (SELECT 1 FROM llm_configs);

ALTER TABLE policy_raw_texts ADD COLUMN IF NOT EXISTS extracted BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE policy_raw_texts ADD COLUMN IF NOT EXISTS extracted_claim_id TEXT;

CREATE INDEX IF NOT EXISTS idx_raw_texts_extracted ON policy_raw_texts(extracted) WHERE NOT extracted;

CREATE TABLE IF NOT EXISTS extract_logs (
    id         BIGSERIAL PRIMARY KEY,
    source_id  TEXT NOT NULL,
    status     TEXT NOT NULL CHECK (status IN ('success', 'failed')),
    message    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
