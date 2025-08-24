
-- name: CreateLink :one
INSERT INTO links (
    alias,
    original_url,
    user_id 
) VALUES (
    $1, $2, $3 
)
RETURNING *;

-- name: GetLinkByAlias :one
SELECT * FROM links
WHERE alias = $1 LIMIT 1;