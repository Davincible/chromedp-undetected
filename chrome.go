package chromedpundetected

import (
	"context"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/Xuanwo/go-locale"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
)

type Chrome struct {
	ctx         context.Context
	config      Config
	cancel      func()
	actions     []chromedp.Action
	frameBuffer *FrameBuffer
}

func New(config Config) (Chrome, context.CancelFunc, error) {
	c := Chrome{config: config}

	port := strconv.Itoa(c.config.Port)
	if c.config.Port == 0 {
		port = getRandomPort()
	}

	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("user-data-dir", c.config.UserDataDir),
		chromedp.Flag("log-level", strconv.Itoa(c.config.LogLevel)),
		chromedp.Flag("remote-debugging-host", "127.0.0.1"),
		chromedp.Flag("remote-debugging-port", port),
	}

	// Locale
	lang := "en-US"
	if tag, err := locale.Detect(); err != nil && len(tag.String()) > 0 {
		lang = tag.String()
	}

	opts = append(opts, chromedp.Flag("lang", lang))

	// Sandbox
	if c.config.NoSandbox {
		opts = append(opts,
			chromedp.Flag("no-sandbox", true), chromedp.Flag("test-type", true))
	}

	// Userdata profile path
	if len(c.config.ChromePath) > 0 {
		opts = append(opts, chromedp.UserDataDir("/tmp/chromedp-data-"+uuid.NewString()))
	}

	// Headless
	if c.config.Headless {
		// Create virtual display
		fb, err := NewFrameBuffer("1920x1080x24")
		if err != nil {
			return c, nil, err
		}

		opts = append(opts,
			// chromedp.Flag("headless", true),
			chromedp.Flag("window-size", "1920,1080"),
			chromedp.Flag("start-maximized", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.ModifyCmdFunc(func(cmd *exec.Cmd) {
				cmd.Env = append(cmd.Env, "DISPLAY=:"+fb.Display)
				cmd.Env = append(cmd.Env, "XAUTHORITY="+fb.AuthPath)

				// Default modify command
				if _, ok := os.LookupEnv("LAMBDA_TASK_ROOT"); ok {
					// do nothing on AWS Lambda
					return
				}

				if cmd.SysProcAttr == nil {
					cmd.SysProcAttr = new(syscall.SysProcAttr)
				}

				// When the parent process dies (Go), kill the child as well.
				cmd.SysProcAttr.Pdeathsig = syscall.SIGKILL
			}),
		)
	}

	ctx := context.Background()

	cancelT := func() {}
	if c.config.Timeout > 0 {
		ctx, cancelT = context.WithTimeout(ctx, c.config.Timeout)
	}

	ctx, cancelA := chromedp.NewExecAllocator(ctx, opts...)
	ctx, cancelC := chromedp.NewContext(ctx, c.config.ContextOptions...)
	cancel := func() {
		cancelT()
		cancelA()
		cancelC()
	}

	c.ctx = ctx

	return c, cancel, nil
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
