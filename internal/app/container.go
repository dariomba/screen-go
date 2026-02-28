package app

import (
	"context"
	"log"
	"net/http"

	"github.com/dariomba/screen-go/internal/openapi"
	oapiv1 "github.com/dariomba/screen-go/internal/openapi/v1"
	"github.com/dariomba/screen-go/internal/postgres"
	"github.com/dariomba/screen-go/internal/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

type params struct {
	HttpHost string
	HttpPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	StatusPollingEndpoint string
}

type services struct {
	httpServer                     *http.Server
	oapiHandler                    http.Handler
	oapiRequestValidatorMiddleware openapi.MiddlewareFunc
	database                       *pgxpool.Pool
	query                          *postgres.Queries
	uuidGenerator                  uuid.UUIDGenerator
	createJobUseCase               openapi.CreateJob
	getJobStatusUseCase            openapi.GetJobStatus
	getScreenshotUseCase           openapi.GetScreenshot
}

type Container struct {
	params
	services
}

func NewContainer() *Container {
	return &Container{
		params: params{
			StatusPollingEndpoint: "/v1/job/",
		},
	}
}

func (ctr *Container) HTTPServer() *http.Server {
	if ctr.httpServer == nil {
		ctr.httpServer = &http.Server{
			Addr:    ctr.HttpHost + ":" + ctr.HttpPort,
			Handler: ctr.OAPIHandler(),
		}
	}
	return ctr.httpServer
}

func (ctr *Container) OAPIHandler() http.Handler {
	if ctr.oapiHandler == nil {
		middlewares := []openapi.MiddlewareFunc{
			ctr.OAPIRequestValidatorMiddleware(),
		}

		ctr.oapiHandler = openapi.HandlerWithOptions(
			openapi.NewStrictHandlerWithOptions(
				openapi.NewServer(
					ctr.CreateJobUseCase(),
					ctr.GetJobStatusUseCase(),
					ctr.GetScreenshotUseCase(),
				),
				nil, // Strict middlewares
				openapi.StrictHTTPServerOptions{
					RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
						log.Printf("Request error: %v", err)

						openapi.WriteErrorJSON(w, "invalid request", http.StatusBadRequest)
					},
					ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
						log.Printf("Response error: %v", err)

						openapi.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
					},
				},
			),
			openapi.StdHTTPServerOptions{
				Middlewares: middlewares,
			},
		)
	}
	return ctr.oapiHandler
}

func (ctr *Container) OAPIRequestValidatorMiddleware() openapi.MiddlewareFunc {
	if ctr.oapiRequestValidatorMiddleware == nil {
		openApiSwagger, err := openapi.GetSwagger()
		if err != nil {
			panic(err)
		}

		openApiSwagger.Servers = nil

		ctr.oapiRequestValidatorMiddleware = nethttpmiddleware.OapiRequestValidatorWithOptions(openApiSwagger, &nethttpmiddleware.Options{
			ErrorHandler: func(w http.ResponseWriter, message string, statusCode int) {
				log.Printf("Request validation error: %s", message)

				openapi.WriteErrorJSON(w, message, http.StatusBadRequest)
			},
		})
	}
	return ctr.oapiRequestValidatorMiddleware
}

func (ctr *Container) Database() *pgxpool.Pool {
	if ctr.database == nil {
		connStr := "postgres://" + ctr.DBUser + ":" + ctr.DBPassword + "@" + ctr.DBHost + ":" + ctr.DBPort + "/" + ctr.DBName
		conn, err := pgxpool.New(context.Background(), connStr)
		if err != nil {
			panic(err)
		}
		ctr.database = conn
	}
	return ctr.database
}

func (ctr *Container) Query() *postgres.Queries {
	if ctr.query == nil {
		ctr.query = postgres.New(ctr.Database())
	}
	return ctr.query
}

func (ctr *Container) UUIDGenerator() uuid.UUIDGenerator {
	if ctr.uuidGenerator == nil {
		ctr.uuidGenerator = uuid.NewUlidGenerator()
	}
	return ctr.uuidGenerator
}

func (ctr *Container) CreateJobUseCase() openapi.CreateJob {
	if ctr.createJobUseCase == nil {
		ctr.createJobUseCase = oapiv1.NewCreateJob(
			ctr.Query(),
			ctr.UUIDGenerator(),
			oapiv1.CreateJobConfig{
				StatusEndpoint: ctr.StatusPollingEndpoint,
			})
	}
	return ctr.createJobUseCase
}

func (ctr *Container) GetJobStatusUseCase() openapi.GetJobStatus {
	if ctr.getJobStatusUseCase == nil {
		ctr.getJobStatusUseCase = oapiv1.NewGetJobStatus()
	}
	return ctr.getJobStatusUseCase
}

func (ctr *Container) GetScreenshotUseCase() openapi.GetScreenshot {
	if ctr.getScreenshotUseCase == nil {
		ctr.getScreenshotUseCase = oapiv1.NewGetScreenshot()
	}
	return ctr.getScreenshotUseCase
}
