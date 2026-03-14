package postgres

import (
	"context"

	"github.com/dariomba/screen-go/internal/adapters/postgres/sqlc"
	"github.com/dariomba/screen-go/internal/domain"
)

type JobRepository struct {
	queries *sqlc.Queries
}

func NewJobRepository(queries *sqlc.Queries) *JobRepository {
	return &JobRepository{
		queries: queries,
	}
}

func (r *JobRepository) CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error) {
	createdJob, err := r.queries.CreateJob(ctx, sqlc.CreateJobParams{
		ID:       job.ID,
		Url:      job.URL,
		Format:   toPgNullJobFormat(job.Format),
		Width:    toPgInt4(job.Width),
		Height:   toPgInt4(job.Height),
		FullPage: toPgBool(job.FullPage),
	})
	if err != nil {
		return nil, err
	}

	return &domain.Job{
		ID:        createdJob.ID,
		URL:       createdJob.Url,
		Format:    domain.JobFormat(createdJob.Format),
		Width:     int(createdJob.Width),
		Height:    int(createdJob.Height),
		FullPage:  createdJob.FullPage,
		Status:    domain.JobStatus(createdJob.Status),
		CreatedAt: createdJob.CreatedAt.Time,
		UpdatedAt: createdJob.UpdatedAt.Time,
	}, nil
}

func (r *JobRepository) UpdateJobToProcessing(ctx context.Context, jobID string) error {
	return r.queries.UpdateJobToProcessing(ctx, jobID)
}

func (r *JobRepository) UpdateJobToCompleted(ctx context.Context, jobID string) error {
	return r.queries.UpdateJobToDone(ctx, jobID)
}

func (r *JobRepository) UpdateJobToFailed(ctx context.Context, jobID string, errorMessage string) error {
	return r.queries.UpdateJobToFailed(ctx, sqlc.UpdateJobToFailedParams{
		ID:    jobID,
		Error: toPgText(errorMessage),
	})
}
