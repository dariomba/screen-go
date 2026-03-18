package processor

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/dariomba/screen-go/internal/domain"
	"github.com/dariomba/screen-go/internal/logger"
	"github.com/dariomba/screen-go/internal/ports"
)

type JobProcessorConfig struct {
	MaxThreads int
}

type jobWithContext struct {
	job *domain.Job
	ctx context.Context
}

type JobProcessor struct {
	jobRepository        ports.JobRepository
	screenshotRepository ports.ScreenshotRepository
	chromeDriver         ports.ChromeDriver
	uuidGenerator        ports.UUIDGenerator
	config               JobProcessorConfig

	jobs chan *jobWithContext

	ctx    context.Context
	cancel context.CancelFunc
}

func NewJobProcessor(
	chromeDriver ports.ChromeDriver,
	jobRepository ports.JobRepository,
	screenshotRepository ports.ScreenshotRepository,
	uuidGenerator ports.UUIDGenerator,
	config JobProcessorConfig,
) *JobProcessor {
	jobs := make(chan *jobWithContext, config.MaxThreads)

	ctx, cancel := context.WithCancel(context.Background())

	jp := &JobProcessor{
		jobRepository:        jobRepository,
		screenshotRepository: screenshotRepository,
		uuidGenerator:        uuidGenerator,
		chromeDriver:         chromeDriver,
		config:               config,
		jobs:                 jobs,
		ctx:                  ctx,
		cancel:               cancel,
	}

	jp.startWorkers()

	return jp
}

func (jp *JobProcessor) Process(ctx context.Context, job *domain.Job) {
	// Simply add the job to the channel for processing by workers
	ctx = logger.WithJobID(ctx, job.ID)
	logger.Ctx(ctx).Info().
		Str("url", job.URL).
		Msg("received new job for processing")
	jp.jobs <- &jobWithContext{
		job: job,
		ctx: ctx,
	}
}

func (jp *JobProcessor) startWorkers() {
	for i := 0; i < jp.config.MaxThreads; i++ {
		go func() {
			for {
				select {
				case <-jp.ctx.Done():
					return
				case jobCtx, ok := <-jp.jobs:
					if !ok {
						return
					}
					jp.processJob(jobCtx.ctx, jobCtx.job)
				}
			}
		}()
	}
}

func (jp *JobProcessor) processJob(ctx context.Context, job *domain.Job) {
	var jobErr error
	defer func() {
		if jobErr != nil {
			_ = jp.markJobAsFailed(ctx, job, jobErr.Error())
		}
	}()

	err := jp.jobRepository.UpdateJobToProcessing(ctx, job.ID)
	if err != nil {
		jobErr = errors.New("failed to update job status to processing")

		logger.Ctx(ctx).Error().
			Str("url", job.URL).
			Err(err).
			Msg("failed to update job status to processing")
		return
	}

	logger.Ctx(ctx).Info().
		Str("url", job.URL).
		Msg("started processing job")

	_, err = url.ParseRequestURI(job.URL)
	if err != nil {
		jobErr = fmt.Errorf("invalid URL: %v", err)

		logger.Ctx(ctx).Error().
			Str("url", job.URL).
			Err(err).
			Msg("failed to parse job URL")
		return
	}

	imgRes, err := jp.chromeDriver.CaptureScreenshot(ctx, job)
	if err != nil {
		jobErr = fmt.Errorf("failed to capture screenshot: %v", err)

		logger.Ctx(ctx).Error().
			Str("url", job.URL).
			Err(err).
			Msg("failed to capture screenshot")
		return
	}

	_, err = jp.screenshotRepository.CreateScreenshot(ctx, &domain.Screenshot{
		ID:          jp.uuidGenerator.Generate(),
		JobID:       job.ID,
		StorageKey:  "screenshot/" + job.ID + "." + string(job.Format), // This would be returned by the storage service
		ContentType: "image/png",                                       // This would be determined by the storage service
		Size:        int64(len(imgRes)),
	})
	if err != nil {
		jobErr = errors.New("failed to save screenshot information to repository")

		logger.Ctx(ctx).Error().
			Str("url", job.URL).
			Err(err).
			Msg("failed to save screenshot information to repository")
		return
	}

	jobErr = nil
	err = jp.jobRepository.UpdateJobToCompleted(ctx, job.ID)
	if err != nil {
		jobErr = errors.New("failed to update job status to completed")

		logger.Ctx(ctx).Error().
			Str("url", job.URL).
			Err(err).
			Msg("failed to update job status to completed")
		return
	}

	logger.Ctx(ctx).Info().
		Str("url", job.URL).
		Msg("finished processing job")
}

func (jp *JobProcessor) markJobAsFailed(ctx context.Context, job *domain.Job, errorMsg string) error {
	err := jp.jobRepository.UpdateJobToFailed(ctx, job.ID, errorMsg)
	if err != nil {
		logger.Ctx(ctx).Error().
			Str("url", job.URL).
			Err(err).
			Msg("failed to update job status to failed")
		return err
	}
	return nil
}

func (jp *JobProcessor) Close() error {
	close(jp.jobs)
	return nil
}
