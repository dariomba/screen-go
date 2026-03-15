package ports

import (
	"context"

	"github.com/dariomba/screen-go/internal/domain"
)

type JobRepository interface {
	GetJobByID(ctx context.Context, jobID string) (*domain.Job, error)
	CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error)
	UpdateJobToProcessing(ctx context.Context, jobID string) error
	UpdateJobToCompleted(ctx context.Context, jobID string) error
	UpdateJobToFailed(ctx context.Context, jobID string, errorMessage string) error
}
