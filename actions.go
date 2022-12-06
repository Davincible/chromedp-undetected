package chromedpundetected

import (
	"context"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

// UserAgentOverride overwrites the Chrome user agent.
//
// It's better to use this method than emulation.UserAgentOverride.
func UserAgentOverride(userAgent string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		return cdp.Execute(ctx, "Network.setUserAgentOverride",
			emulation.SetUserAgentOverride(userAgent), nil)
	}
}
