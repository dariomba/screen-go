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
    $1,
    $2,
    COALESCE(sqlc.narg('format')::job_format, 'png'),
    COALESCE(sqlc.narg('width')::integer, 1280),
    COALESCE(sqlc.narg('height')::integer, 800),
    COALESCE(sqlc.narg('full_page')::boolean, false),
    'pending'
)
RETURNING *;