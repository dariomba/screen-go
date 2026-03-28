package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/dariomba/screen-go/internal/domain"
	"github.com/dariomba/screen-go/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mocksCreateJob struct {
	uuidGenerator *mocks.MockUUIDGenerator
	jobProcessor  *mocks.MockJobProcessor
	jobRepository *mocks.MockJobRepository
}

func TestCreateJob(t *testing.T) {
	t.Parallel()

	validJob := &domain.Job{
		URL:    "https://example.com",
		Format: domain.JobFormatPng,
		Width:  domain.MaxJobWidth,
		Height: domain.MaxJobHeight,
		Status: domain.JobStatusPending,
	}

	tests := []struct {
		name    string
		job     *domain.Job
		want    *domain.Job
		wantErr error
		mocks   func(*mocksCreateJob)
	}{
		{
			name:    "WhenJobValidationFails_ThenErrorIsReturned",
			job:     &domain.Job{},
			wantErr: &domain.JobInvalidError{Message: "job URL is required"},
			mocks: func(m *mocksCreateJob) {
				m.uuidGenerator.EXPECT().Generate().Return("job-123")
			},
		},
		{
			name: "WhenJobValidationFailsOnInvalidURL_ThenErrorIsReturned",
			job: &domain.Job{
				URL:    "not-a-url",
				Format: domain.JobFormatPng,
				Width:  domain.MinJobWidth,
				Height: domain.MaxJobHeight,
			},
			wantErr: &domain.JobInvalidError{Message: "invalid URL format: not-a-url"},
			mocks: func(m *mocksCreateJob) {
				m.uuidGenerator.EXPECT().Generate().Return("job-123")
			},
		},
		{
			name: "WhenJobValidationFailsOnInvalidWidth_ThenErrorIsReturned",
			job: &domain.Job{
				URL:    "https://example.com",
				Format: domain.JobFormatPng,
				Width:  domain.MinJobWidth - 1,
				Height: domain.MaxJobHeight - 1,
			},
			wantErr: &domain.JobInvalidError{Message: "width must be between 100 and 1920"},
			mocks: func(m *mocksCreateJob) {
				m.uuidGenerator.EXPECT().Generate().Return("job-123")
			},
		},
		{
			name:    "WhenCreateJobFails_ThenErrorIsReturned",
			job:     validJob,
			wantErr: assert.AnError,
			mocks: func(m *mocksCreateJob) {
				m.uuidGenerator.EXPECT().Generate().Return("job-123")
				m.jobRepository.EXPECT().
					CreateJob(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
		},
		{
			name: "WhenCreateJobSucceeds_ThenJobIsCreatedAndProcessorIsCalled",
			job:  validJob,
			want: validJob,
			mocks: func(m *mocksCreateJob) {
				m.uuidGenerator.EXPECT().Generate().Return("job-123")
				m.jobRepository.EXPECT().
					CreateJob(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, job *domain.Job) (*domain.Job, error) {
						assert.Equal(t, "job-123", job.ID)
						return validJob, nil
					})
				m.jobProcessor.EXPECT().
					Process(gomock.Any(), validJob).
					Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mocksCreateJob := mocksCreateJob{
				uuidGenerator: mocks.NewMockUUIDGenerator(ctrl),
				jobProcessor:  mocks.NewMockJobProcessor(ctrl),
				jobRepository: mocks.NewMockJobRepository(ctrl),
			}
			tt.mocks(&mocksCreateJob)

			uc := NewCreateJob(
				mocksCreateJob.jobRepository,
				mocksCreateJob.jobProcessor,
				mocksCreateJob.uuidGenerator,
				CreateJobConfig{StatusEndpoint: "/v1/job/"},
			)

			got, err := uc.Execute(context.Background(), tt.job)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)

				time.Sleep(10 * time.Millisecond)
			}
		})
	}
}
