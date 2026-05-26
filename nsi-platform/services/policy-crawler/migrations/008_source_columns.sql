ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS source_id text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS source_name text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS source_url text;
