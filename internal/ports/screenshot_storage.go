package ports

import (
	"context"
	"io"
)

type SaveScreenshotInput struct {
	Key         string
	Body        io.Reader
	ContentType string
}

type SaveScreenshotResult struct {
	Key  string
	Size int64
}

//go:generate go tool go.uber.org/mock/mockgen -source=screenshot_storage.go -destination=../mocks/screenshot_storage_mock.go -package=mocks
type ScreenshotStorage interface {
	Get(ctx context.Context, key string) (io.Reader, error)
	Save(ctx context.Context, input *SaveScreenshotInput) (*SaveScreenshotResult, error)
}
