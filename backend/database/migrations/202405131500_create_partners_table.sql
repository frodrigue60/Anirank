-- Create partners table for community links and alliances
CREATE TABLE IF NOT EXISTS partners (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    banner TEXT, -- Path to the image in storage
    description TEXT,
    type TEXT NOT NULL DEFAULT 'alliance', -- 'source', 'alliance', 'discord'
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for sorting and active status
CREATE INDEX IF NOT EXISTS idx_partners_active_order ON partners(is_active, sort_order);
