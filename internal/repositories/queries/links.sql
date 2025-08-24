
-- name: CreateLink :one
INSERT INTO links (
    alias,
    original_url
) VALUES (
    $1, $2
)
RETURNING *;

-- name: GetLinkByAlias :one
SELECT * FROM links
WHERE alias = $1 LIMIT 1;