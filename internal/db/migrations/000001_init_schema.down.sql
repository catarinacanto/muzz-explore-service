DROP TRIGGER IF EXISTS update_decisions_updated_at ON decisions;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_recipient_likes;
DROP TABLE IF EXISTS decisions;