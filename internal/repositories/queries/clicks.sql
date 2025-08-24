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

-- name: GetLinkAnalytics :many
SELECT
    l.id,
    l.alias,
    l.original_url,
    COUNT(c.id) AS total_clicks
FROM
    links l
LEFT JOIN
    clicks c ON l.id = c.link_id
WHERE
    l.user_id = $1
GROUP BY
    l.id
ORDER BY
    total_clicks DESC;