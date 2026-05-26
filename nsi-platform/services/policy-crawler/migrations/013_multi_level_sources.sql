INSERT INTO policy_sources (source_id, source_name, source_url, source_level, weight, enabled, crawl_type, interval_sec, region_code) VALUES
('MEDIA-RM-SS',  '人民日报-社会频道',   'http://www.people.com.cn/rss/society.xml',    'MEDIUM', 0.7, true, 'rss', 604800, ''),
('MEDIA-XH-SS',  '人民日报-时政频道',   'http://www.people.com.cn/rss/politics.xml',   'MEDIUM', 0.7, true, 'rss', 604800, ''),
('MEDIA-CE-SS',  '人民日报-财经频道',   'http://www.people.com.cn/rss/finance.xml',    'MEDIUM', 0.7, true, 'rss', 604800, ''),
('MEDIA-YL-ZX',  '人民日报-健康频道',   'http://www.people.com.cn/rss/health.xml',     'MEDIUM', 0.7, true, 'rss', 604800, ''),
('MEDIA-21JJ',   '21世纪经济报道-社保',  'https://www.21jingji.com/rss/',               'MEDIUM', 0.7, false, 'rss', 604800, ''),
('MANUAL-WX',    '微信公众号政策汇总',   '',                                             'LOW',    0.3, true, 'manual', 0, ''),
('MANUAL-USER',  '用户反馈/社区提交',    '',                                             'LOW',    0.3, true, 'manual', 0, ''),
('MANUAL-EXPERT','专家审核录入',         '',                                             'LOW',    0.3, true, 'manual', 0, '')
ON CONFLICT (source_id) DO NOTHING;
