-- Migration: create_oauth_accounts_table (rollback)
-- Created at: 2025-10-17 16:27:22

-- Drop indexes
DROP INDEX IF EXISTS idx_oauth_accounts_token_expiry;
DROP INDEX IF EXISTS idx_oauth_accounts_deleted_at;
DROP INDEX IF EXISTS idx_oauth_accounts_sync_status;
DROP INDEX IF EXISTS idx_oauth_accounts_provider;
DROP INDEX IF EXISTS idx_oauth_accounts_user_id;

-- Drop table
DROP TABLE IF EXISTS oauth_accounts;

