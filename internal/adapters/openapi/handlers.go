package openapi

import "context"

type CreateJob interface {
	Execute(ctx context.Context, request CreateJobRequestObject) (CreateJobResponseObject, error)
}

type GetJobStatus interface {
	Execute(ctx context.Context, request GetJobStatusRequestObject) (GetJobStatusResponseObject, error)
}

type GetScreenshot interface {
	Execute(ctx context.Context, request GetScreenshotRequestObject) (GetScreenshotResponseObject, error)
}
