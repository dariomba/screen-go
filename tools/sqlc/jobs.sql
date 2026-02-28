-- name: CreateJob :one
INSERT INTO jobs (
    id,
    url,
    format,
    width,
    height,
    full_page,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, 'pending'
)
RETURNING *;