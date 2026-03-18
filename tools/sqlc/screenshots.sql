-- name: CreateScreenshot :one
INSERT INTO screenshots (
    id,
    job_id,
    storage_key,
    content_type,
    size_bytes
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;