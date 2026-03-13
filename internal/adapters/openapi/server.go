package openapi

import "context"

type Server struct {
	CreateJobHandler     CreateJob
	GetJobStatusHandler  GetJobStatus
	GetScreenshotHandler GetScreenshot
}

func NewServer(
	createJobHandler CreateJob,
	getJobStatusHandler GetJobStatus,
	getScreenshotHandler GetScreenshot,
) *Server {
	return &Server{
		CreateJobHandler:     createJobHandler,
		GetJobStatusHandler:  getJobStatusHandler,
		GetScreenshotHandler: getScreenshotHandler,
	}
}

func (s *Server) CreateJob(ctx context.Context, request CreateJobRequestObject) (CreateJobResponseObject, error) {
	return s.CreateJobHandler.Execute(ctx, request)
}

func (s *Server) GetJobStatus(ctx context.Context, request GetJobStatusRequestObject) (GetJobStatusResponseObject, error) {
	return s.GetJobStatusHandler.Execute(ctx, request)
}

func (s *Server) GetScreenshot(ctx context.Context, request GetScreenshotRequestObject) (GetScreenshotResponseObject, error) {
	return s.GetScreenshotHandler.Execute(ctx, request)
}

var _ StrictServerInterface = (*Server)(nil)
