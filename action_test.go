package chromedpundetected

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestRunCommand(t *testing.T) {
	testRun(t,
		3,
		NewConfig(
			WithTimeout(20*time.Second),
			WithHeadless(),
		),
		func(ctx context.Context) error {
			version := make(map[string]string)
			err := chromedp.Run(ctx,
				RunCommandWithRes("Browser.getVersion", nil, &version),
			)
			t.Log("Version:", version)
			return err
		},
	)
}

func TestBlockURLs(t *testing.T) {
	btn := `//button[@title="Akkoord"]`

	testRun(t,
		3,
		NewConfig(
			WithTimeout(20*time.Second),
			WithHeadless(),
		),
		func(ctx context.Context) error {
			if err := chromedp.Run(ctx,
				chromedp.Navigate("https://www.nu.nl/"),
				chromedp.WaitVisible(btn),
				chromedp.Click(btn),
			); err != nil {
				return err
			}

			if err := chromedp.Run(ctx,
				BlockURLs("*.nu.nl"),
				chromedp.Navigate("https://www.nu.nl/"),
				chromedp.WaitVisible(btn),
			); err != nil && !errors.Is(err, context.DeadlineExceeded) {
				return err
			}

			return nil
		},
	)
}
