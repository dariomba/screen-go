package v1

import (
	"context"

	"github.com/dariomba/screen-go/internal/openapi"
)

type GetJobStatus struct {
}

func NewGetJobStatus() *GetJobStatus {
	return &GetJobStatus{}
}

func (uc *GetJobStatus) Execute(ctx context.Context, request openapi.GetJobStatusRequestObject) (openapi.GetJobStatusResponseObject, error) {
	// Implement the logic to get job status here
	// For example, you might want to validate the request, retrieve job status from the database, etc.

	// Return a dummy response for now
	response := openapi.GetJobStatus200JSONResponse{
		Status: "completed",
	}
	return response, nil
}

var _ openapi.GetJobStatus = (*GetJobStatus)(nil)
