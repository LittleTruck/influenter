-- Migration: add_emails_direction
-- 新增 direction 欄位區分收件/寄件

ALTER TABLE emails ADD COLUMN direction VARCHAR(20) NOT NULL DEFAULT 'incoming';
COMMENT ON COLUMN emails.direction IS 'incoming: 收到的郵件, outgoing: 寄出的郵件';
CREATE INDEX idx_emails_direction ON emails(direction);
