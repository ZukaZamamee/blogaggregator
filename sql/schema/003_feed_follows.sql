-- +goose Up
CREATE TABLE feed_follows(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL,
    constraint fk_users
        foreign key (user_id)
        REFERENCES users (id) ON DELETE CASCADE,
    feed_id UUID NOT NULL,
    constraint fk_feeds
        foreign key (feed_id)
        REFERENCES feeds (id) ON DELETE CASCADE,
    UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;