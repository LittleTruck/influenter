-- Migration: add_emails_direction rollback

DROP INDEX IF EXISTS idx_emails_direction;
ALTER TABLE emails DROP COLUMN direction;
