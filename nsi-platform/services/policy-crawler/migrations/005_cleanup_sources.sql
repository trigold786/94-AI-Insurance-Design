-- 005_cleanup_sources: disable duplicate old sources, keep new configured ones
-- depends on: 003_seed_sources, 004_crawl_tables

BEGIN;

-- 禁用旧的重复源（保留新添加的配置化源）
UPDATE policy_sources SET enabled = false WHERE source_id IN (
  'SRC-SH-GOV', 'SRC-BJ-GOV', 'SRC-SZ-GOV', 'SRC-GZ-GOV', 'SRC-HZ-GOV',
  'SRC-NHSA', 'SRC-MOHRSS'
);

-- 为新源补充 region_code（如某些源尚未设置）
UPDATE policy_sources SET region_code = '310000' WHERE source_id = 'SH-GOV-HRSS' AND region_code = '';
UPDATE policy_sources SET region_code = '110000' WHERE source_id = 'BJ-GOV-HRSS' AND region_code = '';
UPDATE policy_sources SET region_code = '440300' WHERE source_id = 'SZ-GOV-HRSS' AND region_code = '';
UPDATE policy_sources SET region_code = '440100' WHERE source_id = 'GZ-GOV-HRSS' AND region_code = '';
UPDATE policy_sources SET region_code = '330100' WHERE source_id = 'HZ-GOV-HRSS' AND region_code = '';

COMMIT;
