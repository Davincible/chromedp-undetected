// Package chromedpundetected provides a chromedp context with an undetected
// Chrome browser.
package chromedpundetected

import (
	"context"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	
	"github.com/Xuanwo/go-locale"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
)

var (
	// DefaultUserDirPrefix Defaults.
	DefaultUserDirPrefix = "chromedp-undetected-"
)

// New creates a context with an undetected Chrome executor.
func New(config Config) (context.Context, context.CancelFunc, error) {
	var opts []chromedp.ExecAllocatorOption

	userDataDir := path.Join(os.TempDir(), DefaultUserDirPrefix+uuid.NewString())
	if len(config.ChromePath) > 0 {
		userDataDir = config.ChromePath
	}

	opts = append(opts, localeFlag())
	opts = append(opts, supressWelcomeFlag()...)
	opts = append(opts, logLevelFlag(config))
	opts = append(opts, debuggerAddrFlag(config)...)
	opts = append(opts, noSandboxFlag(config)...)
	opts = append(opts, chromedp.UserDataDir(userDataDir))
	opts = append(opts, config.ChromeFlags...)

	ctx := context.Background()
	if config.Ctx != nil {
		ctx = config.Ctx
	}

	cancelT := func() {}
	if config.Timeout > 0 {
		ctx, cancelT = context.WithTimeout(ctx, config.Timeout)
	}

	ctx, cancelA := chromedp.NewExecAllocator(ctx, opts...)
	ctx, cancelC := chromedp.NewContext(ctx, config.ContextOptions...)

	cancel := func() {
		cancelT()
		cancelA()
		cancelC()

		if len(config.ChromePath) == 0 {
			_ = os.RemoveAll(userDataDir) //nolint:errcheck
		}
	}

	return ctx, cancel, nil
}

func supressWelcomeFlag() []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
	}
}

func debuggerAddrFlag(config Config) []chromedp.ExecAllocatorOption {
	port := strconv.Itoa(config.Port)
	if config.Port == 0 {
		port = getRandomPort()
	}

	return []chromedp.ExecAllocatorOption{
		chromedp.Flag("remote-debugging-host", "127.0.0.1"),
		chromedp.Flag("remote-debugging-port", port),
	}
}

func localeFlag() chromedp.ExecAllocatorOption {
	lang := "en-US"
	if tag, err := locale.Detect(); err != nil && len(tag.String()) > 0 {
		lang = tag.String()
	}

	return chromedp.Flag("lang", lang)
}

func noSandboxFlag(config Config) []chromedp.ExecAllocatorOption {
	var opts []chromedp.ExecAllocatorOption

	if config.NoSandbox {
		opts = append(opts,
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("test-type", true))
	}

	return opts
}

func logLevelFlag(config Config) chromedp.ExecAllocatorOption {
	return chromedp.Flag("log-level", strconv.Itoa(config.LogLevel))
}

func getRandomPort() string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err == nil {
		addr := l.Addr().String()
		_ = l.Close() //nolint:errcheck,gosec

		return strings.Split(addr, ":")[1]
	}

	return "42069"
}
