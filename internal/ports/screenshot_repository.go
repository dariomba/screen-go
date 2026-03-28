package ports

import (
	"context"

	"github.com/dariomba/screen-go/internal/domain"
)

//go:generate go tool go.uber.org/mock/mockgen -source=screenshot_repository.go -destination=../mocks/screenshot_repository_mock.go -package=mocks
type ScreenshotRepository interface {
	CreateScreenshot(ctx context.Context, screenshot *domain.Screenshot) (*domain.Screenshot, error)
	GetScreenshotByJobID(ctx context.Context, jobID string) (*domain.Screenshot, error)
}
