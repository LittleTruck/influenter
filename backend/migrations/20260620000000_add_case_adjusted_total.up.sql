-- 案件總價可手動調整（單一廠商殺價時微調合作費用），並保留調整歷史
ALTER TABLE cases ADD COLUMN adjusted_total NUMERIC(12,2) DEFAULT NULL;
COMMENT ON COLUMN cases.adjusted_total IS 'Manually adjusted (negotiated) case total. NULL means use the sum of collaboration item prices.';

-- 總價調整歷史
CREATE TABLE case_total_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID NOT NULL,
    user_id UUID NOT NULL,
    original_total NUMERIC(12,2) NOT NULL,
    adjusted_total NUMERIC(12,2) NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_case_total_adjustments_case FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,
    CONSTRAINT fk_case_total_adjustments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_case_total_adjustments_case_id ON case_total_adjustments(case_id);
