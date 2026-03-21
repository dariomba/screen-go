package ports

import (
	"context"

	"github.com/dariomba/screen-go/internal/domain"
)

type ScreenshotRepository interface {
	CreateScreenshot(ctx context.Context, screenshot *domain.Screenshot) (*domain.Screenshot, error)
	GetScreenshotByJobID(ctx context.Context, jobID string) (*domain.Screenshot, error)
}
