package ports

import (
	"context"

	"github.com/dariomba/screen-go/internal/domain"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error)
}
