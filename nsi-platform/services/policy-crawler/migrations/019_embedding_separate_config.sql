-- 019: 分离 Embedding 配置（支持单独 API Key 和 Endpoint）
-- 火山方舟 Doubao Embedding 使用的 Endpoint 与 LLM 不同

ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS embedding_api_key TEXT NOT NULL DEFAULT '';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS embedding_endpoint TEXT NOT NULL DEFAULT 'https://ark.cn-beijing.volces.com/api/v3/embeddings';

-- 默认迁移：将现有 LLM API Key 复制到 embedding 字段，设置火山方舟 Doubao 为默认 Embedding 服务
UPDATE llm_configs
SET embedding_api_key = api_key,
    embedding_endpoint = 'https://ark.cn-beijing.volces.com/api/v3/embeddings',
    embedding_model = COALESCE(NULLIF(embedding_model,''), 'doubao-embedding-vision'),
    embedding_dimensions = COALESCE(NULLIF(embedding_dimensions,0), 1024)
WHERE embedding_api_key = '';
