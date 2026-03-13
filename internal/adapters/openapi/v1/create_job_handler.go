package v1

import (
	"context"
	"errors"

	"github.com/dariomba/screen-go/internal/adapters/openapi"
	"github.com/dariomba/screen-go/internal/application/usecase"
	"github.com/dariomba/screen-go/internal/domain"
)

type CreateJobConfig struct {
	StatusEndpoint string
}

type CreateJobHandler struct {
	createJobUseCase *usecase.CreateJob
	config           CreateJobConfig
}

func NewCreateJobHandler(createJobUseCase *usecase.CreateJob, config CreateJobConfig) *CreateJobHandler {
	return &CreateJobHandler{
		createJobUseCase: createJobUseCase,
		config:           config,
	}
}

func (uc *CreateJobHandler) Execute(ctx context.Context, request openapi.CreateJobRequestObject) (openapi.CreateJobResponseObject, error) {
	var format *string
	if request.Body.Format != nil {
		f := string(*request.Body.Format)
		format = &f
	}

	job, err := uc.createJobUseCase.Execute(ctx, domain.NewJob(
		request.Body.URL,
		format,
		request.Body.Width,
		request.Body.Height,
		request.Body.FullPage,
	))
	if err != nil {
		if je, ok := errors.AsType[*domain.JobInvalidError](err); ok {
			return openapi.CreateJob400JSONResponse{}, je
		}

		return openapi.CreateJob500JSONResponse{}, err
	}

	return openapi.CreateJob202JSONResponse{
		JobID:     job.ID,
		Status:    openapi.CreateJobResponseStatusPending,
		StatusURL: uc.config.StatusEndpoint + job.ID,
	}, nil
}

var _ openapi.CreateJob = (*CreateJobHandler)(nil)
