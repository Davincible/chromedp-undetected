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
	var gerr error
	var success bool

	// Attempt to run the tests multiple times, as in CI they are unstable.
	for i := 0; i < 3; i++ {
		ctx, cancel, err := New(NewConfig(
			WithTimeout(20*time.Second),
			WithHeadless(),
		))
		if err != nil {
			gerr = multierror.Append(gerr, fmt.Errorf("create context: %w", err))
			continue
		}

		err = chromedp.Run(ctx,
			chromedp.Navigate("https://nowsecure.nl"),
			chromedp.WaitVisible(`//div[@class="hystericalbg"]`),
		)
		if err != nil {
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
			t.Logf("attempt %d, errors: %s", i, gerr.Error())
		}

		cancel()
		break
	}

	if !success {
		t.Fatal(gerr)
	}
}
