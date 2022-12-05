package chromedpundetected

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

func TestChromedpundetected(t *testing.T) {
	c, cancel, err := New(NewConfig(
		WithTimeout(20*time.Second),
		WithHeadless(),
	))
	defer cancel()
	assert.NoError(t, err)

	err = c.Run(
		// chromedp.Navigate("https://nowsecure.nl"),
		Navigate("https://nowsecure.nl"),
		chromedp.WaitVisible(`//div[@class="hystericalbg"]`),
	)
	assert.NoError(t, err)
}
