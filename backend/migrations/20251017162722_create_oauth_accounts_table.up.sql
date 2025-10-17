-- Migration: create_oauth_accounts_table
-- Created at: 2025-10-17 16:27:22

-- Create oauth_accounts table
CREATE TABLE oauth_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    
    -- OAuth provider information
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255),
    email VARCHAR(255) NOT NULL,
    
    -- OAuth tokens (encrypted - AES-256-GCM)
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expiry TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Sync status (for email sync)
    last_sync_at TIMESTAMP WITH TIME ZONE,
    last_history_id VARCHAR(100),
    sync_status VARCHAR(50) DEFAULT 'active',
    sync_error TEXT,
    
    -- Additional metadata (JSON format)
    metadata JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Foreign key
    CONSTRAINT fk_oauth_accounts_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Unique constraint: one provider account per user
    CONSTRAINT unique_user_provider_email UNIQUE (user_id, provider, email)
);

-- Create indexes for faster lookups
CREATE INDEX idx_oauth_accounts_user_id ON oauth_accounts(user_id);
CREATE INDEX idx_oauth_accounts_provider ON oauth_accounts(provider);
CREATE INDEX idx_oauth_accounts_sync_status ON oauth_accounts(sync_status);
CREATE INDEX idx_oauth_accounts_deleted_at ON oauth_accounts(deleted_at);
CREATE INDEX idx_oauth_accounts_token_expiry ON oauth_accounts(token_expiry);

-- Add comments
COMMENT ON TABLE oauth_accounts IS 'Third-party OAuth accounts (Google, Outlook, etc.) linked to users';
COMMENT ON COLUMN oauth_accounts.access_token IS 'Encrypted access token using AES-256-GCM';
COMMENT ON COLUMN oauth_accounts.refresh_token IS 'Encrypted refresh token using AES-256-GCM';
COMMENT ON COLUMN oauth_accounts.sync_status IS 'Sync status: active, paused, error';
COMMENT ON COLUMN oauth_accounts.last_history_id IS 'Gmail API history ID or other provider sync ID';

