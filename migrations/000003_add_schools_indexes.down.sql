-- Filename: migrations/000003_add_schools_indexes.down.sql

DROP INDEX IF EXISTS schools_name_idx;
DROP INDEX IF EXISTS schools_level_idx;
DROP INDEX IF EXISTS schools_mode_idx;
