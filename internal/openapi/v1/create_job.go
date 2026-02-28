package v1

import (
	"context"
	"log"
	"time"

	"github.com/dariomba/screen-go/internal/openapi"
	"github.com/dariomba/screen-go/internal/postgres"
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
	jobQuerier    JobQuerier
	config        CreateJobConfig
}

func NewCreateJob(jobQuerier JobQuerier, uuidGenerator uuid.UUIDGenerator, config CreateJobConfig) *CreateJob {
	return &CreateJob{
		jobQuerier:    jobQuerier,
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

	go uc.processJob(job)

	return openapi.CreateJob202JSONResponse{
		JobID:     job.ID,
		Status:    openapi.CreateJobResponseStatusPending,
		StatusURL: uc.config.StatusEndpoint + job.ID,
	}, nil
}

func (uc *CreateJob) processJob(job postgres.Job) {
	log.Printf("Job %s is being processed...", job.ID)

	time.Sleep(2 * time.Second)

	log.Printf("Job %s processed!", job.ID)
}

var _ openapi.CreateJob = (*CreateJob)(nil)
