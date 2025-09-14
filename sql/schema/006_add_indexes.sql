-- +goose Up
CREATE INDEX idx_posts_feed_id ON posts (feed_id);
CREATE INDEX idx_posts_feed_id_published_at ON posts (feed_id, published_at DESC);
CREATE INDEX idx_feed_follows_user_id ON feed_follows (user_id);
CREATE INDEX idx_feed_follows_feed_id ON feed_follows (feed_id);

-- +goose Down
DROP INDEX IF EXISTS idx_posts_feed_id_published_at;
DROP INDEX IF EXISTS idx_posts_feed_id;
DROP INDEX IF EXISTS idx_feed_follows_user_id;
DROP INDEX IF EXISTS idx_feed_follows_feed_id;