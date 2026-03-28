package domain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJob_Validate(t *testing.T) {
	t.Parallel()

	validJob := &Job{
		ID:     "job-123",
		URL:    "https://example.com",
		Format: JobFormatPng,
		Width:  DefaultJobWidth,
		Height: DefaultJobHeight,
		Status: JobStatusPending,
	}

	tests := []struct {
		name    string
		job     *Job
		wantErr *JobInvalidError
	}{
		{
			name: "WhenJobIsValid_ThenNoErrorIsReturned",
			job:  validJob,
		},
		{
			name: "WhenIDIsMissing_ThenValidationFails",
			job: &Job{
				URL:    "https://example.com",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: DefaultJobHeight,
			},
			wantErr: &JobInvalidError{Message: "job ID is required"},
		},
		{
			name: "WhenURLIsMissing_ThenValidationFails",
			job: &Job{
				ID:     "job-123",
				URL:    "",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: DefaultJobHeight,
			},
			wantErr: &JobInvalidError{Message: "job URL is required"},
		},
		{
			name: "WhenURLIsInvalid_ThenValidationFails",
			job: &Job{
				ID:     "job-123",
				URL:    "not-a-url",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: DefaultJobHeight,
			},
			wantErr: &JobInvalidError{Message: "invalid URL: not-a-url"},
		},
		{
			name: "WhenURLHasNoScheme_ThenValidationFails",
			job: &Job{
				ID:     "job-123",
				URL:    "example.com",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: DefaultJobHeight,
			},
			wantErr: &JobInvalidError{Message: "invalid URL: example.com"},
		},
		{
			name: "WhenURLHasNoHost_ThenValidationFails",
			job: &Job{
				ID:     "job-123",
				URL:    "https://",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: DefaultJobHeight,
			},
			wantErr: &JobInvalidError{Message: "invalid URL: https://"},
		},
		{
			name: "WhenWidthIsBelowMinimum_ThenValidationFails",
			job: &Job{
				ID:     "job-123",
				URL:    "https://example.com",
				Format: JobFormatPng,
				Width:  MinJobWidth - 1,
				Height: DefaultJobHeight,
			},
			wantErr: &JobInvalidError{Message: fmt.Sprintf("width must be between %d and %d, got %d", MinJobWidth, MaxJobWidth, MinJobWidth-1)},
		},
		{
			name: "WhenWidthIsAboveMaximum_ThenValidationFails",
			job: &Job{
				ID:     "job-123",
				URL:    "https://example.com",
				Format: JobFormatPng,
				Width:  MaxJobWidth + 1,
				Height: DefaultJobHeight,
			},
			wantErr: &JobInvalidError{Message: fmt.Sprintf("width must be between %d and %d, got %d", MinJobWidth, MaxJobWidth, MaxJobWidth+1)},
		},
		{
			name: "WhenWidthIsAtMinimum_ThenValidationSucceeds",
			job: &Job{
				ID:     "job-123",
				URL:    "https://example.com",
				Format: JobFormatPng,
				Width:  MinJobWidth,
				Height: DefaultJobHeight,
			},
		},
		{
			name: "WhenWidthIsAtMaximum_ThenValidationSucceeds",
			job: &Job{
				ID:     "job-123",
				URL:    "https://example.com",
				Format: JobFormatPng,
				Width:  MaxJobWidth,
				Height: DefaultJobHeight,
			},
		},
		{
			name: "WhenHeightIsBelowMinimum_ThenValidationFails",
			job: &Job{
				ID:     "job-123",
				URL:    "https://example.com",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: MinJobHeight - 1,
			},
			wantErr: &JobInvalidError{Message: fmt.Sprintf("height must be between %d and %d, got %d", MinJobHeight, MaxJobHeight, MinJobHeight-1)},
		},
		{
			name: "WhenHeightIsAboveMaximum_ThenValidationFails",
			job: &Job{
				ID:     "job-123",
				URL:    "https://example.com",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: MaxJobHeight + 1,
			},
			wantErr: &JobInvalidError{Message: fmt.Sprintf("height must be between %d and %d, got %d", MinJobHeight, MaxJobHeight, MaxJobHeight+1)},
		},
		{
			name: "WhenHeightIsAtMinimum_ThenValidationSucceeds",
			job: &Job{
				ID:     "job-123",
				URL:    "https://example.com",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: MinJobHeight,
			},
		},
		{
			name: "WhenHeightIsAtMaximum_ThenValidationSucceeds",
			job: &Job{
				ID:     "job-123",
				URL:    "https://example.com",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: MaxJobHeight,
			},
		},
		{
			name: "WhenURLHasPath_ThenValidationSucceeds",
			job: &Job{
				ID:     "job-123",
				URL:    "https://example.com/path/to/page",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: DefaultJobHeight,
			},
		},
		{
			name: "WhenURLHasQueryParams_ThenValidationSucceeds",
			job: &Job{
				ID:     "job-123",
				URL:    "https://example.com?param=value",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: DefaultJobHeight,
			},
		},
		{
			name: "WhenURLHasFragment_ThenValidationSucceeds",
			job: &Job{
				ID:     "job-123",
				URL:    "https://example.com?#section",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: DefaultJobHeight,
			},
		},
		{
			name: "WhenURLIsHTTP_ThenValidationSucceeds",
			job: &Job{
				ID:     "job-123",
				URL:    "http://example.com",
				Format: JobFormatPng,
				Width:  DefaultJobWidth,
				Height: DefaultJobHeight,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.job.Validate()

			if tt.wantErr != nil {
				require.Error(t, err)
				var jobErr *JobInvalidError
				assert.ErrorAs(t, err, &jobErr)
				assert.Equal(t, tt.wantErr.Message, jobErr.Message)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
