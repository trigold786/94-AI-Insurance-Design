-- 004_add_policy_compliance: compliance fields + checklist tables
-- depends on: 002_seed_policies

BEGIN;

ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS conditions JSONB DEFAULT '[]';
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS required_documents JSONB DEFAULT '[]';

CREATE TABLE IF NOT EXISTS compliance_checklists (
    id              BIGSERIAL PRIMARY KEY,
    user_id         TEXT NOT NULL,
    city_code       TEXT NOT NULL,
    generated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    matched_policy_ids TEXT[] NOT NULL DEFAULT '{}',
    eligible_tags   TEXT[] NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_compliance_user ON compliance_checklists(user_id);
CREATE INDEX IF NOT EXISTS idx_compliance_city ON compliance_checklists(city_code);

COMMIT;
