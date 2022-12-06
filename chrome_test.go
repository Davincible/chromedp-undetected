package chromedpundetected

import (
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/hashicorp/go-multierror"
)

func TestChromedpundetected(t *testing.T) {
	var gerr error
	var success bool

	for i := 0; i < 3; i++ {
		ctx, cancel, err := New(NewConfig(
			WithTimeout(20*time.Second),
			WithHeadless(),
		))
		defer cancel()
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
			continue
		}

		success = true
		break
	}

	if !success {
		t.Fatal(gerr)
	}
}
