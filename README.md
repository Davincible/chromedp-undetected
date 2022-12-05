# Undetected Chromedp 

[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/Davincible/chromedp-undetected?tab=doc) [![Go Report Card](https://goreportcard.com/badge/github.com/Davincible/chromedp-undetected)](https://goreportcard.com/report/github.com/Davincible/chromedp-undetected) [![Unit Tests](https://github.com/Davincible/chromedp-undetected/actions/workflows/main.yaml/badge.svg)](https://github.com/Davincible/chromedp-undetected/actions/workflows/main.yaml) [![GitHub](https://img.shields.io/github/license/Davincible/chromedp-undetected)](https://github.com/Davincible/chromedp-undetected/blob/master/LICENSE)

A small library that provides a [chromedp](https://github.com/chromedp/chromedp) 
context with a browser configured to mimick a regular browser to prevent 
triggering anti-bot measures. This is not a fool proof method, and how you use 
it will still dictate whether you will run into anti-bot detection, but at 
least it won't trigger on all the basic detection tests.

The headless option only works on linux, and required `Xvfb` to be installed.
Could theoretically work on Mac OS with [xquartz](https://www.xquartz.org/)
but I don't have a Mac to test with, so feel free to PR.

```go
package main

import (
	"time"

	"github.com/chromedp/chromedp"

	cu "github.com/Davincible/chromedp-undetected"
)

func main() {
	// New creates a new context for use with chromedp. With this context
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

