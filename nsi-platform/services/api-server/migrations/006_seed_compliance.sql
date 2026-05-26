-- 006_seed_compliance: add compliance conditions + required documents to seed policies
-- depends on: 004_add_policy_compliance

BEGIN;

-- 上海 灵活就业补贴
UPDATE policy_claims SET
  conditions = '[{"name":"灵活就业登记","description":"已在街道办理灵活就业登记","tag_match":"flexile_employment","required":true},{"name":"年龄要求","description":"女性40岁以上或男性50岁以上","tag_match":"4050","required":true}]',
  required_documents = '[{"name":"身份证","description":"原件及复印件","source":"user","optional":false},{"name":"灵活就业证明","description":"街道开具","source":"user","optional":false},{"name":"社保缴费记录","description":"近6个月","source":"gov","optional":false}]'
WHERE claim_id = 'CLM-SH-SUB-001';

-- 上海 培训补贴
UPDATE policy_claims SET
  conditions = '[{"name":"失业登记","description":"已办理失业登记","tag_match":"unemployed","required":true},{"name":"培训完成","description":"完成指定培训课程","tag_match":"","required":true}]',
  required_documents = '[{"name":"身份证","description":"原件及复印件","source":"user","optional":false},{"name":"失业登记证","description":"街道开具","source":"user","optional":false},{"name":"培训结业证书","description":"培训机构出具","source":"user","optional":false}]'
WHERE claim_id = 'CLM-SH-SUB-002';

-- 北京 灵活就业补贴
UPDATE policy_claims SET
  conditions = '[{"name":"灵活就业登记","description":"已办理灵活就业登记","tag_match":"flexible_employment","required":true}]',
  required_documents = '[{"name":"身份证","description":"原件及复印件","source":"user","optional":false},{"name":"灵活就业申请表","description":"社区领取","source":"user","optional":false}]'
WHERE claim_id LIKE 'CLM-BJ-SUB%';

-- 深圳 4050补贴
UPDATE policy_claims SET
  conditions = '[{"name":"年龄要求","description":"女性40岁以上/男性50岁以上","tag_match":"4050","required":true},{"name":"失业登记","description":"已办理失业登记","tag_match":"unemployed","required":false}]',
  required_documents = '[{"name":"身份证","description":"原件及复印件","source":"user","optional":false},{"name":"年龄证明","description":"身份证即可","source":"user","optional":false},{"name":"失业登记证","description":"如有","source":"user","optional":true}]'
WHERE claim_id LIKE 'CLM-SZ-SUB%' OR claim_id LIKE 'CLM-SZ-TRA%';

-- 广州 技能培训补贴
UPDATE policy_claims SET
  conditions = '[{"name":"社保缴纳","description":"累计缴纳社保满12个月","tag_match":"","required":true}]',
  required_documents = '[{"name":"身份证","description":"原件及复印件","source":"user","optional":false},{"name":"社保缴纳证明","description":"社保局开具","source":"gov","optional":false},{"name":"技能证书","description":"如已取得","source":"user","optional":true}]'
WHERE (claim_id LIKE 'CLM-GZ-SUB%' OR claim_id LIKE 'CLM-GZ-TRA%') AND policy_type IN ('subsidy', 'training');

-- 杭州 人才补贴
UPDATE policy_claims SET
  conditions = '[{"name":"学历要求","description":"本科及以上学历","tag_match":"","required":true},{"name":"社保缴纳","description":"在杭缴纳社保","tag_match":"employed","required":true}]',
  required_documents = '[{"name":"身份证","description":"原件及复印件","source":"user","optional":false},{"name":"学历证书","description":"学信网认证","source":"user","optional":false},{"name":"劳动合同","description":"加盖公章","source":"user","optional":false}]'
WHERE claim_id LIKE 'CLM-HZ-SUB%' OR claim_id LIKE 'CLM-HZ-TRA%';

COMMIT;
