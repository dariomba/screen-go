package ports

import (
	"context"

	"github.com/dariomba/screen-go/internal/domain"
)

type ChromeDriver interface {
	CaptureScreenshot(ctx context.Context, job *domain.Job) ([]byte, error)
}
