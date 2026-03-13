package ports

import (
	"context"

	"github.com/dariomba/screen-go/internal/domain"
)

type JobProcessor interface {
	Process(ctx context.Context, job *domain.Job)
	Close() error
}
