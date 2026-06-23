-- Migration: add_reply_embedding_and_good_example_to_emails
-- Created at: 2026-06-23

-- RAG few-shot：快取「這封寄出回信所回應的來信情境」之向量，供擬信時檢索相似的過往回信
ALTER TABLE emails ADD COLUMN IF NOT EXISTS context_embedding DOUBLE PRECISION[];
ALTER TABLE emails ADD COLUMN IF NOT EXISTS context_embedding_model VARCHAR(100);
ALTER TABLE emails ADD COLUMN IF NOT EXISTS context_embedding_hash VARCHAR(64);

-- 使用者可將某封寄出回信標記為「優質範例」，檢索時優先採用
ALTER TABLE emails ADD COLUMN IF NOT EXISTS is_good_example BOOLEAN NOT NULL DEFAULT FALSE;

-- 加速檢索：只索引被標記為優質範例的回信
CREATE INDEX IF NOT EXISTS idx_emails_good_example ON emails (is_good_example) WHERE is_good_example = TRUE;
