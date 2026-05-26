CREATE EXTENSION IF NOT EXISTS vector;

UPDATE policy_claims SET embedding = NULL WHERE embedding IS NOT NULL;

ALTER TABLE policy_claims
  ALTER COLUMN embedding TYPE vector(1536)
  USING NULL::vector;

CREATE INDEX idx_policy_claims_embedding
  ON policy_claims
  USING ivfflat (embedding vector_cosine_ops)
  WITH (lists = 100);

ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS embedding_model TEXT NOT NULL DEFAULT 'text-embedding-3-small';
ALTER TABLE llm_configs ADD COLUMN IF NOT EXISTS embedding_dimensions INT NOT NULL DEFAULT 1536;

-- NOTE: All existing embeddings are cleared (old 256-dim hash vectors incompatible with 1536-dim vector type).
-- After migration, run LLM extraction or use the admin re-embed endpoint to regenerate embeddings with the configured semantic model.
-- Claims without embeddings will fall back to keyword search via ILIKE matching.
