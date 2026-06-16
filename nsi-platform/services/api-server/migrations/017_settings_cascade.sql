CREATE TABLE IF NOT EXISTS user_settings (
    user_id TEXT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    font_scale TEXT DEFAULT 'medium',
    default_tab TEXT DEFAULT 'profile',
    notifications_on BOOLEAN DEFAULT true,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE plan_snapshots DROP CONSTRAINT IF EXISTS plan_snapshots_user_id_fkey;
ALTER TABLE plan_snapsets ADD CONSTRAINT IF NOT EXISTS plan_snapshots_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;
ALTER TABLE alerts DROP CONSTRAINT IF EXISTS alerts_user_id_fkey;
ALTER TABLE alerts ADD CONSTRAINT IF NOT EXISTS alerts_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;
ALTER TABLE payment_records DROP CONSTRAINT IF EXISTS payment_records_user_id_fkey;
ALTER TABLE payment_records ADD CONSTRAINT IF NOT EXISTS payment_records_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_user_id_fkey;
ALTER TABLE orders ADD CONSTRAINT IF NOT EXISTS orders_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;
ALTER TABLE simulator_scenarios DROP CONSTRAINT IF EXISTS simulator_scenarios_user_id_fkey;
ALTER TABLE simulator_scenarios ADD CONSTRAINT IF NOT EXISTS simulator_scenarios_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;
