-- Add Jiangsu province and all 13 prefecture-level city policy sources

INSERT INTO policy_sources (source_id, source_name, source_url, source_level, weight, enabled, crawl_type, interval_sec, region_code, request_delay_ms, max_concurrent, respect_robots) VALUES
('JS-GOV-HRSS', '江苏省人社厅官网', 'https://jshrss.jiangsu.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '320000', 1000, 1, TRUE),
('NJ-GOV-HRSS', '南京市人社局官网', 'https://rsj.nanjing.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '320100', 1000, 1, TRUE),
('WX-GOV-HRSS', '无锡市人社局官网', 'https://hrss.wuxi.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '320200', 1000, 1, TRUE),
('XZ-GOV-HRSS', '徐州市人社局官网', 'https://hrss.xz.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '320300', 1000, 1, FALSE),
('CZ-GOV-HRSS', '常州市人社局官网', 'https://rsj.changzhou.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '320400', 1000, 1, TRUE),
('SU-GOV-HRSS', '苏州市人社局官网', 'https://hrss.suzhou.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '320500', 1000, 1, TRUE),
('NT-GOV-HRSS', '南通市人社局官网', 'https://rsj.nantong.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '320600', 1000, 1, TRUE),
('LYG-GOV-HRSS', '连云港市人社局官网', 'http://rsj.lyg.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '320700', 1000, 1, TRUE),
('HA-GOV-HRSS', '淮安市人社局官网', 'https://rsj.huaian.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '320800', 1000, 1, TRUE),
('YC-GOV-HRSS', '盐城市人社局官网', 'https://jsychrss.yancheng.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '320900', 1000, 1, TRUE),
('YZ-GOV-HRSS', '扬州市人社局官网', 'https://hrss.yangzhou.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '321000', 1000, 1, TRUE),
('ZJ-GOV-HRSS', '镇江市人社局官网', 'https://hrss.zhenjiang.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '321100', 1000, 1, TRUE),
('TZ-GOV-HRSS', '泰州市人社局官网', 'https://rsj.taizhou.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '321200', 1000, 1, TRUE),
('SQ-GOV-HRSS', '宿迁市人社局官网', 'https://sqhrss.suqian.gov.cn/', 'HIGH', 1.0, TRUE, 'govsite', 86400, '321300', 1000, 1, TRUE)
ON CONFLICT (source_id) DO NOTHING;
