-- 000001_create_jobs.up.sql

CREATE TYPE job_status AS ENUM ('pending', 'processing', 'done', 'failed');
CREATE TYPE job_format AS ENUM ('png', 'pdf');

CREATE TABLE jobs (
    id                  TEXT            PRIMARY KEY,
    url                 TEXT            NOT NULL,
    format              job_format      NOT NULL DEFAULT 'png',
    width               INTEGER         NOT NULL DEFAULT 1280,
    height              INTEGER         NOT NULL DEFAULT 800,
    full_page           BOOLEAN         NOT NULL DEFAULT FALSE,
    status              job_status      NOT NULL DEFAULT 'pending',
    memory_used_mb      INTEGER,
    started_at          TIMESTAMPTZ,
    finished_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Trigger to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_jobs_updated_at
    BEFORE UPDATE ON jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();