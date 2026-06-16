CREATE TABLE IF NOT EXISTS simulator_scenarios (
    id SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '方案A',
    params JSONB NOT NULL,
    result JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sim_scenarios_user ON simulator_scenarios(user_id);

-- 渐进式最低缴费年限政策（2025年延迟退休改革）
-- 2030年起从15年逐步提高到20年（每年+6个月）
INSERT INTO policy_claims (claim_id, policy_id, region_code, policy_type, target_group_tags,
    subsidy_calc_method, effective_date, confidence_score, status, version_number,
    conditions, required_documents, policy_title, issuing_authority, document_number, source_type)
VALUES (
    'POLICY-MIN-YEARS-2030', 'MIN-CONTRIB-YEARS', '000000', 'pension',
    ARRAY['flexible_employment', 'unemployed', 'employed'],
    '', '2025-01-01', 1.0, 'verified', 1,
    '[{"name":"最低缴费年限","type":"gradual_min_years","description":"2030年起最低缴费年限从15年逐步提高到20年","base_year":2030,"base_value":15,"increment_per_year":0.5,"max_value":20,"max_year":2039,"tag_match":""}]'::jsonb,
    '[]'::jsonb,
    '渐进式延迟法定退休年龄办法—最低缴费年限',
    '全国人民代表大会常务委员会',
    '2024年9月',
    'gov_doc'
) ON CONFLICT (claim_id) DO NOTHING;
