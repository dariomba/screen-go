package postgres

import (
	"context"

	"github.com/dariomba/screen-go/internal/adapters/postgres/sqlc"
	"github.com/dariomba/screen-go/internal/domain"
)

type ScreenshotRepository struct {
	queries *sqlc.Queries
}

func NewScreenshotRepository(queries *sqlc.Queries) *ScreenshotRepository {
	return &ScreenshotRepository{
		queries: queries,
	}
}

func (r *ScreenshotRepository) CreateScreenshot(ctx context.Context, screenshot *domain.Screenshot) (*domain.Screenshot, error) {
	createdScreenshot, err := r.queries.CreateScreenshot(ctx, sqlc.CreateScreenshotParams{
		ID:          screenshot.ID,
		JobID:       screenshot.JobID,
		StorageKey:  screenshot.StorageKey,
		ContentType: screenshot.ContentType,
		SizeBytes:   screenshot.Size,
	})
	if err != nil {
		return nil, err
	}

	return &domain.Screenshot{
		ID:          createdScreenshot.ID,
		JobID:       createdScreenshot.JobID,
		StorageKey:  createdScreenshot.StorageKey,
		ContentType: createdScreenshot.ContentType,
		Size:        createdScreenshot.SizeBytes,
		CreatedAt:   createdScreenshot.CreatedAt.Time,
	}, nil
}
