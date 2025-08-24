
-- name: CreateLink :one
INSERT INTO links (
    alias,
    original_url,
    user_id 
) VALUES (
    $1, $2, $3 
)
RETURNING *;

-- name: GetLinksByUserID :many
SELECT * FROM links
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetLinkByAlias :one
SELECT * FROM links
WHERE alias = $1 LIMIT 1;