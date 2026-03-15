package v1

import (
	"context"
	"errors"

	"github.com/dariomba/screen-go/internal/adapters/openapi"
	"github.com/dariomba/screen-go/internal/application/usecase"
	"github.com/dariomba/screen-go/internal/domain"
)

type GetJobStatusHandlerConfig struct {
	ScreenshotEndpoint string
}

type GetJobStatusHandler struct {
	getJobStatusUseCase *usecase.GetJobStatus

	config GetJobStatusHandlerConfig
}

func NewGetJobStatusHandler(getJobStatusUseCase *usecase.GetJobStatus, config GetJobStatusHandlerConfig) *GetJobStatusHandler {
	return &GetJobStatusHandler{
		getJobStatusUseCase: getJobStatusUseCase,
		config:              config,
	}
}

func (uc *GetJobStatusHandler) Execute(ctx context.Context, request openapi.GetJobStatusRequestObject) (openapi.GetJobStatusResponseObject, error) {
	job, err := uc.getJobStatusUseCase.Execute(ctx, request.ID)
	if err != nil {
		if errors.Is(err, domain.ErrJobNotFound) {
			return openapi.GetJobStatus404JSONResponse{
				Error: "job not found",
			}, nil
		}
		return nil, err
	}
	return openapi.GetJobStatus200JSONResponse{
		JobID:         job.ID,
		URL:           job.URL,
		Format:        openapi.JobFormat(job.Format),
		Width:         job.Width,
		Height:        job.Height,
		FullPage:      job.FullPage,
		Status:        openapi.JobStatus(job.Status),
		CreatedAt:     job.CreatedAt,
		StartedAt:     job.StartedAt,
		FinishedAt:    job.FinishedAt,
		Error:         job.Error,
		ScreenshotURL: uc.getScreenshotURL(job),
	}, nil
}

func (uc *GetJobStatusHandler) getScreenshotURL(job *domain.Job) *string {
	if job.Status == domain.JobStatusDone {
		url := uc.config.ScreenshotEndpoint + job.ID
		return &url
	}
	return nil
}

var _ openapi.GetJobStatus = (*GetJobStatusHandler)(nil)
