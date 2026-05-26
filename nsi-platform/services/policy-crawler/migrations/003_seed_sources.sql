-- 003_seed_sources: add crawl config columns + seed data (PRD §4.1.1)
-- depends on: 001_init

BEGIN;

ALTER TABLE policy_sources ADD COLUMN IF NOT EXISTS crawl_type TEXT NOT NULL DEFAULT 'govsite';
ALTER TABLE policy_sources ADD COLUMN IF NOT EXISTS interval_sec INT NOT NULL DEFAULT 86400;
ALTER TABLE policy_sources ADD COLUMN IF NOT EXISTS region_code TEXT NOT NULL DEFAULT '';

INSERT INTO policy_sources (source_id, source_name, source_url, source_level, weight, enabled, crawl_type, interval_sec, region_code) VALUES
('SH-GOV-HRSS', '上海市人社局官网', 'https://rsj.sh.gov.cn', 'HIGH', 1.0, true, 'govsite', 86400, '310000'),
('BJ-GOV-HRSS', '北京市人社局官网', 'http://rsj.beijing.gov.cn', 'HIGH', 1.0, true, 'govsite', 86400, '110000'),
('SZ-GOV-HRSS', '深圳市人社局官网', 'http://hrss.sz.gov.cn', 'HIGH', 1.0, true, 'govsite', 86400, '440300'),
('GZ-GOV-HRSS', '广州市人社局官网', 'http://hrss.gz.gov.cn', 'HIGH', 1.0, true, 'govsite', 86400, '440100'),
('HZ-GOV-HRSS', '杭州市人社局官网', 'http://hrss.hangzhou.gov.cn', 'HIGH', 1.0, true, 'govsite', 86400, '330100'),
('CN-12333', '国家社保服务平台', 'http://si.12333.gov.cn', 'HIGH', 1.0, true, 'govsite', 86400, ''),
('LOCAL-FILE', '本地政策文件导入', '/data/policies', 'HIGH', 1.0, true, 'file', 60, '')
ON CONFLICT (source_id) DO NOTHING;

COMMIT;
