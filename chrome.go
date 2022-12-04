package chromedpundetected

import (
	"context"
	"net"
	"strconv"
	"strings"

	"github.com/Xuanwo/go-locale"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
)

type Chrome struct {
	ctx     context.Context
	config  Config
	cancel  func()
	actions []chromedp.Action
}

func New(config Config) Chrome {
	c := Chrome{config: config}

	port := strconv.Itoa(c.config.Port)
	if c.config.Port == 0 {
		port = getRandomPort()
	}

	var opts []chromedp.ExecAllocatorOption
	opts = append(
		opts,
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("user-data-dir", c.config.UserDataDir),
		chromedp.Flag("log-level", strconv.Itoa(c.config.LogLevel)),
		chromedp.Flag("remote-debugging-host", "127.0.0.1"),
		chromedp.Flag("remote-debugging-port", port),
	)

	lang := "en-US"
	if tag, err := locale.Detect(); err != nil && len(tag.String()) > 0 {
		lang = tag.String()
	}
	opts = append(opts, chromedp.Flag("lang", lang))

	if c.config.NoSandbox {
		opts = append(opts,
			chromedp.Flag("no-sandbox", true), chromedp.Flag("test-type", true))
	}

	if len(c.config.ChromePath) > 0 {
		opts = append(opts, chromedp.UserDataDir("/tmp/chromedp-data-"+uuid.NewString()))
	}

	ctx := context.Background()
	cancelT := func() {}
	if c.config.Timeout > 0 {
		ctx, cancelT = context.WithTimeout(ctx, c.config.Timeout)
	}
	ctx, cancelA := chromedp.NewExecAllocator(ctx, opts...)
	ctx, cancelC := chromedp.NewContext(ctx, c.config.ContextOptions...)
	c.cancel = func() {
		cancelT()
		cancelA()
		cancelC()
	}

	c.ctx = ctx

	return c
}

func (c *Chrome) Run(actions ...chromedp.Action) error {
	return chromedp.Run(c.ctx, append(c.actions, actions...)...)
}

func getRandomPort() string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err == nil {
		addr := l.Addr().String()
		l.Close()

		return strings.Split(addr, ":")[1]
	}

	return "42069"
}
