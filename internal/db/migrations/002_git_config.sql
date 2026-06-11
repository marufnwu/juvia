-- Add git_config column to websites for git deploy support (idempotent)
ALTER TABLE websites ADD COLUMN git_config JSON;
-- Ignore error if column already exists