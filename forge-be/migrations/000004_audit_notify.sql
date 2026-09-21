ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email_notifications_enabled boolean NOT NULL DEFAULT true;

CREATE TABLE IF NOT EXISTS audit_logs (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    ip_address varchar(64) NOT NULL DEFAULT '',
    entity_type varchar(64) NOT NULL,
    entity_id uuid NOT NULL,
    project_id uuid,
    action varchar(32) NOT NULL,
    before jsonb,
    after jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs (user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs (entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_project_id ON audit_logs (project_id);

CREATE TABLE IF NOT EXISTS notifications (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    type varchar(64) NOT NULL,
    title varchar(255) NOT NULL,
    body text NOT NULL DEFAULT '',
    is_read boolean NOT NULL DEFAULT false,
    metadata jsonb,
    entity_id uuid NOT NULL,
    day date NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_notifications_idempotency
    ON notifications (type, user_id, entity_id, day);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications (user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications (user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications (created_at);
