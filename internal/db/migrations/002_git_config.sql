-- Add git_config column to websites for git deploy support
ALTER TABLE websites ADD COLUMN git_config JSON;