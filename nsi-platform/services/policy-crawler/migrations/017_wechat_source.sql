-- 017: WeChat official account source seeds
BEGIN;
INSERT INTO policy_sources (source_id, source_name, source_url, source_level, crawl_type, interval_sec, region_code, enabled)
VALUES ('WECHAT-上海社保', '微信公众号-上海社保', '上海社保', 'MEDIUM', 'wechat', 604800, '310000', false),
       ('WECHAT-社保搜索', '微信公众号-社保政策搜索', 'keyword:社保政策 补缴', 'LOW', 'wechat', 604800, '000000', false)
ON CONFLICT (source_id) DO NOTHING;
COMMIT;
