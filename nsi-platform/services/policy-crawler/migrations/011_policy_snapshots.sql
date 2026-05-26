CREATE TABLE IF NOT EXISTS policy_snapshots (
    id SERIAL PRIMARY KEY,
    claim_id TEXT NOT NULL,
    policy_id TEXT NOT NULL,
    version_number INT NOT NULL,
    snapshot_data JSONB NOT NULL,
    superseded_by TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_snapshots_policy_id ON policy_snapshots(policy_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_claim_id ON policy_snapshots(claim_id);
