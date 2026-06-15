-- Add CSRF token hash to sessions for stateless CSRF validation
ALTER TABLE sessions ADD COLUMN csrf_token_hash VARCHAR(255);