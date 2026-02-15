-- Migration: create_cases_table
-- Created at: 2025-02-11 12:00:00

CREATE TABLE cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,

    title VARCHAR(500) NOT NULL,
    brand_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'to_confirm',
    collaboration_type VARCHAR(255),
    description TEXT,

    quoted_amount DOUBLE PRECISION,
    final_amount DOUBLE PRECISION,
    currency VARCHAR(10),

    deadline_date DATE,

    contact_name VARCHAR(255),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(100),

    notes TEXT,
    tags TEXT[],
    collaboration_items TEXT[],

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_cases_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_cases_user_id ON cases(user_id);
CREATE INDEX idx_cases_status ON cases(status);
CREATE INDEX idx_cases_deleted_at ON cases(deleted_at);

COMMENT ON TABLE cases IS 'Collaboration cases created from emails or manually';
