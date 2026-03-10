package model

import (
	"time"
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
	Width      int32
	Height     int32
	FullPage   bool
	Status     JobStatus
	StartedAt  time.Time
	FinishedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
