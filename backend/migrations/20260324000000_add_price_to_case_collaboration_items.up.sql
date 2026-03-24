-- Add per-case price to case_collaboration_items
-- Each case can negotiate different prices for the same collaboration item
ALTER TABLE case_collaboration_items
    ADD COLUMN price NUMERIC(12, 2) DEFAULT NULL;

COMMENT ON COLUMN case_collaboration_items.price IS 'Per-case negotiated price. NULL means use the default price from collaboration_items.price';
