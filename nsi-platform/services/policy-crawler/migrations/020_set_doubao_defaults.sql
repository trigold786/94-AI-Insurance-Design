-- 020: 设置火山方舟 Doubao Embedding 为默认模型和维度
-- 旧默认值来自 012: text-embedding-3-small / 1536
-- 旧 Endpoint 来自 019 迁移时从 DeepSeek LLM 配置衍生（api.deepseek.com/v1/embeddings 不存在）

UPDATE llm_configs
SET embedding_model = 'doubao-embedding-vision',
    embedding_dimensions = 1024,
    embedding_endpoint = 'https://ark.cn-beijing.volces.com/api/v3/embeddings'
WHERE embedding_model = 'text-embedding-3-small'
   OR embedding_dimensions = 1536
   OR embedding_endpoint LIKE '%deepseek%'
   OR embedding_endpoint LIKE '%openai%';

ALTER TABLE llm_configs ALTER COLUMN embedding_model SET DEFAULT 'doubao-embedding-vision';
ALTER TABLE llm_configs ALTER COLUMN embedding_dimensions SET DEFAULT 1024;
