-- PoSA Database Migration: 001_initial_schema
-- Applied to: Supabase PostgreSQL
-- Date: 2026-05-06
--
-- Run manually via:
--   psql $SUPABASE_DB_URL -f backend/database/migrations/001_initial_schema.sql
--
-- Or auto-applied by the Go backend on startup (database.Open).

-- ============================================================
-- Tables
-- ============================================================

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    github_id BIGINT UNIQUE NOT NULL,
    username TEXT NOT NULL,
    avatar_url TEXT DEFAULT '',
    email TEXT DEFAULT '',
    role TEXT DEFAULT 'user',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS submissions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    score INTEGER DEFAULT 0,
    cid TEXT DEFAULT '',
    tx_hash TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- Indexes
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_submissions_user ON submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_users_github_id ON users(github_id);

-- ============================================================
-- Row Level Security (RLS)
-- ============================================================

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE submissions ENABLE ROW LEVEL SECURITY;

-- Service role (Go backend connects as postgres) has full access.
-- These policies allow the backend to perform all operations.
DROP POLICY IF EXISTS "Service role full access on users" ON users;
CREATE POLICY "Service role full access on users" ON users
    FOR ALL USING (true) WITH CHECK (true);

DROP POLICY IF EXISTS "Service role full access on sessions" ON sessions;
CREATE POLICY "Service role full access on sessions" ON sessions
    FOR ALL USING (true) WITH CHECK (true);

DROP POLICY IF EXISTS "Service role full access on submissions" ON submissions;
CREATE POLICY "Service role full access on submissions" ON submissions
    FOR ALL USING (true) WITH CHECK (true);

-- Client-side policies (for future Supabase JS client access).
-- Users can only read their own data via Supabase client SDK.
DROP POLICY IF EXISTS "Users read own submissions" ON submissions;
CREATE POLICY "Users read own submissions" ON submissions
    FOR SELECT USING (auth.uid()::text = user_id::text);

DROP POLICY IF EXISTS "Users read own profile" ON users;
CREATE POLICY "Users read own profile" ON users
    FOR SELECT USING (auth.uid()::text = github_id::text);
