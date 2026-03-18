-- 000002_create_screenshots.up

CREATE TABLE screenshots (
    id                  TEXT            PRIMARY KEY,
    job_id              TEXT NOT NULL   REFERENCES jobs(id),
    storage_key         TEXT            NOT NULL,
    content_type        TEXT            NOT NULL,
    size_bytes          BIGINT          NOT NULL,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    UNIQUE(job_id)
);