-- 023_relevance_video.sql
CREATE TABLE relevance_rules (
    id SERIAL PRIMARY KEY,
    category TEXT NOT NULL,
    keyword TEXT NOT NULL,
    weight INT NOT NULL DEFAULT 1,
    scope TEXT NOT NULL DEFAULT 'all',
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_relevance_rules_scope ON relevance_rules(scope) WHERE enabled;

CREATE TABLE relevance_thresholds (
    source_id TEXT PRIMARY KEY REFERENCES policy_sources(source_id),
    level1_min_score INT DEFAULT 1,
    level2_min_score INT DEFAULT 2,
    extra_keywords TEXT DEFAULT '',
    updated_at TIMESTAMPTZ DEFAULT now()
);

ALTER TABLE policy_raw_texts ADD COLUMN video_extract_status TEXT DEFAULT NULL;
CREATE INDEX idx_raw_texts_vextract ON policy_raw_texts(video_extract_status)
    WHERE video_extract_status IN ('pending', 'processing');

INSERT INTO relevance_rules (category, keyword, weight, scope) VALUES
('险种','社保',2,'all'),('险种','养老',2,'all'),('险种','养老险',2,'all'),('险种','养老金',2,'all'),
('险种','医疗',2,'all'),('险种','医保',2,'all'),('险种','失业',2,'all'),('险种','工伤',2,'all'),
('险种','生育',2,'all'),('险种','公积金',2,'all'),
('政策动词','缴费',2,'all'),('政策动词','补缴',2,'all'),('政策动词','待遇',2,'all'),
('政策动词','领取',2,'all'),('政策动词','办理',2,'all'),('政策动词','退休',2,'all'),
('政策动词','延迟退休',2,'all'),('政策动词','退休年龄',2,'all'),('政策动词','参保',2,'all'),
('政策动词','参保人',2,'all'),('政策动词','缴费年限',2,'all'),
('金额时间','补贴',1,'all'),('金额时间','报销',1,'all'),('金额时间','基数',1,'all'),
('金额时间','比例',1,'all'),('金额时间','标准',1,'all'),('金额时间','金额',1,'all'),
('金额时间','调整',1,'all'),('金额时间','上涨',1,'all'),
('人群','职工',1,'all'),('人群','灵活就业',1,'all'),('人群','退休人员',1,'all'),
('人群','居民',1,'all'),('人群','外国人',1,'all'),('人群','个体户',1,'all'),
('政策文档','政策',1,'all'),('政策文档','通知',1,'all'),('政策文档','公告',1,'all'),
('政策文档','办法',1,'all'),('政策文档','规定',1,'all'),('政策文档','方案',1,'all'),
('政策文档','意见',1,'all'),('政策文档','意见稿',1,'all');
