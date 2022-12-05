# Undetected Chromedp

A small library that provides a chromedp context with a browser configured to mimick
a regular browser to prevent triggering anti-bot measures. This is not a fool proof
method, and how you use it will still dictate whether you will run into anti-bot
detection, but at least it won't trigger on all the basic detection tests.

The headless option only works on linux, and required `Xvfb` to be installed.
Could theoretically work on Mac OS with [xquartz](https://www.xquartz.org/)
but I don't have a Mac to test with, so feel free to PR.

```go
package main

import (
	"time"

	cu "github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/chromedp"
)

func main() {
	// New creates a new context for use with Chromedp. With this context
	// you can use chromedp as you normally would.
	ctx, cancel, err := cu.New(cu.NewConfig(
		// Remove this if you want to see a browser window.
		cu.WithHeadless(),

		// If the webelement is not found within 10 seconds, timeout.
		cu.WithTimeout(10 * time.Second),
	))
	if err != nil {
		panic(err)
	}
	defer cancel()

	if err := chromedp.Run(ctx,
		// Check if we pass anti-bot measures.
		chromedp.Navigate("https://nowsecure.nl"),
		chromedp.WaitVisible(`//div[@class="hystericalbg"]`),
	); err != nil {
		panic(err)
	}

	fmt.Println("Undetected!")
}
```

> Based on [undetected-chromedriver](https://github.com/ultrafunkamsterdam/undetected-chromedriver)

