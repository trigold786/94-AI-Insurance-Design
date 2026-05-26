INSERT INTO policy_sources (source_id, source_name, source_url, source_level, weight, enabled, crawl_type, interval_sec, region_code) VALUES
('DOUYIN-SB',    '抖音-社保视频',     '',     'LOW', 0.3, true, 'douyin', 604800, '')
ON CONFLICT (source_id) DO NOTHING;
