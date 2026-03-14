package chromedp

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/dariomba/screen-go/internal/domain"
)

type Chromedp struct {
	allocatorCtx    context.Context
	allocatorCancel context.CancelFunc
	browserCtx      context.Context
	browserCancel   context.CancelFunc
}

func NewChromedp() (*Chromedp, error) {
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Headless,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoFirstRun,
		chromedp.DisableGPU,
		chromedp.IgnoreCertErrors,
		chromedp.Flag("disable-features", "MediaRouter"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("deny-permission-prompts", true),
		chromedp.Flag("https-upgrades-enabled", false),
		chromedp.Flag("disable-features", "HttpsUpgrades"),
		chromedp.Flag("no-sandbox", true),
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
	}

	return driver, nil
}

func (d *Chromedp) CaptureScreenshot(ctx context.Context, job *domain.Job) ([]byte, error) {
	// TODO: Move this default timeout to config
	timeout := 30 * time.Second

	tabCtx, tabCancel := chromedp.NewContext(d.browserCtx)
	defer tabCancel()
	tabCtx, timeoutCancel := context.WithTimeout(tabCtx, timeout)
	defer timeoutCancel()

	var buf []byte
	err := chromedp.Run(tabCtx, snapshot(job.URL, &buf))

	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return buf, nil
}

func snapshot(url string, buf *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.CaptureScreenshot(buf),
	}
}
