package domain

import (
	"fmt"
	"net/url"
	"time"
)

const (
	MinJobWidth      = 320
	MaxJobWidth      = 3840
	DefaultJobWidth  = 1280
	MinJobHeight     = 240
	MaxJobHeight     = 2160
	DefaultJobHeight = 800
)

type JobFormat string

const (
	JobFormatPng JobFormat = "png"
	JobFormatPdf JobFormat = "pdf"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusDone       JobStatus = "done"
	JobStatusFailed     JobStatus = "failed"
)

type Job struct {
	ID         string
	URL        string
	Format     JobFormat
	Width      int
	Height     int
	FullPage   bool
	Status     JobStatus
	StartedAt  time.Time
	FinishedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type JobInvalidError struct {
	Message string
}

func (e *JobInvalidError) Error() string {
	return e.Message
}

func NewJob(url string, format *string, width, height *int, fullPage *bool) *Job {
	f := JobFormatPng
	w := DefaultJobWidth
	h := DefaultJobHeight
	fp := false

	if format != nil {
		f = JobFormat(*format)
	}
	if width != nil {
		w = *width
	}
	if height != nil {
		h = *height
	}
	if fullPage != nil {
		fp = *fullPage
	}

	return &Job{
		URL:      url,
		Format:   f,
		Width:    w,
		Height:   h,
		FullPage: fp,
		Status:   JobStatusPending,
	}
}

func (j *Job) Validate() error {
	if j.ID == "" {
		return &JobInvalidError{Message: "job ID is required"}
	}
	if j.URL == "" {
		return &JobInvalidError{Message: "job URL is required"}
	}
	if url, err := url.ParseRequestURI(j.URL); err != nil || url.Scheme == "" || url.Host == "" {
		return &JobInvalidError{Message: fmt.Sprintf("invalid URL: %s", j.URL)}
	}
	if j.Width < MinJobWidth || j.Width > MaxJobWidth {
		return &JobInvalidError{Message: fmt.Sprintf("width must be between %d and %d, got %d", MinJobWidth, MaxJobWidth, j.Width)}
	}
	if j.Height < MinJobHeight || j.Height > MaxJobHeight {
		return &JobInvalidError{Message: fmt.Sprintf("height must be between %d and %d, got %d", MinJobHeight, MaxJobHeight, j.Height)}
	}
	return nil
}
