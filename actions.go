package chromedpundetected

import (
	"context"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"golang.org/x/exp/slog"
)

func Navigate(url string) chromedp.NavigateAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		if isWebdriver(ctx) {
			slog.Info("Webdriver detected, running scripts")
			if _, err := page.AddScriptToEvaluateOnNewDocument(scriptWebDriverAttr).Do(ctx); err != nil {
				slog.Error("failed to add script webdriver attr", err)
			}

			// Remove 'Headless' from user agent.
			slog.Info("User Agent: " + getUserAgent(ctx))
			if err := cdp.Execute(ctx, "Network.setUserAgentOverride", //nolint:errcheck
				userAgent{strings.ReplaceAll(getUserAgent(ctx), "Headless", "")}, nil); err != nil {
				slog.Error("failed to override user-agent", err)
			}

			slog.Info("User Agent after: " + getUserAgent(ctx))

			if _, err := page.AddScriptToEvaluateOnNewDocument(scriptHeadless).Do(ctx); err != nil {
				slog.Error("failed to add script headless", err)
			}
		}

		slog.Info("detect webdriver", slog.Bool("webdriver", isWebdriver(ctx)))

		return chromedp.Navigate(url).Do(ctx)
	})
}

func isWebdriver(ctx context.Context) bool {
	var webdriver bool

	if err := chromedp.Evaluate("navigator.webdriver", &webdriver).Do(ctx); err != nil {
		return false
	}

	return webdriver
}

func getUserAgent(ctx context.Context) string {
	var ua string
	_ = chromedp.Evaluate("navigator.userAgent", &ua).Do(ctx)

	return ua
}
