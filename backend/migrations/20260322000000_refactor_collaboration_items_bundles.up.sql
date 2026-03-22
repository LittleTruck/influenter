-- Migration: refactor_collaboration_items_bundles
-- Created at: 2026-03-22
-- Description: Refactor collaboration items from parent-child tree to individual+bundle model,
--              add case<->item many-to-many, add parallel/sequential flow layout

-- 1. Add type column to collaboration_items
ALTER TABLE collaboration_items
  ADD COLUMN IF NOT EXISTS type VARCHAR(20) NOT NULL DEFAULT 'individual';

-- 2. Create bundle_items junction table
CREATE TABLE IF NOT EXISTS bundle_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bundle_id UUID NOT NULL,
    item_id UUID NOT NULL,
    "order" INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_bundle_items_bundle FOREIGN KEY (bundle_id)
      REFERENCES collaboration_items(id) ON DELETE CASCADE,
    CONSTRAINT fk_bundle_items_item FOREIGN KEY (item_id)
      REFERENCES collaboration_items(id) ON DELETE CASCADE,
    CONSTRAINT uq_bundle_items UNIQUE (bundle_id, item_id)
);
CREATE INDEX IF NOT EXISTS idx_bundle_items_bundle_id ON bundle_items(bundle_id);
CREATE INDEX IF NOT EXISTS idx_bundle_items_item_id ON bundle_items(item_id);

-- 3. Create case_collaboration_items junction table (many-to-many)
CREATE TABLE IF NOT EXISTS case_collaboration_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID NOT NULL,
    collaboration_item_id UUID NOT NULL,
    "order" INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_cci_case FOREIGN KEY (case_id)
      REFERENCES cases(id) ON DELETE CASCADE,
    CONSTRAINT fk_cci_item FOREIGN KEY (collaboration_item_id)
      REFERENCES collaboration_items(id) ON DELETE CASCADE,
    CONSTRAINT uq_case_collaboration_items UNIQUE (case_id, collaboration_item_id)
);
CREATE INDEX IF NOT EXISTS idx_cci_case_id ON case_collaboration_items(case_id);
CREATE INDEX IF NOT EXISTS idx_cci_item_id ON case_collaboration_items(collaboration_item_id);

-- 4. Add flow_layout to cases
ALTER TABLE cases
  ADD COLUMN IF NOT EXISTS flow_layout VARCHAR(20) NOT NULL DEFAULT 'parallel';

-- 5. Add collaboration_item_id to case_phases
ALTER TABLE case_phases
  ADD COLUMN IF NOT EXISTS collaboration_item_id UUID REFERENCES collaboration_items(id) ON DELETE SET NULL;

-- 6. Migrate existing data: mark parents as bundles
UPDATE collaboration_items SET type = 'bundle'
WHERE id IN (
  SELECT DISTINCT parent_id FROM collaboration_items
  WHERE parent_id IS NOT NULL AND deleted_at IS NULL
)
AND type != 'bundle';

-- 7. Create bundle_items from existing parent-child relationships
INSERT INTO bundle_items (bundle_id, item_id, "order")
SELECT parent_id, id, "order"
FROM collaboration_items
WHERE parent_id IS NOT NULL AND deleted_at IS NULL
ON CONFLICT (bundle_id, item_id) DO NOTHING;

-- 8. Remove workflow_id from bundles (bundles should not have workflows)
UPDATE collaboration_items SET workflow_id = NULL WHERE type = 'bundle' AND workflow_id IS NOT NULL
