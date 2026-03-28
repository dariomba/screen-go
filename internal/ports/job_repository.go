package ports

import (
	"context"

	"github.com/dariomba/screen-go/internal/domain"
)

//go:generate go tool go.uber.org/mock/mockgen -source=job_repository.go -destination=../mocks/job_repository_mock.go -package=mocks
type JobRepository interface {
	GetJobByID(ctx context.Context, jobID string) (*domain.Job, error)
	CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error)
	UpdateJobToProcessing(ctx context.Context, jobID string) error
	UpdateJobToCompleted(ctx context.Context, jobID string) error
	UpdateJobToFailed(ctx context.Context, jobID string, errorMessage string) error
}
