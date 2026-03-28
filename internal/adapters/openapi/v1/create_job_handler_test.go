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

type mocksCreateJobHandler struct {
	uuidGenerator *mocks.MockUUIDGenerator
	jobProcessor  *mocks.MockJobProcessor
	jobRepository *mocks.MockJobRepository
}

func TestCreateJobHandler(t *testing.T) {
	t.Parallel()

	validURL := "https://example.com"
	format := openapi.CreateJobRequestFormatPng
	defaultWidth := domain.MaxJobWidth
	defaultHeight := domain.MaxJobHeight

	tests := []struct {
		name    string
		request openapi.CreateJobRequestObject
		want    openapi.CreateJobResponseObject
		wantErr bool
		mocks   func(*mocksCreateJobHandler)
	}{
		{
			name: "WhenValidationFails_Then422IsReturned",
			request: openapi.CreateJobRequestObject{
				Body: &openapi.CreateJobJSONRequestBody{
					URL: "",
				},
			},
			want: openapi.CreateJob422JSONResponse{
				Error: "job URL is required",
			},
			mocks: func(m *mocksCreateJobHandler) {
				m.uuidGenerator.EXPECT().
					Generate().
					Return("job-123")
			},
		},
		{
			name: "WhenRepositoryFails_Then500IsReturned",
			request: openapi.CreateJobRequestObject{
				Body: &openapi.CreateJobJSONRequestBody{
					URL:    validURL,
					Format: &format,
					Width:  &defaultWidth,
					Height: &defaultHeight,
				},
			},
			want:    openapi.CreateJob500JSONResponse{},
			wantErr: true,
			mocks: func(m *mocksCreateJobHandler) {
				m.uuidGenerator.EXPECT().
					Generate().
					Return("job-123")

				m.jobRepository.EXPECT().
					CreateJob(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
		},
		{
			name: "WhenCreateJobSucceeds_Then202IsReturned",
			request: openapi.CreateJobRequestObject{
				Body: &openapi.CreateJobJSONRequestBody{
					URL:    validURL,
					Format: &format,
					Width:  &defaultWidth,
					Height: &defaultHeight,
				},
			},
			want: openapi.CreateJob202JSONResponse{
				JobID:     "job-123",
				Status:    openapi.CreateJobResponseStatusPending,
				StatusURL: "/v1/job/job-123",
			},
			mocks: func(m *mocksCreateJobHandler) {
				m.uuidGenerator.EXPECT().
					Generate().
					Return("job-123")

				m.jobRepository.EXPECT().
					CreateJob(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, job *domain.Job) (*domain.Job, error) {
						return job, nil
					})

				m.jobProcessor.EXPECT().
					Process(gomock.Any(), gomock.Any()).
					Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocksCreateJobHandler{
				uuidGenerator: mocks.NewMockUUIDGenerator(ctrl),
				jobProcessor:  mocks.NewMockJobProcessor(ctrl),
				jobRepository: mocks.NewMockJobRepository(ctrl),
			}

			tt.mocks(&m)

			uc := usecase.NewCreateJob(
				m.jobRepository,
				m.jobProcessor,
				m.uuidGenerator,
				usecase.CreateJobConfig{},
			)

			handler := NewCreateJobHandler(
				uc,
				CreateJobConfig{
					StatusEndpoint: "/v1/job/",
				},
			)

			got, err := handler.Execute(context.Background(), tt.request)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.want, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			time.Sleep(10 * time.Millisecond)
		})
	}
}
