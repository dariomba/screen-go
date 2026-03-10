package processor

import (
	"context"

	"github.com/dariomba/screen-go/internal/model"
)

type Processor interface {
	Process(ctx context.Context, job *model.Job)
	Close() error
}
