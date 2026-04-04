package app

import (
	"context"
	"net/http"
	"time"

	"github.com/dariomba/screen-go/internal/adapters/chromedp"
	"github.com/dariomba/screen-go/internal/adapters/openapi"
	"github.com/dariomba/screen-go/internal/adapters/openapi/middleware"
	oapiv1 "github.com/dariomba/screen-go/internal/adapters/openapi/v1"
	"github.com/dariomba/screen-go/internal/adapters/postgres"
	"github.com/dariomba/screen-go/internal/adapters/postgres/sqlc"
	"github.com/dariomba/screen-go/internal/adapters/processor"
	"github.com/dariomba/screen-go/internal/adapters/storage"
	"github.com/dariomba/screen-go/internal/adapters/uuid"
	"github.com/dariomba/screen-go/internal/application/usecase"
	"github.com/dariomba/screen-go/internal/logger"
	"github.com/dariomba/screen-go/internal/ports"
	"github.com/getkin/kin-openapi/openapi3filter"
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

	APIKeys []string

	LogLevel  string
	LogPretty bool

	StatusPollingEndpoint string
	MaxProcessingThreads  int

	ChromeTimeout time.Duration
	ChromeWindowX int
	ChromeWindowY int

	ShutdownTimeout time.Duration
}

type services struct {
	httpServer                     *http.Server
	oapiHandler                    http.Handler
	oapiRequestValidatorMiddleware openapi.MiddlewareFunc
	database                       *pgxpool.Pool
	query                          *sqlc.Queries
	postgresJobRepository          *postgres.JobRepository
	postgresScreenshotRepository   *postgres.ScreenshotRepository
	uuidGenerator                  ports.UUIDGenerator
	jobProcessor                   ports.JobProcessor
	chromeDriver                   ports.ChromeDriver
	localStorage                   ports.ScreenshotStorage
	createJobUseCase               *usecase.CreateJob
	createJobHandler               openapi.CreateJob
	getJobStatusUseCase            *usecase.GetJobStatus
	getJobStatusHandler            openapi.GetJobStatus
	getScreenshotUseCase           *usecase.GetScreenshot
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
			LogLevel:              "info",
			LogPretty:             false,
			ShutdownTimeout:       30 * time.Second,
		},
	}
}

func (ctr *Container) HTTPServer() *http.Server {
	if ctr.httpServer == nil {
		handler := middleware.Recovery(
			middleware.RequestLogger(
				ctr.OAPIHandler(),
			),
		)

		ctr.httpServer = &http.Server{
			Addr:    ctr.HttpHost + ":" + ctr.HttpPort,
			Handler: handler,
		}
	}
	return ctr.httpServer
}

func (ctr *Container) OAPIHandler() http.Handler {
	if ctr.oapiHandler == nil {
		middlewares := []openapi.MiddlewareFunc{
			middleware.APIKeyAuthMiddleware(ctr.APIKeys),
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
						logger.Ctx(r.Context()).Error().
							Err(err).
							Str("error_type", "request_error").
							Msg("Request validation failed")

						openapi.WriteErrorJSON(w, "invalid request", http.StatusBadRequest)
					},
					ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
						logger.Ctx(r.Context()).Error().
							Err(err).
							Str("error_type", "response_error").
							Msg("Response generation failed")

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
			Options: openapi3filter.Options{
				AuthenticationFunc: openapi3filter.NoopAuthenticationFunc, // We handle authentication separately in our own middleware, so we can skip it here
			},
			ErrorHandler: func(w http.ResponseWriter, message string, statusCode int) {
				logger.Error().
					Str("error_type", "validation_error").
					Int("status", statusCode).
					Str("error_message", message).
					Msg("OpenAPI validation failed")

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
		ctr.jobProcessor = processor.NewJobProcessor(
			ctr.ChromeDriver(),
			ctr.PostgresJobRepository(),
			ctr.PostgresScreenshotRepository(),
			ctr.LocalStorage(),
			ctr.UUIDGenerator(),
			processor.JobProcessorConfig{
				MaxThreads: ctr.MaxProcessingThreads,
			})
	}
	return ctr.jobProcessor
}

func (ctr *Container) ChromeDriver() ports.ChromeDriver {
	if ctr.chromeDriver == nil {
		driver, err := chromedp.NewChromedp(&chromedp.ChromedpConfig{
			Timeout: ctr.ChromeTimeout,
			WindowX: ctr.ChromeWindowX,
			WindowY: ctr.ChromeWindowY,
		})
		if err != nil {
			panic(err)
		}
		ctr.chromeDriver = driver
	}
	return ctr.chromeDriver
}

func (ctr *Container) PostgresJobRepository() *postgres.JobRepository {
	if ctr.postgresJobRepository == nil {
		ctr.postgresJobRepository = postgres.NewJobRepository(ctr.Query())
	}
	return ctr.postgresJobRepository
}

func (ctr *Container) PostgresScreenshotRepository() *postgres.ScreenshotRepository {
	if ctr.postgresScreenshotRepository == nil {
		ctr.postgresScreenshotRepository = postgres.NewScreenshotRepository(ctr.Query())
	}
	return ctr.postgresScreenshotRepository
}

func (ctr *Container) LocalStorage() ports.ScreenshotStorage {
	if ctr.localStorage == nil {
		ctr.localStorage = storage.NewLocalStorage("/tmp/screen-go")
	}
	return ctr.localStorage
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

func (ctr *Container) GetJobStatusUseCase() *usecase.GetJobStatus {
	if ctr.getJobStatusUseCase == nil {
		ctr.getJobStatusUseCase = usecase.NewGetJobStatus(
			ctr.PostgresJobRepository(),
		)
	}
	return ctr.getJobStatusUseCase
}

func (ctr *Container) GetScreenshotUseCase() *usecase.GetScreenshot {
	if ctr.getScreenshotUseCase == nil {
		ctr.getScreenshotUseCase = usecase.NewGetScreenshot(
			ctr.PostgresScreenshotRepository(),
			ctr.LocalStorage(),
		)
	}
	return ctr.getScreenshotUseCase
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
		ctr.getJobStatusHandler = oapiv1.NewGetJobStatusHandler(ctr.GetJobStatusUseCase(), oapiv1.GetJobStatusHandlerConfig{
			ScreenshotEndpoint: "/v1/screenshot/",
		})
	}
	return ctr.getJobStatusHandler
}
func (ctr *Container) GetScreenshotHandler() openapi.GetScreenshot {
	if ctr.getScreenshotHandler == nil {
		ctr.getScreenshotHandler = oapiv1.NewGetScreenshotHandler(ctr.GetScreenshotUseCase())
	}
	return ctr.getScreenshotHandler
}

func (ctr *Container) Shutdown() {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), ctr.ShutdownTimeout)
	defer shutdownCancel()
	logger.Info().Msg("Shutdown signal received, stopping server...")
	if err := ctr.HTTPServer().Shutdown(context.Background()); err != nil {
		logger.Error().Err(err).Msg("Server shutdown failed")
	}
	logger.Info().Msg("Server gracefully stopped")

	if err := ctr.JobProcessor().Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Failed to shutdown job processor")
	}

	ctr.Database().Close()
}
