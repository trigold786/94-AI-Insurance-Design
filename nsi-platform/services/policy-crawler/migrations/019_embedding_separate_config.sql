-- 019: 分离 Embedding 配置（支持单独 API Key 和 Endpoint）
-- 火山方舟 Doubao Embedding 使用的 Endpoint 与 LLM 不同

ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS embedding_api_key TEXT NOT NULL DEFAULT '';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS embedding_endpoint TEXT NOT NULL DEFAULT 'https://ark.cn-beijing.volces.com/api/v3/embeddings';

-- 默认迁移：将现有 LLM API Key 和 Endpoint 复制到 embedding 字段
UPDATE llm_configs
SET embedding_api_key = api_key,
    embedding_endpoint = CASE
        WHEN endpoint LIKE '%/chat/%' THEN regexp_replace(endpoint, '/chat/completions$', '/embeddings')
        ELSE 'https://ark.cn-beijing.volces.com/api/v3/embeddings'
    END
WHERE embedding_api_key = '';
