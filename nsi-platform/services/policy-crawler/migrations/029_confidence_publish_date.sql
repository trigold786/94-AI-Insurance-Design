ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS publish_date TEXT DEFAULT '';

-- Confidence scoring support columns (Gap#2: persist verifier inputs)
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS match_rate DOUBLE PRECISION DEFAULT 0;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS conflict_score DOUBLE PRECISION DEFAULT 0;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS verified_by TEXT DEFAULT '';
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS source_level TEXT DEFAULT '';
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS fetched_at TEXT DEFAULT '';

-- Confidence weight config table (Gap#2: hot-reloadable weights)
CREATE TABLE IF NOT EXISTS confidence_config (
    id SERIAL PRIMARY KEY,
    w_source DOUBLE PRECISION DEFAULT 0.30,
    w_match DOUBLE PRECISION DEFAULT 0.20,
    w_conflict DOUBLE PRECISION DEFAULT 0.20,
    w_freshness DOUBLE PRECISION DEFAULT 0.15,
    w_expert DOUBLE PRECISION DEFAULT 0.15,
    verified_threshold DOUBLE PRECISION DEFAULT 0.85,
    pending_threshold DOUBLE PRECISION DEFAULT 0.60,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO confidence_config (id) VALUES (1) ON CONFLICT DO NOTHING;
