-- Down migration: revert collaboration items bundle refactor

-- 1. Remove collaboration_item_id from case_phases
ALTER TABLE case_phases
  DROP CONSTRAINT IF EXISTS fk_case_phases_collaboration_item,
  DROP COLUMN IF EXISTS collaboration_item_id;

-- 2. Remove flow_layout from cases
ALTER TABLE cases
  DROP COLUMN IF EXISTS flow_layout;

-- 3. Drop case_collaboration_items table
DROP TABLE IF EXISTS case_collaboration_items;

-- 4. Drop bundle_items table
DROP TABLE IF EXISTS bundle_items;

-- 5. Remove type column from collaboration_items
ALTER TABLE collaboration_items
  DROP COLUMN IF EXISTS type;
