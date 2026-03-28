package usecase

import (
	"context"
	"testing"

	"github.com/dariomba/screen-go/internal/domain"
	"github.com/dariomba/screen-go/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mocksGetJobStatus struct {
	jobRepository *mocks.MockJobRepository
}

func TestGetJobStatus(t *testing.T) {
	t.Parallel()

	validJob := &domain.Job{
		ID:     "job-123",
		URL:    "https://example.com",
		Status: domain.JobStatusPending,
		Width:  1280,
		Height: 800,
	}

	tests := []struct {
		name    string
		jobID   string
		want    *domain.Job
		wantErr error
		mocks   func(*mocksGetJobStatus)
	}{
		{
			name:    "WhenGetJobByIDReturnsError_ThenErrorIsReturned",
			jobID:   "jobID-123",
			wantErr: assert.AnError,
			mocks: func(m *mocksGetJobStatus) {
				m.jobRepository.EXPECT().GetJobByID(gomock.Any(), "jobID-123").Return(nil, assert.AnError)
			},
		},
		{
			name:    "WhenGetJobByIDReturnsNotFound_ThenErrorIsReturned",
			jobID:   "non-existent-job",
			wantErr: domain.ErrJobNotFound,
			mocks: func(m *mocksGetJobStatus) {
				m.jobRepository.EXPECT().GetJobByID(gomock.Any(), "non-existent-job").Return(nil, domain.ErrJobNotFound)
			},
		},
		{
			name:  "WhenGetJobByIDReturnsValidJob_ThenJobIsReturned",
			jobID: "jobID-123",
			want:  validJob,
			mocks: func(m *mocksGetJobStatus) {
				m.jobRepository.EXPECT().GetJobByID(gomock.Any(), "jobID-123").Return(validJob, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mocksGetJobStatus := mocksGetJobStatus{
				jobRepository: mocks.NewMockJobRepository(ctrl),
			}
			tt.mocks(&mocksGetJobStatus)

			uc := NewGetJobStatus(mocksGetJobStatus.jobRepository)

			got, err := uc.Execute(context.Background(), tt.jobID)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
