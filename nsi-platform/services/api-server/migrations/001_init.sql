-- 001_init: core tables for api-server
-- depends on: none

BEGIN;

-- users: external auth gateway provides user_id; local table for sync/audit
CREATE TABLE IF NOT EXISTS users (
    user_id         TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL DEFAULT 'default',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- user_profiles: social insurance profile per PRD §4.6
CREATE TABLE IF NOT EXISTS user_profiles (
    id                          BIGSERIAL PRIMARY KEY,
    user_id                     TEXT NOT NULL REFERENCES users(user_id),
    tenant_id                   TEXT NOT NULL DEFAULT 'default',
    age                         INT NOT NULL CHECK (age >= 16 AND age <= 70),
    gender                      TEXT NOT NULL CHECK (gender IN ('male', 'female')),
    household_region_code       TEXT NOT NULL,
    current_residence_code      TEXT NOT NULL,
    employment_status           TEXT NOT NULL CHECK (employment_status IN ('employed', 'unemployed', 'flexible', 'self_employed', 'retired')),
    unemployment_reg_date       TEXT,
    flexible_employment_reg_date TEXT,
    social_security_years       INT NOT NULL DEFAULT 0 CHECK (social_security_years >= 0),
    skill_certificate_level     TEXT,
    has_children                BOOLEAN NOT NULL DEFAULT FALSE,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

-- plan_snapshots: generated insurance plan per PRD §4.8
CREATE TABLE IF NOT EXISTS plan_snapshots (
    plan_id                      TEXT PRIMARY KEY,
    user_id                      TEXT NOT NULL REFERENCES users(user_id),
    tenant_id                    TEXT NOT NULL DEFAULT 'default',
    policy_version_snapshot_id   TEXT,
    recommended_schemes          JSONB NOT NULL DEFAULT '[]',
    total_cost                   DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_subsidy                DOUBLE PRECISION NOT NULL DEFAULT 0,
    generated_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- events: event store for event-driven architecture (SSD §10.1)
CREATE TABLE IF NOT EXISTS events (
    id          BIGSERIAL PRIMARY KEY,
    event_type  TEXT NOT NULL,
    stream      TEXT NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_stream ON events(stream);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);
CREATE INDEX IF NOT EXISTS idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_plan_snapshots_user_id ON plan_snapshots(user_id);

COMMIT;
