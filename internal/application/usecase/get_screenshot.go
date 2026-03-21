package usecase

import (
	"context"
	"io"

	"github.com/dariomba/screen-go/internal/ports"
)

type GetScreenshot struct {
	screenshotRepository ports.ScreenshotRepository
	screenshotStorage    ports.ScreenshotStorage
}

func NewGetScreenshot(screenshotRepository ports.ScreenshotRepository, screenshotStorage ports.ScreenshotStorage) *GetScreenshot {
	return &GetScreenshot{
		screenshotRepository: screenshotRepository,
		screenshotStorage:    screenshotStorage,
	}
}

type GetScreenshotResult struct {
	Data        io.Reader
	ContentType string
	Size        int64
}

func (uc *GetScreenshot) Execute(ctx context.Context, jobID string) (*GetScreenshotResult, error) {
	screenshot, err := uc.screenshotRepository.GetScreenshotByJobID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	data, err := uc.screenshotStorage.Get(ctx, screenshot.StorageKey)
	if err != nil {
		return nil, err
	}

	return &GetScreenshotResult{
		Data:        data,
		ContentType: screenshot.ContentType,
		Size:        screenshot.Size,
	}, nil
}
