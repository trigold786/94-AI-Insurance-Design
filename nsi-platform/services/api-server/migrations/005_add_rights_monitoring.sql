-- 005_add_rights_monitoring: payment records + alerts
-- depends on: 001_init

BEGIN;

CREATE TABLE IF NOT EXISTS payment_records (
    record_id   TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(user_id),
    policy_type TEXT NOT NULL,
    month       TEXT NOT NULL,          -- YYYY-MM
    amount      DOUBLE PRECISION NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('paid', 'pending', 'missed')),
    due_date    TEXT NOT NULL,          -- YYYY-MM-DD
    paid_date   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_user ON payment_records(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_status ON payment_records(status);

CREATE TABLE IF NOT EXISTS alerts (
    alert_id   TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(user_id),
    alert_type TEXT NOT NULL CHECK (alert_type IN ('disconnection_risk', 'policy_change')),
    severity   TEXT NOT NULL DEFAULT 'medium' CHECK (severity IN ('high', 'medium', 'low')),
    title      TEXT NOT NULL,
    message    TEXT NOT NULL,
    is_read    BOOLEAN NOT NULL DEFAULT FALSE,
    policy_id  TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_user ON alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_alerts_unread ON alerts(user_id, is_read) WHERE NOT is_read;

COMMIT;
