package usecase

import (
	"context"

	"github.com/dariomba/screen-go/internal/domain"
	"github.com/dariomba/screen-go/internal/ports"
)

type CreateJobConfig struct {
	StatusEndpoint string
}

type CreateJob struct {
	uuidGenerator ports.UUIDGenerator
	jobProcessor  ports.JobProcessor
	jobRepository ports.JobRepository
	config        CreateJobConfig
}

func NewCreateJob(jobRepository ports.JobRepository, jobProcessor ports.JobProcessor, uuidGenerator ports.UUIDGenerator, config CreateJobConfig) *CreateJob {
	return &CreateJob{
		jobRepository: jobRepository,
		jobProcessor:  jobProcessor,
		uuidGenerator: uuidGenerator,
		config:        config,
	}
}

func (uc *CreateJob) Execute(ctx context.Context, job *domain.Job) (*domain.Job, error) {
	job.ID = uc.uuidGenerator.Generate()
	if err := job.Validate(); err != nil {
		return nil, err
	}

	job, err := uc.jobRepository.CreateJob(ctx, job)
	if err != nil {
		return nil, err
	}

	processingCtx := context.WithoutCancel(ctx)

	go func() {
		uc.jobProcessor.Process(processingCtx, job)
	}()

	return job, nil
}
