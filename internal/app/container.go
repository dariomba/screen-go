package app

import (
	"net/http"

	"github.com/dariomba/screen-go/internal/openapi"
	oapiv1 "github.com/dariomba/screen-go/internal/openapi/v1"
)

type params struct {
	Addr string
}

type services struct {
	httpServer           *http.Server
	oapiHandler          http.Handler
	createJobUseCase     openapi.CreateJob
	getJobStatusUseCase  openapi.GetJobStatus
	getScreenshotUseCase openapi.GetScreenshot
}

type Container struct {
	params
	services
}

func NewContainer() *Container {
	return &Container{
		// Initialize shared dependencies here
	}
}

func (ctr *Container) HTTPServer() *http.Server {
	if ctr.services.httpServer == nil {
		ctr.services.httpServer = &http.Server{
			Addr:    ctr.params.Addr,
			Handler: ctr.OAPIHandler(),
		}
	}
	return ctr.services.httpServer
}

func (ctr *Container) OAPIHandler() http.Handler {
	if ctr.services.oapiHandler == nil {
		ctr.services.oapiHandler = openapi.HandlerFromMux(
			openapi.NewStrictHandler(
				openapi.NewServer(
					ctr.CreateJobUseCase(),
					ctr.GetJobStatusUseCase(),
					ctr.GetScreenshotUseCase(),
				),
				nil, // You can add middleware here if needed
			),
			http.NewServeMux(),
		)
	}
	return ctr.services.oapiHandler
}

func (ctr *Container) CreateJobUseCase() openapi.CreateJob {
	if ctr.services.createJobUseCase == nil {
		ctr.services.createJobUseCase = oapiv1.NewCreateJob()
	}
	return ctr.services.createJobUseCase
}

func (ctr *Container) GetJobStatusUseCase() openapi.GetJobStatus {
	if ctr.services.getJobStatusUseCase == nil {
		ctr.services.getJobStatusUseCase = oapiv1.NewGetJobStatus()
	}
	return ctr.services.getJobStatusUseCase
}

func (ctr *Container) GetScreenshotUseCase() openapi.GetScreenshot {
	if ctr.services.getScreenshotUseCase == nil {
		ctr.services.getScreenshotUseCase = oapiv1.NewGetScreenshot()
	}
	return ctr.services.getScreenshotUseCase
}
