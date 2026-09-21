CREATE TABLE IF NOT EXISTS projects (
    id uuid PRIMARY KEY,
    name varchar(200) NOT NULL,
    description text NOT NULL DEFAULT '',
    status varchar(32) NOT NULL,
    priority varchar(32) NOT NULL,
    start_date date NOT NULL,
    target_end_date date NOT NULL,
    owner_id uuid NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT,
    department_id uuid REFERENCES departments (id) ON UPDATE CASCADE ON DELETE SET NULL,
    tags jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT projects_status_check CHECK (status IN ('DRAFT', 'ACTIVE', 'ON_HOLD', 'COMPLETED', 'ARCHIVED')),
    CONSTRAINT projects_priority_check CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    CONSTRAINT projects_dates_check CHECK (target_end_date >= start_date)
);
CREATE INDEX IF NOT EXISTS idx_projects_status_department_id ON projects (status, department_id);
CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects (owner_id);
CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON projects (deleted_at);

CREATE TABLE IF NOT EXISTS project_members (
    id uuid PRIMARY KEY,
    project_id uuid NOT NULL REFERENCES projects (id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT,
    role varchar(32) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT project_members_role_check CHECK (role IN ('LEAD', 'MEMBER', 'VIEWER')),
    CONSTRAINT project_members_unique UNIQUE (project_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members (user_id);

CREATE TABLE IF NOT EXISTS tasks (
    id uuid PRIMARY KEY,
    project_id uuid NOT NULL REFERENCES projects (id) ON UPDATE CASCADE ON DELETE CASCADE,
    parent_task_id uuid REFERENCES tasks (id) ON UPDATE CASCADE ON DELETE RESTRICT,
    name varchar(200) NOT NULL,
    description text NOT NULL DEFAULT '',
    status varchar(32) NOT NULL,
    priority varchar(32) NOT NULL,
    estimated_hours numeric(10, 2),
    start_date date,
    due_date date,
    labels jsonb NOT NULL DEFAULT '[]'::jsonb,
    position integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT tasks_status_check CHECK (status IN ('BACKLOG', 'TODO', 'IN_PROGRESS', 'IN_REVIEW', 'DONE', 'BLOCKED')),
    CONSTRAINT tasks_priority_check CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL'))
);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id_status ON tasks (project_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks (due_date);
CREATE INDEX IF NOT EXISTS idx_tasks_parent_task_id ON tasks (parent_task_id);
CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks (deleted_at);

CREATE TABLE IF NOT EXISTS task_assignees (
    task_id uuid NOT NULL REFERENCES tasks (id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT,
    PRIMARY KEY (task_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_task_assignees_user_id ON task_assignees (user_id);

CREATE TABLE IF NOT EXISTS task_dependencies (
    id uuid PRIMARY KEY,
    task_id uuid NOT NULL REFERENCES tasks (id) ON UPDATE CASCADE ON DELETE CASCADE,
    depends_on_task_id uuid NOT NULL REFERENCES tasks (id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT task_dependencies_unique UNIQUE (task_id, depends_on_task_id),
    CONSTRAINT task_dependencies_no_self CHECK (task_id <> depends_on_task_id)
);
CREATE INDEX IF NOT EXISTS idx_task_dependencies_depends_on ON task_dependencies (depends_on_task_id);

CREATE TABLE IF NOT EXISTS task_comments (
    id uuid PRIMARY KEY,
    task_id uuid NOT NULL REFERENCES tasks (id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT,
    body text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_task_comments_task_id ON task_comments (task_id);
