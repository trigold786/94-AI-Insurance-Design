-- 024_asr_llm_backup.sql
CREATE TABLE asr_configs (
    id SERIAL PRIMARY KEY,
    provider TEXT NOT NULL DEFAULT 'volcengine',
    api_key TEXT NOT NULL DEFAULT '',
    endpoint TEXT NOT NULL DEFAULT '',
    language TEXT NOT NULL DEFAULT 'zh',
    sample_rate INT NOT NULL DEFAULT 16000,
    enabled BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ DEFAULT now()
);
INSERT INTO asr_configs (provider, api_key, endpoint, enabled)
VALUES ('volcengine', '', 'https://openspeech.bytedance.com/api/v1/auc', false);

ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS backup_provider TEXT DEFAULT '';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS backup_api_key TEXT DEFAULT '';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS backup_endpoint TEXT DEFAULT '';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS backup_model_name TEXT DEFAULT '';
