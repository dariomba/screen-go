package v1

import (
	"context"

	"github.com/dariomba/screen-go/internal/adapters/openapi"
)

type GetJobStatusHandler struct {
}

func NewGetJobStatusHandler() *GetJobStatusHandler {
	return &GetJobStatusHandler{}
}

func (uc *GetJobStatusHandler) Execute(ctx context.Context, request openapi.GetJobStatusRequestObject) (openapi.GetJobStatusResponseObject, error) {
	// Implement the logic to get job status here
	// For example, you might want to validate the request, retrieve job status from the database, etc.

	// Return a dummy response for now
	response := openapi.GetJobStatus200JSONResponse{
		Status: "completed",
	}
	return response, nil
}

var _ openapi.GetJobStatus = (*GetJobStatusHandler)(nil)
