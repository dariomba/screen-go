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

-- name: UpdateJobToProcessing :exec
UPDATE jobs
SET 
    status = 'processing',
    started_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND status = 'pending';

-- name: UpdateJobToDone :exec
UPDATE jobs
SET 
    status = 'done',
    finished_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND status = 'processing';

-- name: UpdateJobToFailed :exec
UPDATE jobs
SET 
    status = 'failed',
    finished_at = NOW(),
    error = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: GetJobByID :one
SELECT * FROM jobs
WHERE id = $1;