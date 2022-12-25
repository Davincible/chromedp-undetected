package chromedpundetected

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

func TestRunCommand(t *testing.T) {
	ctx, cancel, err := New(NewConfig(
		WithTimeout(10*time.Second),
		WithHeadless(),
	))
	require.NoError(t, err, "create context")
	defer cancel()

	version := make(map[string]string)
	err = chromedp.Run(ctx,
		RunCommandWithRes("Browser.getVersion", nil, &version),
	)
	require.NoError(t, err, "run")

	t.Log("Version:", version)
}

func TestBlockURLs(t *testing.T) {
	ctx, cancel, err := New(NewConfig(
		WithTimeout(10*time.Second),
		WithHeadless(),
	))
	require.NoError(t, err, "create context")
	defer cancel()

	btn := `//button[@title="Akkoord"]`
	err = chromedp.Run(ctx,
		chromedp.Navigate("https://www.nu.nl/"),
		chromedp.WaitVisible(btn),
		chromedp.Click(btn),
	)
	require.NoError(t, err, "check button")

	err = chromedp.Run(ctx,
		BlockURLs("*.nu.nl"),
		chromedp.Navigate("https://www.nu.nl/"),
		chromedp.WaitVisible(btn),
	)
	t.Log(err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
