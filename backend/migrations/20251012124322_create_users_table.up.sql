-- Migration: create_users_table
-- Created at: 2025-10-12 20:43:22

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    profile_picture_url TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create index on email for faster lookups
CREATE INDEX idx_users_email ON users(email);

-- Create index on deleted_at for soft delete queries
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Add comment to table
COMMENT ON TABLE users IS 'User accounts table. OAuth credentials are stored in oauth_accounts table.';
