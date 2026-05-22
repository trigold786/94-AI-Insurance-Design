-- 001_init: core tables for policy-crawler
-- depends on: none

BEGIN;

-- policy_sources: data source registry (SSD §5.2 - SourcePlugin architecture)
CREATE TABLE IF NOT EXISTS policy_sources (
    source_id    TEXT PRIMARY KEY,
    source_name  TEXT NOT NULL,
    source_url   TEXT NOT NULL,
    source_level TEXT NOT NULL CHECK (source_level IN ('HIGH', 'MEDIUM', 'LOW')),
    weight       DOUBLE PRECISION NOT NULL DEFAULT 1.0 CHECK (weight > 0 AND weight <= 1.0),
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- policy_claims: structured policy claims (SSD §5.2 - Policy Verification Pipeline)
CREATE TABLE IF NOT EXISTS policy_claims (
    claim_id             TEXT PRIMARY KEY,
    policy_id            TEXT NOT NULL,
    region_code          TEXT NOT NULL,
    policy_type          TEXT NOT NULL CHECK (policy_type IN ('pension', 'medical', 'unemployment', 'injury', 'maternity', 'housing_fund', 'subsidy', 'training')),
    target_group_tags    TEXT[] NOT NULL DEFAULT '{}',
    subsidy_calc_method  TEXT NOT NULL,
    subsidy_amount_min   DOUBLE PRECISION,
    subsidy_amount_max   DOUBLE PRECISION,
    subsidy_duration     INT,
    effective_date       DATE NOT NULL,
    expire_date          DATE,
    confidence_score     DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (confidence_score >= 0 AND confidence_score <= 1),
    status               TEXT NOT NULL DEFAULT 'pending_review' CHECK (status IN ('verified', 'pending_review', 'unverified')),
    version_number       INT NOT NULL DEFAULT 1,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- fetch_history: audit log for crawler runs
CREATE TABLE IF NOT EXISTS fetch_history (
    id           BIGSERIAL PRIMARY KEY,
    source_id    TEXT NOT NULL REFERENCES policy_sources(source_id),
    result_count INT NOT NULL DEFAULT 0,
    success      BOOLEAN NOT NULL DEFAULT FALSE,
    error_msg    TEXT,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_policy_claims_region ON policy_claims(region_code);
CREATE INDEX IF NOT EXISTS idx_policy_claims_policy_id ON policy_claims(policy_id);
CREATE INDEX IF NOT EXISTS idx_policy_claims_status ON policy_claims(status);
CREATE INDEX IF NOT EXISTS idx_fetch_history_source ON fetch_history(source_id);

COMMIT;
