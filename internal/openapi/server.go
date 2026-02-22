package openapi

import "context"

type Server struct {
	CreateJobUseCase     CreateJob
	GetJobStatusUseCase  GetJobStatus
	GetScreenshotUseCase GetScreenshot
}

func NewServer(
	createJobUseCase CreateJob,
	getJobStatusUseCase GetJobStatus,
	getScreenshotUseCase GetScreenshot,
) *Server {
	return &Server{
		CreateJobUseCase:     createJobUseCase,
		GetJobStatusUseCase:  getJobStatusUseCase,
		GetScreenshotUseCase: getScreenshotUseCase,
	}
}

func (s *Server) CreateJob(ctx context.Context, request CreateJobRequestObject) (CreateJobResponseObject, error) {
	return s.CreateJobUseCase.Execute(ctx, request)
}

func (s *Server) GetJobStatus(ctx context.Context, request GetJobStatusRequestObject) (GetJobStatusResponseObject, error) {
	return s.GetJobStatusUseCase.Execute(ctx, request)
}

func (s *Server) GetScreenshot(ctx context.Context, request GetScreenshotRequestObject) (GetScreenshotResponseObject, error) {
	return s.GetScreenshotUseCase.Execute(ctx, request)
}

var _ StrictServerInterface = (*Server)(nil)
