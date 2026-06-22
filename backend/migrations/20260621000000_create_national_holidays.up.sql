-- 國定假日 / 補班日資料表
-- 用途：作為「工作日」計算（排除週末與國定假日）的單一真相來源；前端透過 API 取得同一份資料。
-- 設計：
--   * source='national'  → 系統內建國定假日（全域，oauth_account_id 為 NULL）。
--   * source='custom' + oauth_account_id → 未來的帳號自訂假日（drop-in，無需改結構）。
--   * is_workday=true     → 補班日（即使落在週末仍須上班）。台灣自 2025 下半年起取消補班，
--                            故目前無補班列，欄位保留以備未來。
-- 注意：僅收錄「落在平日（週一~週五）」的放假日；週末由計算引擎自動視為非工作日，不重複收錄。
CREATE TABLE national_holidays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_workday BOOLEAN NOT NULL DEFAULT FALSE,
    source VARCHAR(50) NOT NULL DEFAULT 'national',
    oauth_account_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_national_holidays_account FOREIGN KEY (oauth_account_id) REFERENCES oauth_accounts(id) ON DELETE CASCADE
);
CREATE INDEX idx_national_holidays_date ON national_holidays(date);
CREATE INDEX idx_national_holidays_source ON national_holidays(source);
-- 防止內建來源重複收錄同一天（national 為全域）
CREATE UNIQUE INDEX uq_national_holidays_national_date ON national_holidays(date) WHERE source = 'national';

-- ── Seed：2026（民國115年）平日國定假日（行政院人事行政總處核定，全年 0 補班日）──
INSERT INTO national_holidays (date, name) VALUES
    ('2026-01-01', '元旦'),
    ('2026-02-16', '農曆除夕'),
    ('2026-02-17', '春節（初一）'),
    ('2026-02-18', '春節（初二）'),
    ('2026-02-19', '春節（初三）'),
    ('2026-02-20', '小年夜補假'),
    ('2026-02-27', '和平紀念日補假'),
    ('2026-04-03', '兒童節補假'),
    ('2026-04-06', '清明節補假'),
    ('2026-05-01', '勞動節'),
    ('2026-06-19', '端午節'),
    ('2026-09-25', '中秋節'),
    ('2026-09-28', '教師節'),
    ('2026-10-09', '國慶日補假'),
    ('2026-10-26', '臺灣光復節補假'),
    ('2026-12-25', '行憲紀念日');

-- ── Seed：2027（民國116年）平日國定假日（行政院人事行政總處核定，全年 0 補班日）──
INSERT INTO national_holidays (date, name) VALUES
    ('2027-01-01', '元旦'),
    ('2027-02-04', '小年夜'),
    ('2027-02-05', '農曆除夕'),
    ('2027-02-08', '春節（初三）'),
    ('2027-02-09', '春節補假'),
    ('2027-02-10', '春節補假'),
    ('2027-03-01', '和平紀念日補假'),
    ('2027-04-05', '清明節'),
    ('2027-04-06', '兒童節補假'),
    ('2027-04-30', '勞動節補假'),
    ('2027-06-09', '端午節'),
    ('2027-09-15', '中秋節'),
    ('2027-09-28', '教師節'),
    ('2027-10-11', '國慶日補假'),
    ('2027-10-25', '臺灣光復節'),
    ('2027-12-24', '行憲紀念日補假'),
    ('2027-12-31', '開國紀念日補假');
