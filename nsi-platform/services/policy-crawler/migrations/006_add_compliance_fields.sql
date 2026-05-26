-- 006_add_compliance_fields: add compliance columns to policy_claims (sync with api-server)
-- depends on: 001_init

BEGIN;

ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS conditions JSONB DEFAULT '[]';
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS required_documents JSONB DEFAULT '[]';

COMMIT;
