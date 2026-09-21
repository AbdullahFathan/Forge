CREATE TABLE IF NOT EXISTS permissions (
    id uuid PRIMARY KEY,
    code varchar(64) NOT NULL,
    name varchar(128) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_code ON permissions (code);

CREATE TABLE IF NOT EXISTS roles (
    id uuid PRIMARY KEY,
    code varchar(64) NOT NULL,
    name varchar(128) NOT NULL,
    is_system boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_code ON roles (code);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id uuid NOT NULL REFERENCES roles (id) ON UPDATE CASCADE ON DELETE CASCADE,
    permission_id uuid NOT NULL REFERENCES permissions (id) ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS departments (
    id uuid PRIMARY KEY,
    name varchar(128) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_departments_name ON departments (name);
CREATE INDEX IF NOT EXISTS idx_departments_deleted_at ON departments (deleted_at);

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    name varchar(128) NOT NULL,
    email varchar(255) NOT NULL,
    password_hash text NOT NULL,
    role_id uuid NOT NULL REFERENCES roles (id) ON UPDATE CASCADE ON DELETE RESTRICT,
    department_id uuid REFERENCES departments (id) ON UPDATE CASCADE ON DELETE SET NULL,
    capacity_hours_per_day integer NOT NULL DEFAULT 8,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users (role_id);
CREATE INDEX IF NOT EXISTS idx_users_department_id ON users (department_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
