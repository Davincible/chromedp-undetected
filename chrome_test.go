package chromedpundetected

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

func TestChromedpundetected(t *testing.T) {
	c := New(NewConfig(WithTimeout(20 * time.Second)))
	err := c.Run(
		chromedp.Navigate("https://nowsecure.nl"),
		chromedp.WaitVisible(`//div[@class="hystericalbg"]`),
	)
	assert.NoError(t, err)
}
