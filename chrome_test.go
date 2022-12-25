package chromedpundetected

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/hashicorp/go-multierror"
)

func TestChromedpundetected(t *testing.T) {
	testRun(t,
		3,
		NewConfig(
			WithTimeout(20*time.Second),
			WithHeadless(),
		),
		func(ctx context.Context) error {
			return chromedp.Run(ctx,
				chromedp.Navigate("https://nowsecure.nl"),
				chromedp.WaitVisible(`//div[@class="hystericalbg"]`),
			)
		},
	)

	t.Logf("Undetected!")
}

func testRun(t *testing.T, n int, cfg Config, run func(ctx context.Context) error) {
	var gerr error
	var success bool

	// Attempt to run the tests multiple times, as in CI they are unstable.
	for i := 0; i < n; i++ {
		t.Logf("Attempt %d/%d", i+1, n)

		ctx, cancel, err := New(cfg)
		if err != nil {
			gerr = multierror.Append(gerr, fmt.Errorf("create context: %w", err))
			continue
		}

		if err = run(ctx); err != nil {
			gerr = multierror.Append(gerr, fmt.Errorf("chromedp run: %w", err))

			// Close Chrome instance.
			ctxN, cancelN := context.WithTimeout(ctx, time.Second*10)
			if err := chromedp.Cancel(ctxN); err != nil {
				t.Logf("failed to cancel: %v", err)
			}
			cancelN()
			cancel()

			continue
		}

		success = true
		if gerr != nil {
			t.Logf("attempt %d, errors: %s", i+1, gerr.Error())
		}

		cancel()
		break
	}

	if !success {
		t.Fatal(gerr)
	}
}
