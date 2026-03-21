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

type ScreenshotStorage interface {
	Save(ctx context.Context, input *SaveScreenshotInput) (*SaveScreenshotResult, error)
}
