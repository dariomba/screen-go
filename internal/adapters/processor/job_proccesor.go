package processor

import (
	"context"
	"log"
	"net/url"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/dariomba/screen-go/internal/domain"
)

type JobProcessorConfig struct {
	MaxThreads int
}

type JobProcessor struct {
	config JobProcessorConfig

	jobs chan *domain.Job

	ctx    context.Context
	cancel context.CancelFunc
}

func NewJobProcessor(config JobProcessorConfig) *JobProcessor {
	jobs := make(chan *domain.Job, config.MaxThreads)

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
	log.Printf("Received job %s: %s\n", job.ID, job.URL)
	jp.jobs <- job
}

func (jp *JobProcessor) startWorkers() {
	for i := 0; i < jp.config.MaxThreads; i++ {
		go func() {
			for {
				select {
				case <-jp.ctx.Done():
					return
				case job, ok := <-jp.jobs:
					if !ok {
						return
					}
					log.Printf("Processing job %s: %s\n", job.ID, job.URL)

					url, err := url.ParseRequestURI(job.URL)
					if err != nil {
						log.Printf("error parsing job URL %s: %v\n", job.ID, err)
						return
					}
					ctx, cancel := chromedp.NewContext(
						jp.ctx,
					)
					defer cancel()

					var buf []byte
					if job.Format == domain.JobFormatPdf {
						if err := chromedp.Run(ctx, printToPDF(url.String(), &buf)); err != nil {
							log.Printf("failed to capture PDF for job %s: %v", job.ID, err)
							return
						}
						if err := os.WriteFile("page.pdf", buf, 0o644); err != nil {
							log.Printf("failed to write PDF for job %s: %v", job.ID, err)
						}
					} else {
						if err := chromedp.Run(ctx, chromedp.Tasks{
							chromedp.Navigate(url.String()),
							chromedp.FullScreenshot(&buf, 100),
						}); err != nil {
							log.Printf("failed to capture screenshot for job %s: %v", job.ID, err)
							return
						}
						if err := os.WriteFile("elementScreenshot.png", buf, 0o644); err != nil {
							log.Printf("failed to write screenshot for job %s: %v", job.ID, err)
						}
					}
					log.Printf("Finished processing job %s\n", job.ID)
				}
			}
		}()
	}
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
