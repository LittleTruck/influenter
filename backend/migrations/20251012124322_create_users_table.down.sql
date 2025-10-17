-- Migration: create_users_table (Rollback)
-- Created at: 2025-10-12 20:43:22

-- Drop indexes first
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_email;

-- Drop users table
DROP TABLE IF EXISTS users;
