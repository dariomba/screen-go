package usecase

import (
	"context"
	"io"
	"testing"

	"github.com/dariomba/screen-go/internal/domain"
	"github.com/dariomba/screen-go/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mocksGetScreenshot struct {
	screenshotRepository *mocks.MockScreenshotRepository
	screenshotStorage    *mocks.MockScreenshotStorage
}

func TestGetScreenshot(t *testing.T) {
	t.Parallel()

	jobID := "job-123"
	validSc := &domain.Screenshot{
		ID:          "screenshot-id",
		JobID:       jobID,
		StorageKey:  "/path/to/file",
		ContentType: "image/png",
		Size:        300,
	}
	dataRes := io.NopCloser(nil)

	tests := []struct {
		name    string
		jobID   string
		want    *GetScreenshotResult
		wantErr error
		mocks   func(*mocksGetScreenshot)
	}{
		{
			name:    "WhenGetScreenshotByJobIDReturnsError_ThenErrorIsReturned",
			jobID:   jobID,
			wantErr: assert.AnError,
			mocks: func(m *mocksGetScreenshot) {
				m.screenshotRepository.EXPECT().GetScreenshotByJobID(gomock.Any(), jobID).Return(nil, assert.AnError)
			},
		},
		{
			name:    "WhenGetScreenshotByJobIDReturnsNotFound_ThenErrorIsReturned",
			jobID:   jobID,
			wantErr: domain.ErrScreenshotNotFound,
			mocks: func(m *mocksGetScreenshot) {
				m.screenshotRepository.EXPECT().GetScreenshotByJobID(gomock.Any(), jobID).Return(nil, domain.ErrScreenshotNotFound)
			},
		},
		{
			name:    "WhenGetScreenshotStorageReturnsError_ThenErrorIsReturned",
			jobID:   jobID,
			wantErr: assert.AnError,
			mocks: func(m *mocksGetScreenshot) {
				m.screenshotRepository.EXPECT().GetScreenshotByJobID(gomock.Any(), jobID).Return(validSc, nil)
				m.screenshotStorage.EXPECT().Get(gomock.Any(), validSc.StorageKey).Return(nil, assert.AnError)
			},
		},
		{
			name:  "WhenGetScreenshotByJobIDReturnsValidScreenshot_ThenScreenshotDataIsReturned",
			jobID: jobID,
			want: &GetScreenshotResult{
				Data:        dataRes,
				ContentType: validSc.ContentType,
				Size:        validSc.Size,
			},
			mocks: func(m *mocksGetScreenshot) {
				m.screenshotRepository.EXPECT().GetScreenshotByJobID(gomock.Any(), jobID).Return(validSc, nil)
				m.screenshotStorage.EXPECT().Get(gomock.Any(), validSc.StorageKey).Return(dataRes, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mocksGetScreenshot := mocksGetScreenshot{
				screenshotRepository: mocks.NewMockScreenshotRepository(ctrl),
				screenshotStorage:    mocks.NewMockScreenshotStorage(ctrl),
			}
			tt.mocks(&mocksGetScreenshot)

			uc := NewGetScreenshot(mocksGetScreenshot.screenshotRepository, mocksGetScreenshot.screenshotStorage)

			got, err := uc.Execute(context.Background(), tt.jobID)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
