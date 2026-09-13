-- aasplit 数据库初始化脚本（参考）
-- 实际表结构由后端 GORM AutoMigrate 自动维护；本脚本用于手工初始化参考。
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(64)  NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    nickname      VARCHAR(64)  NOT NULL,
    email         VARCHAR(128) UNIQUE,
    avatar        VARCHAR(512) DEFAULT '',
    role          VARCHAR(16)  NOT NULL DEFAULT 'user',
    status        VARCHAR(16)  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- 分账群组表
CREATE TABLE IF NOT EXISTS groups (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(512) DEFAULT '',
    owner_id    BIGINT       NOT NULL,
    status      VARCHAR(16)  NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- 群组成员表
CREATE TABLE IF NOT EXISTS group_members (
    id         BIGSERIAL PRIMARY KEY,
    group_id   BIGINT      NOT NULL,
    user_id    BIGINT      NOT NULL,
    role       VARCHAR(16) NOT NULL DEFAULT 'normal',
    invited_by BIGINT      NOT NULL DEFAULT 0,
    status     VARCHAR(16) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (group_id, user_id)
);

-- 消费记录表
CREATE TABLE IF NOT EXISTS expenses (
    id          BIGSERIAL PRIMARY KEY,
    group_id    BIGINT           NOT NULL,
    title       VARCHAR(128)     NOT NULL,
    amount      DOUBLE PRECISION NOT NULL,
    category    VARCHAR(32)      NOT NULL,
    payer_id    BIGINT           NOT NULL,
    split_type  VARCHAR(16)      NOT NULL,
    paid_at     TIMESTAMPTZ      NOT NULL,
    receipt_url VARCHAR(512) DEFAULT '',
    status      VARCHAR(16)      NOT NULL DEFAULT 'active',
    created_by  BIGINT           NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ      NOT NULL DEFAULT now()
);

-- 分摊明细表
CREATE TABLE IF NOT EXISTS expense_shares (
    id           BIGSERIAL PRIMARY KEY,
    expense_id   BIGINT           NOT NULL,
    user_id      BIGINT           NOT NULL,
    share_amount DOUBLE PRECISION NOT NULL,
    ratio        DOUBLE PRECISION NOT NULL DEFAULT 0,
    status       VARCHAR(16)      NOT NULL DEFAULT 'unsettled',
    created_at   TIMESTAMPTZ      NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ      NOT NULL DEFAULT now(),
    UNIQUE (expense_id, user_id)
);

-- 结算建议表
CREATE TABLE IF NOT EXISTS settlements (
    id           BIGSERIAL PRIMARY KEY,
    group_id     BIGINT           NOT NULL,
    from_user_id BIGINT           NOT NULL,
    to_user_id   BIGINT           NOT NULL,
    amount       DOUBLE PRECISION NOT NULL,
    status       VARCHAR(16)      NOT NULL DEFAULT 'pending',
    settled_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ      NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ      NOT NULL DEFAULT now()
);

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT        NOT NULL,
    action        VARCHAR(64)   NOT NULL,
    resource_type VARCHAR(64)   NOT NULL,
    resource_id   VARCHAR(64)   DEFAULT '',
    detail        VARCHAR(1024) DEFAULT '',
    ip            VARCHAR(64)   DEFAULT '',
    request_id    VARCHAR(64)   DEFAULT '',
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_expenses_group ON expenses (group_id);
CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses (category);
CREATE INDEX IF NOT EXISTS idx_expenses_paid_at ON expenses (paid_at);
CREATE INDEX IF NOT EXISTS idx_settlements_group ON settlements (group_id);
CREATE INDEX IF NOT EXISTS idx_settlements_status ON settlements (status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs (user_id);
