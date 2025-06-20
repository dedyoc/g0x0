CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sha256 VARCHAR(64) UNIQUE NOT NULL,
    original_name TEXT NOT NULL,
    mime_type VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    upload_ip INET NOT NULL,
    user_agent TEXT,
    secret VARCHAR(32),
    mgmt_token VARCHAR(32) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    removed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE urls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    original_url TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_files_sha256 ON files(sha256);
CREATE INDEX idx_files_expires_at ON files(expires_at);
CREATE INDEX idx_files_removed ON files(removed);
CREATE INDEX idx_urls_original_url ON urls(original_url);
