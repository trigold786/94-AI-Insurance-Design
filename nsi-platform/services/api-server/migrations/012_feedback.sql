CREATE TABLE IF NOT EXISTS feedback (
    id SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    content TEXT NOT NULL,
    contact TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT now()
);
