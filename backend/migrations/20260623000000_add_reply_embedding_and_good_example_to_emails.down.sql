-- Rollback: add_reply_embedding_and_good_example_to_emails

DROP INDEX IF EXISTS idx_emails_good_example;
ALTER TABLE emails DROP COLUMN IF EXISTS is_good_example;
ALTER TABLE emails DROP COLUMN IF EXISTS context_embedding_hash;
ALTER TABLE emails DROP COLUMN IF EXISTS context_embedding_model;
ALTER TABLE emails DROP COLUMN IF EXISTS context_embedding;
