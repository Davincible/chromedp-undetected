package chromedpundetected

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestNetworkIdleListener(t *testing.T) {
	testRun(t,
		n,
		NewConfig(
			WithTimeout(20*time.Second),
			WithHeadless(),
		),
		func(ctx context.Context) error {
			idleListener := NetworkIdleListener(ctx, time.Second, time.Second*10)

			if err := chromedp.Run(ctx,
				chromedp.Navigate("https://nowsecure.nl"),
			); err != nil {
				return err
			}

			if event := <-idleListener; !event.IsIdle {
				return fmt.Errorf("expected idle event, got %v", event)
			}

			return nil
		},
	)
}
