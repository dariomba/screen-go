package ports

import (
	"context"

	"github.com/dariomba/screen-go/internal/domain"
)

//go:generate go tool go.uber.org/mock/mockgen -source=chrome_driver.go -destination=../mocks/chrome_driver_mock.go -package=mocks
type ChromeDriver interface {
	CaptureScreenshot(ctx context.Context, job *domain.Job) ([]byte, error)
}
