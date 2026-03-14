package processor

import (
	"context"
	"net/url"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/dariomba/screen-go/internal/domain"
	"github.com/dariomba/screen-go/internal/logger"
)

type JobProcessorConfig struct {
	MaxThreads int
}

type jobWithContext struct {
	job *domain.Job
	ctx context.Context
}

type JobProcessor struct {
	config JobProcessorConfig

	jobs chan *jobWithContext

	ctx    context.Context
	cancel context.CancelFunc
}

func NewJobProcessor(config JobProcessorConfig) *JobProcessor {
	jobs := make(chan *jobWithContext, config.MaxThreads)

	ctx, cancel := context.WithCancel(context.Background())

	jp := &JobProcessor{
		config: config,
		jobs:   jobs,
		ctx:    ctx,
		cancel: cancel,
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
	logger.Ctx(ctx).Info().
		Str("url", job.URL).
		Msg("Started processing job")

	url, err := url.ParseRequestURI(job.URL)
	if err != nil {
		logger.Ctx(ctx).Error().
			Str("url", job.URL).
			Err(err).
			Msg("Failed to parse job URL")
		return
	}
	chromeCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var buf []byte
	if job.Format == domain.JobFormatPdf {
		if err := chromedp.Run(chromeCtx, printToPDF(url.String(), &buf)); err != nil {
			logger.Ctx(ctx).Error().
				Str("url", job.URL).
				Err(err).
				Msg("Failed to capture PDF for job")
			return
		}
		if err := os.WriteFile("page.pdf", buf, 0o644); err != nil {
			logger.Ctx(ctx).Error().
				Str("url", job.URL).
				Err(err).
				Msg("Failed to write PDF for job")
		}
	} else {
		if err := chromedp.Run(chromeCtx, chromedp.Tasks{
			chromedp.Navigate(url.String()),
			chromedp.FullScreenshot(&buf, 100),
		}); err != nil {
			logger.Ctx(ctx).Error().
				Str("url", job.URL).
				Err(err).
				Msg("Failed to capture screenshot for job")
			return
		}
		if err := os.WriteFile("elementScreenshot.png", buf, 0o644); err != nil {
			logger.Ctx(ctx).Error().
				Str("url", job.URL).
				Err(err).
				Msg("Failed to write screenshot for job")
		}
	}
	logger.Ctx(ctx).Info().
		Str("url", job.URL).
		Msg("Finished processing job")
}

func printToPDF(urlstr string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(false).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}

func (jp *JobProcessor) Close() error {
	close(jp.jobs)
	return nil
}
