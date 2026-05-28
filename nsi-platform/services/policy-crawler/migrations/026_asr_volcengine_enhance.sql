-- 026_asr_volcengine_enhance.sql
ALTER TABLE asr_configs ADD COLUMN IF NOT EXISTS app_id TEXT DEFAULT '';
ALTER TABLE asr_configs ADD COLUMN IF NOT EXISTS resource_id TEXT DEFAULT 'volc.bigasr.auc';
ALTER TABLE asr_configs ADD COLUMN IF NOT EXISTS max_wait_seconds INT DEFAULT 300;
ALTER TABLE asr_configs ADD COLUMN IF NOT EXISTS poll_interval_seconds INT DEFAULT 5;
UPDATE asr_configs SET app_id = '', resource_id = 'volc.bigasr.auc', max_wait_seconds = 300, poll_interval_seconds = 5 WHERE id = (SELECT MIN(id) FROM asr_configs);
