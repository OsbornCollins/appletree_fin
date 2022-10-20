-- Filename: migrations/000001_create_schools_table.up.sql

CREATE TABLE IF NOT EXISTS schools (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    level text NOT NULL,
    contact text NOT NULL,
    phone text NOT NULL,
    email text NOT NULL,
    website text NOT NULL,
    address text NOT NULL,
    mode text[] NOT NULL,
    version int NOT NULL DEFAULT 1
);