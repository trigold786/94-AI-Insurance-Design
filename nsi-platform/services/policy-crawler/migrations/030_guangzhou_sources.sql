-- 广州(Guangzhou)政策数据源补强
-- 当前广州仅7条政策，需大幅增加覆盖
INSERT INTO policy_sources (source_id, source_name, source_url, source_level, weight, enabled, crawl_type, interval_sec, region_code) VALUES
('GZ-GOV-HRSS', '广州市人社局', 'http://hrss.gz.gov.cn/zwgk/zcfg/', 'HIGH', 1.0, true, 'govsite', 86400, '440100'),
('GZ-GOV-HRSS-TZGG', '广州市人社局-通知公告', 'http://hrss.gz.gov.cn/zwgk/tzgg/', 'HIGH', 1.0, true, 'govsite', 86400, '440100'),
('GZ-GOV-HRSS-LHJY', '广州市人社局-灵活就业', 'http://hrss.gz.gov.cn/ztzl/lhjy/', 'HIGH', 1.0, true, 'govsite', 86400, '440100'),
('GZ-GOV-HRSS-SBZX', '广州市人社局-社保专题', 'http://hrss.gz.gov.cn/sbzx/', 'HIGH', 1.0, true, 'govsite', 86400, '440100'),
('GZ-GOV-HRSS-YL', '广州市人社局-医保', 'http://hrss.gz.gov.cn/ylbx/', 'HIGH', 0.9, true, 'govsite', 86400, '440100'),
('GZ-GOV-GENERAL', '广州市人民政府-社保政策', 'https://www.gz.gov.cn/zwgk/zfxxgkml/bmfg/list_15.shtml', 'HIGH', 0.9, true, 'govsite', 86400, '440100')
ON CONFLICT (source_id) DO NOTHING;
