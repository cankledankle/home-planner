-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'editor' CHECK (role IN ('admin', 'editor')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Home Plans
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    type TEXT CHECK (type IN ('single_level', 'multi_level')),
    style TEXT CHECK (style IN ('cabin', 'lodge', 'modern', 'ranch', 'farmhouse')),
    status TEXT NOT NULL DEFAULT 'incomplete' CHECK (status IN ('complete', 'incomplete', 'flagged')),
    beds INTEGER,
    baths INTEGER,
    half_baths INTEGER,
    main_sf INTEGER,
    upper_sf INTEGER,
    lower_sf INTEGER,
    porch_deck_sf INTEGER,
    garage_sf INTEGER,
    garage_apartment_sf INTEGER,
    unfinished_sf INTEGER DEFAULT 0,
    heated_sf INTEGER,
    total_sf INTEGER,
    notes TEXT,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Files
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    category TEXT NOT NULL CHECK (category IN ('website', 'reference', 'technical', '3d', 'other')),
    slot TEXT CHECK (slot IN (
        'render-front',
        'elevation-front',
        'elevation-left',
        'elevation-rear',
        'elevation-right',
        'floor-plan-main',
        'floor-plan-upper',
        'floor-plan-lower',
        'poster'
    )),
    filename TEXT NOT NULL,
    storage_key TEXT NOT NULL UNIQUE,
    file_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Activity Log
CREATE TABLE activity_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    plan_id UUID REFERENCES plans(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    detail JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Refresh Tokens
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_plans_status ON plans(status);
CREATE INDEX idx_plans_slug ON plans(slug);
CREATE INDEX idx_plans_deleted_at ON plans(deleted_at);
CREATE INDEX idx_files_plan_id ON files(plan_id);
CREATE INDEX idx_files_slot ON files(slot);
CREATE INDEX idx_files_category ON files(category);
CREATE INDEX idx_activity_log_plan_id ON activity_log(plan_id);
CREATE INDEX idx_activity_log_user_id ON activity_log(user_id);

-- Full text search on plan name
ALTER TABLE plans ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (to_tsvector('english', name)) STORED;
CREATE INDEX idx_plans_search ON plans USING GIN(search_vector);
