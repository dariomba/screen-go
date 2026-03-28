package ports

import (
	"context"

	"github.com/dariomba/screen-go/internal/domain"
)

//go:generate go tool go.uber.org/mock/mockgen -source=job_processor.go -destination=../mocks/job_processor_mock.go -package=mocks
type JobProcessor interface {
	Process(ctx context.Context, job *domain.Job)
	Close() error
}
