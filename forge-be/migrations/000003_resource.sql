CREATE TABLE IF NOT EXISTS resource_allocations (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT,
    project_id uuid NOT NULL REFERENCES projects (id) ON UPDATE CASCADE ON DELETE RESTRICT,
    allocation_percent numeric(5, 2) NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    role varchar(32) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT resource_allocations_percent_check CHECK (allocation_percent >= 0 AND allocation_percent <= 100),
    CONSTRAINT resource_allocations_dates_check CHECK (end_date >= start_date),
    CONSTRAINT resource_allocations_role_check CHECK (role IN ('LEAD', 'MEMBER', 'VIEWER'))
);
CREATE INDEX IF NOT EXISTS idx_resource_allocations_user_dates ON resource_allocations (user_id, start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_resource_allocations_project_id ON resource_allocations (project_id);
CREATE INDEX IF NOT EXISTS idx_resource_allocations_deleted_at ON resource_allocations (deleted_at);

CREATE TABLE IF NOT EXISTS holidays (
    id uuid PRIMARY KEY,
    date date NOT NULL,
    name varchar(128) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT holidays_date_unique UNIQUE (date)
);

CREATE TABLE IF NOT EXISTS user_skills (
    user_id uuid NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    skill varchar(64) NOT NULL,
    PRIMARY KEY (user_id, skill)
);
CREATE INDEX IF NOT EXISTS idx_user_skills_skill ON user_skills (skill);
