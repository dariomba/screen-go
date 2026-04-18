package v1

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/dariomba/screen-go/internal/adapters/openapi"
	"github.com/dariomba/screen-go/internal/application/usecase"
	"github.com/dariomba/screen-go/internal/domain"
	"github.com/dariomba/screen-go/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mocksGetScreenshotHandler struct {
	screenshotRepository *mocks.MockScreenshotRepository
	screenshotStorage    *mocks.MockScreenshotStorage
}

func TestGetScreenshotHandler(t *testing.T) {
	t.Parallel()

	jobID := "job-123"
	storageKey := "screenshots/job-123.png"
	data := []byte("image-bytes")

	tests := []struct {
		name    string
		request openapi.GetScreenshotRequestObject
		want    openapi.GetScreenshotResponseObject
		wantErr bool
		mocks   func(*mocksGetScreenshotHandler)
	}{
		{
			name: "WhenScreenshotNotFound_Then404IsReturned",
			request: openapi.GetScreenshotRequestObject{
				ID: jobID,
			},
			want: openapi.GetScreenshot404JSONResponse{
				Error: "screenshot not found",
			},
			mocks: func(m *mocksGetScreenshotHandler) {
				m.screenshotRepository.EXPECT().
					GetScreenshotByJobID(gomock.Any(), jobID).
					Return(nil, domain.ErrScreenshotNotFound)
			},
		},
		{
			name: "WhenRepositoryFails_ThenErrorIsReturned",
			request: openapi.GetScreenshotRequestObject{
				ID: jobID,
			},
			wantErr: true,
			mocks: func(m *mocksGetScreenshotHandler) {
				m.screenshotRepository.EXPECT().
					GetScreenshotByJobID(gomock.Any(), jobID).
					Return(nil, assert.AnError)
			},
		},
		{
			name: "WhenStorageFails_ThenErrorIsReturned",
			request: openapi.GetScreenshotRequestObject{
				ID: jobID,
			},
			wantErr: true,
			mocks: func(m *mocksGetScreenshotHandler) {
				m.screenshotRepository.EXPECT().
					GetScreenshotByJobID(gomock.Any(), jobID).
					Return(&domain.Screenshot{
						StorageKey:  storageKey,
						ContentType: "image/png",
						Size:        int64(len(data)),
					}, nil)

				m.screenshotStorage.EXPECT().
					Get(gomock.Any(), storageKey).
					Return(nil, assert.AnError)
			},
		},
		{
			name: "WhenPNG_ThenImageResponseIsReturned",
			request: openapi.GetScreenshotRequestObject{
				ID: jobID,
			},
			want: openapi.GetScreenshot200ImagePngResponse{
				Body:          bytes.NewReader(data),
				ContentLength: int64(len(data)),
				Headers: openapi.GetScreenshot200ResponseHeaders{
					ContentType:   "image/png",
					ContentLength: len(data),
				},
			},
			mocks: func(m *mocksGetScreenshotHandler) {
				m.screenshotRepository.EXPECT().
					GetScreenshotByJobID(gomock.Any(), jobID).
					Return(&domain.Screenshot{
						StorageKey:  storageKey,
						ContentType: "image/png",
						Size:        int64(len(data)),
					}, nil)

				m.screenshotStorage.EXPECT().
					Get(gomock.Any(), storageKey).
					Return(io.NopCloser(bytes.NewReader(data)), nil)
			},
		},
		{
			name: "WhenPDF_ThenPDFResponseIsReturned",
			request: openapi.GetScreenshotRequestObject{
				ID: jobID,
			},
			want: openapi.GetScreenshot200ApplicationPdfResponse{
				Body:          bytes.NewReader(data),
				ContentLength: int64(len(data)),
				Headers: openapi.GetScreenshot200ResponseHeaders{
					ContentType:   "application/pdf",
					ContentLength: len(data),
				},
			},
			mocks: func(m *mocksGetScreenshotHandler) {
				m.screenshotRepository.EXPECT().
					GetScreenshotByJobID(gomock.Any(), jobID).
					Return(&domain.Screenshot{
						StorageKey:  storageKey,
						ContentType: "application/pdf",
						Size:        int64(len(data)),
					}, nil)

				m.screenshotStorage.EXPECT().
					Get(gomock.Any(), storageKey).
					Return(io.NopCloser(bytes.NewReader(data)), nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocksGetScreenshotHandler{
				screenshotRepository: mocks.NewMockScreenshotRepository(ctrl),
				screenshotStorage:    mocks.NewMockScreenshotStorage(ctrl),
			}

			tt.mocks(&m)

			uc := usecase.NewGetScreenshot(
				m.screenshotRepository,
				m.screenshotStorage,
			)

			handler := NewGetScreenshotHandler(uc)

			got, err := handler.Execute(context.Background(), tt.request)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// special handling because io.Reader cannot be compared directly
			switch expected := tt.want.(type) {
			case openapi.GetScreenshot200ImagePngResponse:
				actual := got.(openapi.GetScreenshot200ImagePngResponse)
				assert.Equal(t, expected.ContentLength, actual.ContentLength)
				assert.Equal(t, expected.Headers, actual.Headers)
			case openapi.GetScreenshot200ApplicationPdfResponse:
				actual := got.(openapi.GetScreenshot200ApplicationPdfResponse)
				assert.Equal(t, expected.ContentLength, actual.ContentLength)
				assert.Equal(t, expected.Headers, actual.Headers)
			default:
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
