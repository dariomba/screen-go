-- 000001_create_jobs.down.sql

DROP TRIGGER IF EXISTS update_jobs_updated_at ON jobs;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS jobs;
DROP TYPE IF EXISTS job_status;
DROP TYPE IF EXISTS job_format;