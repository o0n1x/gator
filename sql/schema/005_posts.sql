-- +goose Up
CREATE TABLE posts(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    title TEXT,
    url TEXT UNIQUE,
    description TEXT,
    published_at TIMESTAMP,
    feed_id UUID,
    FOREIGN KEY(feed_id) REFERENCES feeds (id)
);

-- +goose Down
DROP TABLE posts;