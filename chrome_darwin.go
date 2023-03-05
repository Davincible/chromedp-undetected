//go:build darwin

// Package chromedpundetected provides a chromedp context with an undetected
// Chrome browser.
package chromedpundetected

import (
	"errors"

	"github.com/chromedp/chromedp"
)

func headlessOpts() (opts []chromedp.ExecAllocatorOption, cleanup func() error, err error) {
	return nil, nil, errors.New("headless mode not supported in darwin")
}
