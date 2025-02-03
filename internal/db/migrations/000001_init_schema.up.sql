CREATE TABLE decisions (
                           actor_user_id TEXT NOT NULL,
                           recipient_user_id TEXT NOT NULL,
                           liked BOOLEAN NOT NULL,
                           created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           PRIMARY KEY (actor_user_id, recipient_user_id)
);

CREATE INDEX idx_recipient_likes ON decisions (recipient_user_id, liked, created_at DESC);

CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_decisions_updated_at
    BEFORE UPDATE ON decisions
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
