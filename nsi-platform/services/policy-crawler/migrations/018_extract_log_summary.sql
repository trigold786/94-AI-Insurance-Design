-- 018: Add content_summary field to extract_logs for LLM extraction summary
BEGIN;
ALTER TABLE extract_logs ADD COLUMN IF NOT EXISTS content_summary TEXT;
COMMIT;
