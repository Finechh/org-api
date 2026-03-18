-- +goose Up
CREATE TABLE IF NOT EXISTS departments (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(200) NOT NULL,
    parent_id  INTEGER REFERENCES departments(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_dept_name_parent ON departments (name, COALESCE(parent_id, 0));

CREATE TABLE IF NOT EXISTS employees (
    id            SERIAL PRIMARY KEY,
    department_id INTEGER NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    full_name     VARCHAR(200) NOT NULL,
    position      VARCHAR(200) NOT NULL,
    hired_at      DATE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS departments;