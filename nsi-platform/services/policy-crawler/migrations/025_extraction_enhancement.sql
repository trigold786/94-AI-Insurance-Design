-- 025_extraction_enhancement.sql
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS policy_title text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS issuing_authority text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS document_number text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS application_process jsonb DEFAULT '[]'::jsonb;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS contact_info text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS source_type text;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS extraction_method text DEFAULT 'full';
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS raw_text_length int DEFAULT 0;
ALTER TABLE policy_claims ADD COLUMN IF NOT EXISTS split_count int DEFAULT 0;
