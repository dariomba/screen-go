package v1

import (
	"context"

	"github.com/dariomba/screen-go/internal/openapi"
)

type CreateJob struct {
}

func NewCreateJob() *CreateJob {
	return &CreateJob{}
}

func (uc *CreateJob) Execute(ctx context.Context, request openapi.CreateJobRequestObject) (openapi.CreateJobResponseObject, error) {
	// Implement the logic to create a job here
	// For example, you might want to validate the request, create a job in the database, etc.

	// Return a dummy response for now
	response := openapi.CreateJob202JSONResponse{
		JobID: "12345", // This should be generated dynamically
	}
	return response, nil
}

var _ openapi.CreateJob = (*CreateJob)(nil)
