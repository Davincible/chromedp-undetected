package chromedpundetected

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

func TestChromedpundetected(t *testing.T) {
	ctx, cancel, err := New(NewConfig(
		WithTimeout(20*time.Second),
		WithHeadless(),
	))
	defer cancel()
	assert.NoError(t, err)

	err = chromedp.Run(ctx,
		chromedp.Navigate("https://nowsecure.nl"),
		chromedp.WaitVisible(`//div[@class="hystericalbg"]`),
	)
	assert.NoError(t, err)
}
