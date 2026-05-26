-- 003_add_plan_updated_at: add updated_at to plan_snapshots
-- depends on: 001_init

BEGIN;

ALTER TABLE plan_snapshots ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

COMMIT;
