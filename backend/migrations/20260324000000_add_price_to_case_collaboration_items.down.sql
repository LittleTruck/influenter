-- Remove per-case price from case_collaboration_items
ALTER TABLE case_collaboration_items
    DROP COLUMN IF EXISTS price;
