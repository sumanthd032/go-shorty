-- name: CreateClick :one
INSERT INTO clicks (
    link_id,
    ip_address,
    user_agent,
    referrer
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;