INSERT INTO policy_sources (source_id, source_name, source_url, source_level, weight, enabled, crawl_type, interval_sec, region_code) VALUES
('BILIBILI-SB', 'B站-社保政策视频', E'search:社保政策\nsearch:养老保险\nsearch:医疗保险\nsearch:失业保险\nsearch:工伤保险\nsearch:生育保险', 'LOW', 0.3, true, 'bilibili', 604800, '')
ON CONFLICT (source_id) DO NOTHING;
