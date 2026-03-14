package processor

import (
	"context"
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
	jobRepository ports.JobRepository
	chromeDriver  ports.ChromeDriver
	config        JobProcessorConfig

	jobs chan *jobWithContext

	ctx    context.Context
	cancel context.CancelFunc
}

func NewJobProcessor(chromeDriver ports.ChromeDriver, jobRepository ports.JobRepository, config JobProcessorConfig) *JobProcessor {
	jobs := make(chan *jobWithContext, config.MaxThreads)

	ctx, cancel := context.WithCancel(context.Background())

	jp := &JobProcessor{
		jobRepository: jobRepository,
		chromeDriver:  chromeDriver,
		config:        config,
		jobs:          jobs,
		ctx:           ctx,
		cancel:        cancel,
	}

	jp.startWorkers()

	return jp
}

func (jp *JobProcessor) Process(ctx context.Context, job *domain.Job) {
	// Simply add the job to the channel for processing by workers
	ctx = logger.WithJobID(ctx, job.ID)
	logger.Ctx(ctx).Info().
		Str("url", job.URL).
		Msg("Received new job for processing")
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
	err := jp.jobRepository.UpdateJobToProcessing(ctx, job.ID)
	if err != nil {
		logger.Ctx(ctx).Error().
			Str("url", job.URL).
			Err(err).
			Msg("Failed to update job status to processing")
		return
	}

	logger.Ctx(ctx).Info().
		Str("url", job.URL).
		Msg("Started processing job")

	_, err = url.ParseRequestURI(job.URL)
	if err != nil {
		err = jp.jobRepository.UpdateJobToFailed(ctx, job.ID, "Invalid URL: "+err.Error())
		if err != nil {
			logger.Ctx(ctx).Error().
				Str("url", job.URL).
				Err(err).
				Msg("Failed to update job status to failed")
			return
		}

		logger.Ctx(ctx).Error().
			Str("url", job.URL).
			Err(err).
			Msg("Failed to parse job URL")
		return
	}

	_, err = jp.chromeDriver.CaptureScreenshot(ctx, job)
	if err != nil {
		err = jp.jobRepository.UpdateJobToFailed(ctx, job.ID, err.Error())
		if err != nil {
			logger.Ctx(ctx).Error().
				Str("url", job.URL).
				Err(err).
				Msg("Failed to update job status to failed")
			return
		}

		logger.Ctx(ctx).Error().
			Str("url", job.URL).
			Err(err).
			Msg("Failed to capture screenshot")
		return
	}

	err = jp.jobRepository.UpdateJobToCompleted(ctx, job.ID)
	if err != nil {
		logger.Ctx(ctx).Error().
			Str("url", job.URL).
			Err(err).
			Msg("Failed to update job status to completed")
		return
	}

	logger.Ctx(ctx).Info().
		Str("url", job.URL).
		Msg("Finished processing job")
}

func (jp *JobProcessor) Close() error {
	close(jp.jobs)
	return nil
}
