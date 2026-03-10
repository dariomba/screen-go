package v1

import (
	"context"

	"github.com/dariomba/screen-go/internal/model"
	"github.com/dariomba/screen-go/internal/openapi"
	"github.com/dariomba/screen-go/internal/postgres"
	"github.com/dariomba/screen-go/internal/processor"
	"github.com/dariomba/screen-go/internal/uuid"
)

type JobQuerier interface {
	CreateJob(ctx context.Context, arg postgres.CreateJobParams) (postgres.Job, error)
}

type CreateJobConfig struct {
	StatusEndpoint string
}

type CreateJob struct {
	uuidGenerator uuid.UUIDGenerator
	jobProcessor  processor.Processor
	jobQuerier    JobQuerier
	config        CreateJobConfig
}

func NewCreateJob(jobQuerier JobQuerier, jobProcessor processor.Processor, uuidGenerator uuid.UUIDGenerator, config CreateJobConfig) *CreateJob {
	return &CreateJob{
		jobQuerier:    jobQuerier,
		jobProcessor:  jobProcessor,
		uuidGenerator: uuidGenerator,
		config:        config,
	}
}

func (uc *CreateJob) Execute(ctx context.Context, request openapi.CreateJobRequestObject) (openapi.CreateJobResponseObject, error) {
	params := toCreateJobParams(uc.uuidGenerator.Generate(), request.Body)

	job, err := uc.jobQuerier.CreateJob(ctx, params)
	if err != nil {
		return openapi.CreateJob500JSONResponse{}, err
	}

	go func() {
		uc.jobProcessor.Process(context.Background(), &model.Job{
			ID:       job.ID,
			URL:      job.Url,
			Format:   model.JobFormat(job.Format),
			Width:    job.Width,
			Height:   job.Height,
			FullPage: job.FullPage,
		})
	}()

	return openapi.CreateJob202JSONResponse{
		JobID:     job.ID,
		Status:    openapi.CreateJobResponseStatusPending,
		StatusURL: uc.config.StatusEndpoint + job.ID,
	}, nil
}

var _ openapi.CreateJob = (*CreateJob)(nil)
