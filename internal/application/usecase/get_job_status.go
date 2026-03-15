package usecase

import (
	"context"

	"github.com/dariomba/screen-go/internal/domain"
	"github.com/dariomba/screen-go/internal/ports"
)

type GetJobStatus struct {
	jobRepository ports.JobRepository
}

func NewGetJobStatus(jobRepository ports.JobRepository) *GetJobStatus {
	return &GetJobStatus{
		jobRepository: jobRepository,
	}
}

func (uc *GetJobStatus) Execute(ctx context.Context, jobID string) (*domain.Job, error) {
	job, err := uc.jobRepository.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	return job, nil
}
