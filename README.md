# Undetected chromedp 

[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/Davincible/chromedp-undetected?tab=doc) [![Go Report Card](https://goreportcard.com/badge/github.com/Davincible/chromedp-undetected)](https://goreportcard.com/report/github.com/Davincible/chromedp-undetected) [![Unit Tests](https://github.com/Davincible/chromedp-undetected/actions/workflows/main.yaml/badge.svg)](https://github.com/Davincible/chromedp-undetected/actions/workflows/main.yaml) [![GitHub](https://img.shields.io/github/license/Davincible/chromedp-undetected)](https://github.com/Davincible/chromedp-undetected/blob/master/LICENSE)

A small library that provides a [chromedp](https://github.com/chromedp/chromedp) 
context with a browser configured to mimick a regular browser to prevent 
triggering anti-bot measures. This is not a fool proof method, and how you use 
it will still dictate whether you will run into anti-bot detection, but at 
least it won't trigger on all the basic detection tests.

The headless option only works on linux, and requires `Xvfb` to be installed.
Could theoretically work on Mac OS with [xquartz](https://www.xquartz.org/)
but I don't have a Mac to test with, so feel free to PR.

A Docker container example is provided in `Dockerfile`. The most important things
to note is to not use the headless chrome image as base, but to normally install 
chrome or chromium, and to install xvfb. Note that this image is neither secure
nor optimized, and merely serves as an example.

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

### Utilities

Some utility functions are included I was missing in chromedp itself.

```go
// BlockURLs blocks a set of URLs in Chrome.
func BlockURLs(url ...string) chromedp.ActionFunc 

// LoadCookies will load a set of cookies into the browser.
func LoadCookies(cookies []Cookie) chromedp.ActionFunc

// LoadCookiesFromFile takes a file path to a json file containing cookies,
// and loads in the cookies into the browser.
func LoadCookiesFromFile(path string) chromedp.ActionFunc

// SaveCookies extracts the cookies from the current URL and appends them to
// provided array.
func SaveCookies(cookies *[]Cookie) chromedp.ActionFunc

// SaveCookiesTo extracts the cookies from the current page and saves them
// as JSON to the provided path.
func SaveCookiesTo(path string) chromedp.ActionFunc

// RunCommand runs any Chrome Dev Tools command, with any params.
// 
// In contrast to the native method of chromedp, with this method you can
// directly pass in a map with the data passed to the command.
func RunCommand(method string, params any) chromedp.ActionFunc

// RunCommandWithRes runs any Chrome Dev Tools command, with any params and
// sets the result to the res parameter. Make sure it is a pointer.
// 
// In contrast to the native method of chromedp, with this method you can
// directly pass in a map with the data passed to the command.
func RunCommandWithRes(method string, params, res any) chromedp.ActionFunc

// UserAgentOverride overwrites the Chrome user agent.
// 
// It's better to use this method than emulation.UserAgentOverride.
func UserAgentOverride(userAgent string) chromedp.ActionFunc
```
