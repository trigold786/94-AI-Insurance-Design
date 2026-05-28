BEGIN;

CREATE TABLE IF NOT EXISTS model_configs (
  id SERIAL PRIMARY KEY,
  function_key VARCHAR(50) NOT NULL UNIQUE,
  provider VARCHAR(50) NOT NULL,
  model_id VARCHAR(100) NOT NULL,
  api_endpoint TEXT NOT NULL,
  api_key TEXT NOT NULL DEFAULT '',
  extra_params JSONB DEFAULT '{}',
  max_tokens INT DEFAULT 4096,
  enabled BOOLEAN DEFAULT false,
  backup_provider VARCHAR(50) DEFAULT '',
  backup_model_id VARCHAR(100) DEFAULT '',
  backup_api_endpoint TEXT DEFAULT '',
  backup_api_key TEXT DEFAULT '',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO model_configs (function_key, provider, model_id, api_endpoint, enabled) VALUES
  ('llm_extract', 'deepseek', 'deepseek-chat', 'https://api.deepseek.com/v1/chat/completions', false),
  ('llm_plan', 'deepseek', 'deepseek-chat', 'https://api.deepseek.com/v1/chat/completions', false),
  ('embedding', 'volc_ark', 'doubao-embedding-vision', 'https://ark.cn-beijing.volces.com/api/v3/embeddings/multimodal', false),
  ('asr', 'volcengine_asr', 'volc.bigasr.auc', 'https://openspeech.bytedance.com/api/v3/auc/bigmodel', false)
ON CONFLICT (function_key) DO NOTHING;

COMMIT;
