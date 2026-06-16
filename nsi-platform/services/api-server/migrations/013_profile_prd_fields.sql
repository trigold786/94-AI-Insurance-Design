ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS is_local_hukou BOOLEAN DEFAULT false;
ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS child_age_range TEXT DEFAULT '';
ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS has_elderly_dependents BOOLEAN DEFAULT false;
