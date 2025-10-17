-- Migration: create_emails_table (rollback)
-- Created at: 2025-10-17 16:27:23

-- Drop indexes
DROP INDEX IF EXISTS idx_emails_subject_gin;
DROP INDEX IF EXISTS idx_emails_ai_analyzed;
DROP INDEX IF EXISTS idx_emails_deleted_at;
DROP INDEX IF EXISTS idx_emails_case_id;
DROP INDEX IF EXISTS idx_emails_received_at;
DROP INDEX IF EXISTS idx_emails_to_email;
DROP INDEX IF EXISTS idx_emails_from_email;
DROP INDEX IF EXISTS idx_emails_thread_id;
DROP INDEX IF EXISTS idx_emails_provider_message_id;
DROP INDEX IF EXISTS idx_emails_oauth_account_id;

-- Drop table
DROP TABLE IF EXISTS emails;

