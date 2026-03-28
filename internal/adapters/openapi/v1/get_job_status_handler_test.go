package v1

import (
	"context"
	"testing"
	"time"

	"github.com/dariomba/screen-go/internal/adapters/openapi"
	"github.com/dariomba/screen-go/internal/application/usecase"
	"github.com/dariomba/screen-go/internal/domain"
	"github.com/dariomba/screen-go/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mocksGetJobStatusHandler struct {
	jobRepository *mocks.MockJobRepository
}

func TestGetJobStatusHandler(t *testing.T) {
	t.Parallel()

	now := time.Now()
	jobID := "job-123"

	baseJob := &domain.Job{
		ID:        jobID,
		URL:       "https://example.com",
		Format:    domain.JobFormatPng,
		Width:     800,
		Height:    600,
		FullPage:  true,
		Status:    domain.JobStatusPending,
		CreatedAt: now,
	}

	tests := []struct {
		name    string
		request openapi.GetJobStatusRequestObject
		want    openapi.GetJobStatusResponseObject
		wantErr bool
		mocks   func(*mocksGetJobStatusHandler)
	}{
		{
			name: "WhenJobNotFound_Then404IsReturned",
			request: openapi.GetJobStatusRequestObject{
				ID: jobID,
			},
			want: openapi.GetJobStatus404JSONResponse{
				Error: "job not found",
			},
			mocks: func(m *mocksGetJobStatusHandler) {
				m.jobRepository.EXPECT().
					GetJobByID(gomock.Any(), jobID).
					Return(nil, domain.ErrJobNotFound)
			},
		},
		{
			name: "WhenRepositoryFails_ThenErrorIsReturned",
			request: openapi.GetJobStatusRequestObject{
				ID: jobID,
			},
			wantErr: true,
			mocks: func(m *mocksGetJobStatusHandler) {
				m.jobRepository.EXPECT().
					GetJobByID(gomock.Any(), jobID).
					Return(nil, assert.AnError)
			},
		},
		{
			name: "WhenJobIsPending_Then200WithoutScreenshotURL",
			request: openapi.GetJobStatusRequestObject{
				ID: jobID,
			},
			want: openapi.GetJobStatus200JSONResponse{
				JobID:         baseJob.ID,
				URL:           baseJob.URL,
				Format:        openapi.JobFormat(baseJob.Format),
				Width:         baseJob.Width,
				Height:        baseJob.Height,
				FullPage:      baseJob.FullPage,
				Status:        openapi.JobStatus(baseJob.Status),
				CreatedAt:     baseJob.CreatedAt,
				StartedAt:     baseJob.StartedAt,
				FinishedAt:    baseJob.FinishedAt,
				Error:         baseJob.Error,
				ScreenshotURL: nil,
			},
			mocks: func(m *mocksGetJobStatusHandler) {
				m.jobRepository.EXPECT().
					GetJobByID(gomock.Any(), jobID).
					Return(baseJob, nil)
			},
		},
		{
			name: "WhenJobIsDone_Then200WithScreenshotURL",
			request: openapi.GetJobStatusRequestObject{
				ID: jobID,
			},
			want: func() openapi.GetJobStatusResponseObject {
				doneJob := *baseJob
				doneJob.Status = domain.JobStatusDone
				doneJob.FinishedAt = &now

				url := "/v1/screenshot/" + jobID

				return openapi.GetJobStatus200JSONResponse{
					JobID:         doneJob.ID,
					URL:           doneJob.URL,
					Format:        openapi.JobFormat(doneJob.Format),
					Width:         doneJob.Width,
					Height:        doneJob.Height,
					FullPage:      doneJob.FullPage,
					Status:        openapi.JobStatus(doneJob.Status),
					CreatedAt:     doneJob.CreatedAt,
					StartedAt:     doneJob.StartedAt,
					FinishedAt:    doneJob.FinishedAt,
					Error:         doneJob.Error,
					ScreenshotURL: &url,
				}
			}(),
			mocks: func(m *mocksGetJobStatusHandler) {
				doneJob := *baseJob
				doneJob.Status = domain.JobStatusDone
				doneJob.FinishedAt = &now

				m.jobRepository.EXPECT().
					GetJobByID(gomock.Any(), jobID).
					Return(&doneJob, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocksGetJobStatusHandler{
				jobRepository: mocks.NewMockJobRepository(ctrl),
			}

			tt.mocks(&m)

			uc := usecase.NewGetJobStatus(m.jobRepository)

			handler := NewGetJobStatusHandler(
				uc,
				GetJobStatusHandlerConfig{
					ScreenshotEndpoint: "/v1/screenshot/",
				},
			)

			got, err := handler.Execute(context.Background(), tt.request)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
