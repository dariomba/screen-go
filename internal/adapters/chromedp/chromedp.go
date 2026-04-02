package chromedp

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/dariomba/screen-go/internal/domain"
)

type ChromedpConfig struct {
	Timeout time.Duration
	WindowX int
	WindowY int
}

type Chromedp struct {
	allocatorCtx    context.Context
	allocatorCancel context.CancelFunc
	browserCtx      context.Context
	browserCancel   context.CancelFunc

	config *ChromedpConfig
}

func NewChromedp(config *ChromedpConfig) (*Chromedp, error) {
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Headless,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoFirstRun,
		chromedp.DisableGPU,
		chromedp.IgnoreCertErrors,
		chromedp.ModifyCmdFunc(func(cmd *exec.Cmd) {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Setpgid: true, // Start the process in a new process group for graceful shutdown
			}
		}),
		chromedp.Flag("disable-features", "MediaRouter"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("deny-permission-prompts", true),
		chromedp.Flag("https-upgrades-enabled", false),
		chromedp.Flag("disable-features", "HttpsUpgrades"),
		chromedp.Flag("no-sandbox", true),
		chromedp.WindowSize(config.WindowX, config.WindowY),
	)

	// Create a temporary allocator context to warm up Chrome
	allocatorCtx, allocatorCancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)

	browserCtx, browserCancel := chromedp.NewContext(allocatorCtx)

	// Run a simple task to warm up Chrome and ensure it's ready to accept commands
	if err := chromedp.Run(browserCtx, chromedp.ActionFunc(func(context.Context) error { return nil })); err != nil {
		browserCancel()
		allocatorCancel()
		return nil, fmt.Errorf("failed to warm up chrome: %w", err)
	}

	driver := &Chromedp{
		allocatorCtx:    allocatorCtx,
		allocatorCancel: allocatorCancel,
		browserCtx:      browserCtx,
		browserCancel:   browserCancel,
		config:          config,
	}

	return driver, nil
}

func (d *Chromedp) CaptureScreenshot(ctx context.Context, job *domain.Job) ([]byte, error) {
	tabCtx, tabCancel := chromedp.NewContext(d.browserCtx)
	defer tabCancel()
	tabCtx, timeoutCancel := context.WithTimeout(tabCtx, d.config.Timeout)
	defer timeoutCancel()

	var imgBuf []byte

	tasks := append(chromedp.Tasks{},
		chromedp.Navigate(job.URL),
		chromedp.WaitReady("body", chromedp.ByQuery),
	)
	if job.Format == domain.JobFormatPdf {
		tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(false).Do(ctx)
			if err != nil {
				return err
			}
			imgBuf = buf
			return nil
		}))
	} else {
		tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
			imgParams := page.CaptureScreenshot().WithFormat(page.CaptureScreenshotFormatPng)

			if job.FullPage {
				imgParams = imgParams.WithCaptureBeyondViewport(true)
			} else {
				imgParams = imgParams.WithClip(&page.Viewport{
					X:      0,
					Y:      0,
					Width:  float64(job.Width),
					Height: float64(job.Height),
					Scale:  1,
				})
			}
			var err error
			imgBuf, err = imgParams.Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}))
	}

	err := chromedp.Run(tabCtx, tasks)

	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return imgBuf, nil
}

func (d *Chromedp) Shutdown(ctx context.Context) error {
	if err := chromedp.Cancel(d.browserCtx); err != nil {
		return fmt.Errorf("failed to cancel browser context: %w", err)
	}
	d.allocatorCancel()
	return nil
}
