-- 004_crawl_tables: raw text storage + crawl logs
-- depends on: 003_seed_sources

BEGIN;

CREATE TABLE IF NOT EXISTS policy_raw_texts (
    id           BIGSERIAL PRIMARY KEY,
    source_id    TEXT NOT NULL REFERENCES policy_sources(source_id),
    title        TEXT NOT NULL DEFAULT '',
    content      TEXT NOT NULL,
    source_url   TEXT NOT NULL DEFAULT '',
    version_hash TEXT NOT NULL DEFAULT '',
    fetch_time   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_raw_texts_source ON policy_raw_texts(source_id);
CREATE INDEX IF NOT EXISTS idx_raw_texts_hash ON policy_raw_texts(version_hash);

CREATE TABLE IF NOT EXISTS crawl_logs (
    id            BIGSERIAL PRIMARY KEY,
    source_id     TEXT NOT NULL REFERENCES policy_sources(source_id),
    status        TEXT NOT NULL CHECK (status IN ('success', 'failed')),
    error_message TEXT NOT NULL DEFAULT '',
    crawled_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_crawl_logs_source ON crawl_logs(source_id);
CREATE INDEX IF NOT EXISTS idx_crawl_logs_time ON crawl_logs(crawled_at DESC);

COMMIT;
