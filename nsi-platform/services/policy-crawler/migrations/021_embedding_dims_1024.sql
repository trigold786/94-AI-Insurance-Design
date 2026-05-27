-- 021: 修改向量维度为 1024（火山方舟 Doubao Embedding 默认维度）
-- 旧的 1536 维向量来自 OpenAI text-embedding-3-small 或 hash 回退

UPDATE policy_claims SET embedding = NULL WHERE embedding IS NOT NULL;

DROP INDEX IF EXISTS idx_policy_claims_embedding;

ALTER TABLE policy_claims ALTER COLUMN embedding TYPE vector(1024);

CREATE INDEX idx_policy_claims_embedding
    ON policy_claims
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

ALTER TABLE llm_configs ALTER COLUMN embedding_dimensions SET DEFAULT 1024;
