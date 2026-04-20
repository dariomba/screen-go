package processor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

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
	screenshotStorage    ports.ScreenshotStorage
	chromeDriver         ports.ChromeDriver
	uuidGenerator        ports.UUIDGenerator
	config               JobProcessorConfig

	jobs chan *jobWithContext

	wg       sync.WaitGroup
	shutdown chan struct{}
	once     sync.Once
}

func NewJobProcessor(
	chromeDriver ports.ChromeDriver,
	jobRepository ports.JobRepository,
	screenshotRepository ports.ScreenshotRepository,
	screenshotStorage ports.ScreenshotStorage,
	uuidGenerator ports.UUIDGenerator,
	config JobProcessorConfig,
) *JobProcessor {
	jobs := make(chan *jobWithContext, config.MaxThreads)

	jp := &JobProcessor{
		jobRepository:        jobRepository,
		screenshotRepository: screenshotRepository,
		screenshotStorage:    screenshotStorage,
		uuidGenerator:        uuidGenerator,
		chromeDriver:         chromeDriver,
		config:               config,
		jobs:                 jobs,
		shutdown:             make(chan struct{}),
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
	select {
	case <-jp.shutdown:
		logger.Ctx(ctx).Warn().
			Str("url", job.URL).
			Msg("job rejected, processor is shutting down")
		jp.markJobAsFailed(ctx, job, "job processor is shutting down, please retry later") //nolint:errcheck
		return
	case jp.jobs <- &jobWithContext{job: job, ctx: ctx}:
	}
}

func (jp *JobProcessor) startWorkers() {
	for i := 0; i < jp.config.MaxThreads; i++ {
		jp.wg.Go(func() {
			for jobCtx := range jp.jobs {
				jp.processJob(jobCtx.ctx, jobCtx.job)
			}
		})
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

	storageKey := fmt.Sprintf("screenshot/%s.%s", job.ID, job.Format)
	contentType := contentTypeFromFormat(job.Format)
	saveStorageRes, err := jp.screenshotStorage.Save(ctx, &ports.SaveScreenshotInput{
		Key:         storageKey,
		Body:        bytes.NewReader(imgRes),
		ContentType: contentType,
	})
	if err != nil {
		jobErr = errors.New("failed to save screenshot to storage")

		logger.Ctx(ctx).Error().
			Str("url", job.URL).
			Err(err).
			Msg("failed to save screenshot to storage")
		return
	}

	_, err = jp.screenshotRepository.CreateScreenshot(ctx, &domain.Screenshot{
		ID:          jp.uuidGenerator.Generate(),
		JobID:       job.ID,
		StorageKey:  saveStorageRes.Key,
		ContentType: contentType,
		Size:        saveStorageRes.Size,
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

func contentTypeFromFormat(format domain.JobFormat) string {
	switch format {
	case domain.JobFormatPdf:
		return "application/pdf"
	case domain.JobFormatPng:
		return "image/png"
	default:
		return "application/octet-stream"
	}
}

func (jp *JobProcessor) Shutdown(ctx context.Context) error {
	jp.once.Do(func() {
		close(jp.shutdown)
		close(jp.jobs)
	})

	done := make(chan struct{})
	go func() {
		jp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info().Msg("all workers have completed, shutdown complete")

		return jp.chromeDriver.Shutdown(ctx)
	case <-ctx.Done():
		logger.Warn().Msg("shutdown timeout reached, forcing shutdown with active workers")

		jp.chromeDriver.Shutdown(ctx) //nolint:errcheck

		logger.Info().Msg("forced shutdown complete")
		return context.DeadlineExceeded
	}
}
