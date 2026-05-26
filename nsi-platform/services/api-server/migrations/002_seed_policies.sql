-- 002_seed_policies: policy_claims table + seed data for MVP 5 cities
-- depends on: 001_init

BEGIN;

CREATE TABLE IF NOT EXISTS policy_claims (
    claim_id             TEXT PRIMARY KEY,
    policy_id            TEXT NOT NULL,
    region_code          TEXT NOT NULL,
    policy_type          TEXT NOT NULL CHECK (policy_type IN ('pension', 'medical', 'unemployment', 'injury', 'maternity', 'housing_fund', 'subsidy', 'training')),
    target_group_tags    TEXT[] NOT NULL DEFAULT '{}',
    subsidy_calc_method  TEXT NOT NULL,
    subsidy_amount_min   DOUBLE PRECISION,
    subsidy_amount_max   DOUBLE PRECISION,
    subsidy_duration     INT,
    effective_date       DATE NOT NULL,
    expire_date          DATE,
    confidence_score     DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (confidence_score >= 0 AND confidence_score <= 1),
    status               TEXT NOT NULL DEFAULT 'pending_review' CHECK (status IN ('verified', 'pending_review', 'unverified')),
    version_number       INT NOT NULL DEFAULT 1,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_policy_claims_region ON policy_claims(region_code);
CREATE INDEX IF NOT EXISTS idx_policy_claims_policy_id ON policy_claims(policy_id);
CREATE INDEX IF NOT EXISTS idx_policy_claims_status ON policy_claims(status);
CREATE INDEX IF NOT EXISTS idx_policy_claims_updated_at ON policy_claims(updated_at DESC);

-- ======== 上海 (310000) ========
INSERT INTO policy_claims (claim_id, policy_id, region_code, policy_type, target_group_tags, subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration, effective_date, confidence_score, status)
VALUES
('CLM-SH-PEN-001', 'SH-2025-PENSION', '310000', 'pension', '{all,employed,flexible}',
 '基数*8%+基数*16%', NULL, NULL, NULL, '2025-01-01', 0.98, 'verified'),
('CLM-SH-MED-001', 'SH-2025-MEDICAL', '310000', 'medical', '{all,employed,flexible}',
 '基数*2%+基数*9%', NULL, NULL, NULL, '2025-01-01', 0.98, 'verified'),
('CLM-SH-UNE-001', 'SH-2025-UNEMPLOYMENT', '310000', 'unemployment', '{all,employed}',
 '基数*0.5%+基数*0.5%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-SH-INJ-001', 'SH-2025-INJURY', '310000', 'injury', '{all,employed}',
 '基数*0.5%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-SH-MAT-001', 'SH-2025-MATERNITY', '310000', 'maternity', '{all,employed,female}',
 '基数*1%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-SH-HOU-001', 'SH-2025-HOUSING', '310000', 'housing_fund', '{all,employed}',
 '基数*5%~7%+基数*5%~7%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-SH-SUB-001', 'SH-2025-FLEX-SUBSIDY', '310000', 'subsidy', '{flexible,4050}',
 '基数*50%', 7380, 12383, 36, '2025-01-01', 0.90, 'verified'),
('CLM-SH-SUB-002', 'SH-2025-TRAINING', '310000', 'training', '{flexible,unemployed}',
 '按课时补贴', 500, 3000, 12, '2025-01-01', 0.85, 'verified');

-- ======== 北京 (110000) ========
INSERT INTO policy_claims (claim_id, policy_id, region_code, policy_type, target_group_tags, subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration, effective_date, confidence_score, status)
VALUES
('CLM-BJ-PEN-001', 'BJ-2025-PENSION', '110000', 'pension', '{all,employed,flexible}',
 '基数*8%+基数*16%', NULL, NULL, NULL, '2025-01-01', 0.98, 'verified'),
('CLM-BJ-MED-001', 'BJ-2025-MEDICAL', '110000', 'medical', '{all,employed,flexible}',
 '基数*2%+3+基数*9.8%', NULL, NULL, NULL, '2025-01-01', 0.98, 'verified'),
('CLM-BJ-UNE-001', 'BJ-2025-UNEMPLOYMENT', '110000', 'unemployment', '{all,employed}',
 '基数*0.5%+基数*0.5%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-BJ-INJ-001', 'BJ-2025-INJURY', '110000', 'injury', '{all,employed}',
 '基数*0.4%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-BJ-MAT-001', 'BJ-2025-MATERNITY', '110000', 'maternity', '{all,employed,female}',
 '基数*0.8%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-BJ-HOU-001', 'BJ-2025-HOUSING', '110000', 'housing_fund', '{all,employed}',
 '基数*5%~12%+基数*5%~12%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-BJ-SUB-001', 'BJ-2025-FLEX-SUBSIDY', '110000', 'subsidy', '{flexible,4050}',
 '基数*50%', 9460, 15764, 36, '2025-01-01', 0.90, 'verified');

-- ======== 深圳 (440300) ========
INSERT INTO policy_claims (claim_id, policy_id, region_code, policy_type, target_group_tags, subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration, effective_date, confidence_score, status)
VALUES
('CLM-SZ-PEN-001', 'SZ-2025-PENSION', '440300', 'pension', '{all,employed,flexible}',
 '基数*8%+基数*16%', NULL, NULL, NULL, '2025-01-01', 0.98, 'verified'),
('CLM-SZ-MED-001', 'SZ-2025-MEDICAL', '440300', 'medical', '{all,employed,flexible}',
 '基数*2%+基数*6.2%', NULL, NULL, NULL, '2025-01-01', 0.98, 'verified'),
('CLM-SZ-UNE-001', 'SZ-2025-UNEMPLOYMENT', '440300', 'unemployment', '{all,employed}',
 '基数*0.3%+基数*0.7%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-SZ-INJ-001', 'SZ-2025-INJURY', '440300', 'injury', '{all,employed}',
 '基数*0.14%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-SZ-MAT-001', 'SZ-2025-MATERNITY', '440300', 'maternity', '{all,employed,female}',
 '基数*0.45%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-SZ-HOU-001', 'SZ-2025-HOUSING', '440300', 'housing_fund', '{all,employed}',
 '基数*5%~12%+基数*5%~12%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-SZ-SUB-001', 'SZ-2025-FLEX-SUBSIDY', '440300', 'subsidy', '{flexible,4050}',
 '基数*50%', 6525, 14530, 36, '2025-01-01', 0.90, 'verified');

-- ======== 广州 (440100) ========
INSERT INTO policy_claims (claim_id, policy_id, region_code, policy_type, target_group_tags, subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration, effective_date, confidence_score, status)
VALUES
('CLM-GZ-PEN-001', 'GZ-2025-PENSION', '440100', 'pension', '{all,employed,flexible}',
 '基数*8%+基数*16%', NULL, NULL, NULL, '2025-01-01', 0.98, 'verified'),
('CLM-GZ-MED-001', 'GZ-2025-MEDICAL', '440100', 'medical', '{all,employed,flexible}',
 '基数*2%+基数*6.85%', NULL, NULL, NULL, '2025-01-01', 0.98, 'verified'),
('CLM-GZ-UNE-001', 'GZ-2025-UNEMPLOYMENT', '440100', 'unemployment', '{all,employed}',
 '基数*0.2%+基数*0.8%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-GZ-INJ-001', 'GZ-2025-INJURY', '440100', 'injury', '{all,employed}',
 '基数*0.16%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-GZ-MAT-001', 'GZ-2025-MATERNITY', '440100', 'maternity', '{all,employed,female}',
 '基数*0.85%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-GZ-HOU-001', 'GZ-2025-HOUSING', '440100', 'housing_fund', '{all,employed}',
 '基数*5%~12%+基数*5%~12%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-GZ-SUB-001', 'GZ-2025-FLEX-SUBSIDY', '440100', 'subsidy', '{flexible,4050}',
 '基数*50%', 8280, 13795, 36, '2025-01-01', 0.90, 'verified');

-- ======== 杭州 (330100) ========
INSERT INTO policy_claims (claim_id, policy_id, region_code, policy_type, target_group_tags, subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration, effective_date, confidence_score, status)
VALUES
('CLM-HZ-PEN-001', 'HZ-2025-PENSION', '330100', 'pension', '{all,employed,flexible}',
 '基数*8%+基数*15%', NULL, NULL, NULL, '2025-01-01', 0.98, 'verified'),
('CLM-HZ-MED-001', 'HZ-2025-MEDICAL', '330100', 'medical', '{all,employed,flexible}',
 '基数*2%+基数*9.5%', NULL, NULL, NULL, '2025-01-01', 0.98, 'verified'),
('CLM-HZ-UNE-001', 'HZ-2025-UNEMPLOYMENT', '330100', 'unemployment', '{all,employed}',
 '基数*0.5%+基数*0.5%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-HZ-INJ-001', 'HZ-2025-INJURY', '330100', 'injury', '{all,employed}',
 '基数*0.4%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-HZ-MAT-001', 'HZ-2025-MATERNITY', '330100', 'maternity', '{all,employed,female}',
 '基数*1%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-HZ-HOU-001', 'HZ-2025-HOUSING', '330100', 'housing_fund', '{all,employed}',
 '基数*5%~12%+基数*5%~12%', NULL, NULL, NULL, '2025-01-01', 0.95, 'verified'),
('CLM-HZ-SUB-001', 'HZ-2025-FLEX-SUBSIDY', '330100', 'subsidy', '{flexible,4050}',
 '基数*50%', 5850, 9625, 36, '2025-01-01', 0.90, 'verified');

COMMIT;
