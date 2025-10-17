-- Migration: create_emails_table
-- Created at: 2025-10-17 16:27:23

-- Create emails table
CREATE TABLE emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    oauth_account_id UUID NOT NULL,
    
    -- Email provider original information
    provider_message_id VARCHAR(255) NOT NULL UNIQUE,
    thread_id VARCHAR(255),
    
    -- Email basic information
    from_email VARCHAR(255) NOT NULL,
    from_name VARCHAR(255),
    to_email VARCHAR(255),
    subject TEXT,
    body_text TEXT,
    body_html TEXT,
    snippet TEXT,
    
    -- Email properties
    received_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    has_attachments BOOLEAN DEFAULT FALSE,
    labels TEXT[],
    
    -- AI analysis status
    ai_analyzed BOOLEAN DEFAULT FALSE,
    ai_analysis_id UUID,
    
    -- Case association
    case_id UUID,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Foreign keys
    CONSTRAINT fk_emails_oauth_account FOREIGN KEY (oauth_account_id) REFERENCES oauth_accounts(id) ON DELETE CASCADE
    -- CONSTRAINT fk_emails_ai_analysis FOREIGN KEY (ai_analysis_id) REFERENCES ai_analysis(id) ON DELETE SET NULL,
    -- CONSTRAINT fk_emails_case FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE SET NULL
);

-- Create indexes for faster lookups
CREATE INDEX idx_emails_oauth_account_id ON emails(oauth_account_id);
CREATE INDEX idx_emails_provider_message_id ON emails(provider_message_id);
CREATE INDEX idx_emails_thread_id ON emails(thread_id);
CREATE INDEX idx_emails_from_email ON emails(from_email);
CREATE INDEX idx_emails_to_email ON emails(to_email);
CREATE INDEX idx_emails_received_at ON emails(received_at DESC);
CREATE INDEX idx_emails_case_id ON emails(case_id);
CREATE INDEX idx_emails_deleted_at ON emails(deleted_at);

-- Partial index for unanalyzed emails (performance optimization)
CREATE INDEX idx_emails_ai_analyzed ON emails(ai_analyzed) WHERE ai_analyzed = FALSE;

-- Index for searching by subject (using GIN index for full-text search)
CREATE INDEX idx_emails_subject_gin ON emails USING GIN (to_tsvector('english', COALESCE(subject, '')));

-- Add comments
COMMENT ON TABLE emails IS 'Emails synced from third-party accounts (Gmail, Outlook, etc.)';
COMMENT ON COLUMN emails.provider_message_id IS 'Gmail message ID or other provider message ID';
COMMENT ON COLUMN emails.thread_id IS 'Email thread ID';
COMMENT ON COLUMN emails.labels IS 'Email labels (Gmail labels or other provider tags)';
COMMENT ON COLUMN emails.snippet IS 'Email preview snippet (first 150 characters)';
COMMENT ON COLUMN emails.ai_analyzed IS 'Whether the email has been analyzed by AI';

