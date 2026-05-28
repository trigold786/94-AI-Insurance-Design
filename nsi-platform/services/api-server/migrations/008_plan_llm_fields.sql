BEGIN;

ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS free_form_text TEXT DEFAULT '';
ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS structured_schemes JSONB DEFAULT '[]';
ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS policy_references JSONB DEFAULT '[]';
ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS recommendation TEXT DEFAULT '';
ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS recommendation_reason TEXT DEFAULT '';
ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS verification_result JSONB;

CREATE TABLE IF NOT EXISTS plan_verification_logs (
  id BIGSERIAL PRIMARY KEY,
  plan_id VARCHAR(100),
  llm_provider VARCHAR(50),
  llm_scheme_name VARCHAR(200),
  metric VARCHAR(50),
  llm_value DECIMAL(15,2),
  actuary_value DECIMAL(15,2),
  deviation_pct DECIMAL(5,2),
  root_cause VARCHAR(50),
  resolution TEXT,
  resolved BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_verification_logs_plan_id ON plan_verification_logs(plan_id);
CREATE INDEX IF NOT EXISTS idx_verification_logs_resolved ON plan_verification_logs(resolved);

COMMIT;
