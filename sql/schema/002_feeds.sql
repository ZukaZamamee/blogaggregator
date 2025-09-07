-- +goose Up
CREATE TABLE feeds(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT UNIQUE NOT NULL,
    url TEXT UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    constraint fk_feeds_users
        foreign key (user_id)
        REFERENCES users (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;