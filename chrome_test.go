package chromedpundetected

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

func TestChromedpundetected(t *testing.T) {
	ctx, cancel, err := New(NewConfig(
		WithTimeout(20*time.Second),
		WithHeadless(),
	))
	defer cancel()
	require.NoError(t, err, "create new context")

	err = chromedp.Run(ctx,
		chromedp.Navigate("https://nowsecure.nl"),
		chromedp.WaitVisible(`//div[@class="hystericalbg"]`),
	)
	require.NoError(t, err, "chromedp run")
}
