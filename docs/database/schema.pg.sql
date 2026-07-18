-- =============================================================================
-- 人间烟火 (Delicious) - PostgreSQL DDL（Vercel Postgres / Neon）
-- =============================================================================

-- 1. 用户表（个人使用，无登录，仅保留 owner 标识）
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(64)  NOT NULL UNIQUE,
    email         VARCHAR(128) UNIQUE,
    nickname      VARCHAR(64)  NOT NULL DEFAULT '',
    avatar_url    VARCHAR(512),
    status        SMALLINT     NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

-- 2. 百科菜谱
CREATE TABLE IF NOT EXISTS encyclopedia_recipes (
    id              BIGSERIAL PRIMARY KEY,
    encyclopedia_recipes_name VARCHAR(128) NOT NULL,
    description     TEXT,
    cover_image_url VARCHAR(512),
    category        VARCHAR(64),
    tags            JSONB,
    ingredients     JSONB        NOT NULL,
    process_steps   JSONB        NOT NULL,
    source          VARCHAR(128),
    view_count      INTEGER      NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_encyclopedia_name ON encyclopedia_recipes (name);
CREATE INDEX IF NOT EXISTS idx_encyclopedia_category ON encyclopedia_recipes (category);
CREATE INDEX IF NOT EXISTS idx_encyclopedia_name_trgm ON encyclopedia_recipes USING gin (name gin_trgm_ops);

-- 3. 我的菜谱主表
CREATE TABLE IF NOT EXISTS my_recipes (
    id                     BIGSERIAL PRIMARY KEY,
    user_id                BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    encyclopedia_recipes_name VARCHAR(128) NOT NULL,
    current_version_id     BIGINT,
    user_rating            SMALLINT CHECK (user_rating BETWEEN 1 AND 5),
    cover_image_url        VARCHAR(512),
    encyclopedia_recipe_id BIGINT REFERENCES encyclopedia_recipes (id) ON DELETE SET NULL,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at             TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_my_recipes_user_id ON my_recipes (user_id);
CREATE INDEX IF NOT EXISTS idx_my_recipes_user_rating ON my_recipes (user_rating);
CREATE INDEX IF NOT EXISTS idx_my_recipes_created_at ON my_recipes (created_at);
CREATE INDEX IF NOT EXISTS idx_my_recipes_deleted_at ON my_recipes (deleted_at);
CREATE INDEX IF NOT EXISTS idx_my_recipes_encyclopedia ON my_recipes (encyclopedia_recipe_id);

-- 4. 菜谱版本（不可变）
CREATE TABLE IF NOT EXISTS recipe_versions (
    id             BIGSERIAL PRIMARY KEY,
    recipe_id      BIGINT       NOT NULL REFERENCES my_recipes (id) ON DELETE CASCADE,
    version_number INTEGER      NOT NULL CHECK (version_number > 0),
    ingredients    JSONB        NOT NULL,
    process_steps  JSONB        NOT NULL,
    process_text   TEXT,
    images         JSONB,
    commit_msg     VARCHAR(255) NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (recipe_id, version_number)
);
CREATE INDEX IF NOT EXISTS idx_recipe_versions_recipe_id ON recipe_versions (recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_versions_created_at ON recipe_versions (created_at);

ALTER TABLE my_recipes
    DROP CONSTRAINT IF EXISTS fk_my_recipes_current_version;
ALTER TABLE my_recipes
    ADD CONSTRAINT fk_my_recipes_current_version
        FOREIGN KEY (current_version_id) REFERENCES recipe_versions (id) ON DELETE SET NULL;

-- 可选：启用 pg_trgm 提升模糊搜索（Vercel Postgres 支持）
-- CREATE EXTENSION IF NOT EXISTS pg_trgm;
