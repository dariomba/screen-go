package app

import (
	"context"
	"log"
	"net/http"

	"github.com/dariomba/screen-go/internal/adapters/openapi"
	oapiv1 "github.com/dariomba/screen-go/internal/adapters/openapi/v1"
	"github.com/dariomba/screen-go/internal/adapters/postgres"
	"github.com/dariomba/screen-go/internal/adapters/postgres/sqlc"
	"github.com/dariomba/screen-go/internal/adapters/processor"
	"github.com/dariomba/screen-go/internal/adapters/uuid"
	"github.com/dariomba/screen-go/internal/application/usecase"
	"github.com/dariomba/screen-go/internal/ports"
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
	MaxProcessingThreads  int
}

type services struct {
	httpServer                     *http.Server
	oapiHandler                    http.Handler
	oapiRequestValidatorMiddleware openapi.MiddlewareFunc
	database                       *pgxpool.Pool
	query                          *sqlc.Queries
	postgresJobRepository          *postgres.JobRepository
	uuidGenerator                  ports.UUIDGenerator
	jobProcessor                   ports.JobProcessor
	createJobUseCase               *usecase.CreateJob
	createJobHandler               openapi.CreateJob
	getJobStatusHandler            openapi.GetJobStatus
	getScreenshotHandler           openapi.GetScreenshot
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
					ctr.CreateJobHandler(),
					ctr.GetJobStatusHandler(),
					ctr.GetScreenshotHandler(),
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

func (ctr *Container) Query() *sqlc.Queries {
	if ctr.query == nil {
		ctr.query = sqlc.New(ctr.Database())
	}
	return ctr.query
}

func (ctr *Container) UUIDGenerator() ports.UUIDGenerator {
	if ctr.uuidGenerator == nil {
		ctr.uuidGenerator = uuid.NewUlidGenerator()
	}
	return ctr.uuidGenerator
}

func (ctr *Container) JobProcessor() ports.JobProcessor {
	if ctr.jobProcessor == nil {
		ctr.jobProcessor = processor.NewJobProcessor(processor.JobProcessorConfig{
			MaxThreads: ctr.MaxProcessingThreads,
		})
	}
	return ctr.jobProcessor
}

func (ctr *Container) PostgresJobRepository() *postgres.JobRepository {
	if ctr.postgresJobRepository == nil {
		ctr.postgresJobRepository = postgres.NewJobRepository(ctr.Query())
	}
	return ctr.postgresJobRepository
}

func (ctr *Container) CreateJobUseCase() *usecase.CreateJob {
	if ctr.createJobUseCase == nil {
		ctr.createJobUseCase = usecase.NewCreateJob(
			ctr.PostgresJobRepository(),
			ctr.JobProcessor(),
			ctr.UUIDGenerator(),
			usecase.CreateJobConfig{
				StatusEndpoint: ctr.StatusPollingEndpoint,
			},
		)
	}
	return ctr.createJobUseCase
}

func (ctr *Container) CreateJobHandler() openapi.CreateJob {
	if ctr.createJobHandler == nil {
		ctr.createJobHandler = oapiv1.NewCreateJobHandler(
			ctr.CreateJobUseCase(),
			oapiv1.CreateJobConfig{
				StatusEndpoint: ctr.StatusPollingEndpoint,
			})
	}
	return ctr.createJobHandler
}

func (ctr *Container) GetJobStatusHandler() openapi.GetJobStatus {
	if ctr.getJobStatusHandler == nil {
		ctr.getJobStatusHandler = oapiv1.NewGetJobStatusHandler()
	}
	return ctr.getJobStatusHandler
}
func (ctr *Container) GetScreenshotHandler() openapi.GetScreenshot {
	if ctr.getScreenshotHandler == nil {
		ctr.getScreenshotHandler = oapiv1.NewGetScreenshotHandler()
	}
	return ctr.getScreenshotHandler
}
