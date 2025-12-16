-- name: GetPostsForUser :many

SELECT posts.* 
FROM posts
INNER JOIN feeds ON posts.feed_id = feeds.id
WHERE feeds.id IN (
    SELECT feeds.id UNIQUE FROM feeds
    INNER JOIN feed_follows ON feed_follows.feed_id = feeds.id
    INNER JOIN users ON feed_follows.user_id = users.id
    WHERE users.id = $1
)
ORDER BY posts.published_at DESC
LIMIT $2;