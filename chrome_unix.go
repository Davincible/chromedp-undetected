//go:build unix

// Package chromedpundetected provides a chromedp context with an undetected
// Chrome browser.
package chromedpundetected

import (
	"context"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/Xuanwo/go-locale"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
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

	headlessOpts, closeFrameBuffer, err := headlessFlag(config)
	if err != nil {
		return nil, func() {}, err
	}

	opts = append(opts, localeFlag())
	opts = append(opts, supressWelcomeFlag()...)
	opts = append(opts, logLevelFlag(config))
	opts = append(opts, debuggerAddrFlag(config)...)
	opts = append(opts, noSandboxFlag(config)...)
	opts = append(opts, chromedp.UserDataDir(userDataDir))
	opts = append(opts, headlessOpts...)
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

		if err := closeFrameBuffer(); err != nil {
			slog.Error("failed to close Xvfb", err)
		}

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

func headlessFlag(config Config) ([]chromedp.ExecAllocatorOption, func() error, error) {
	var opts []chromedp.ExecAllocatorOption

	cleanup := func() error { return nil }

	if config.Headless {
		// Create virtual display
		frameBuffer, err := newFrameBuffer("1920x1080x24")
		if err != nil {
			return nil, nil, err
		}

		cleanup = frameBuffer.Stop

		opts = append(opts,
			// chromedp.Flag("headless", true),
			chromedp.Flag("window-size", "1920,1080"),
			chromedp.Flag("start-maximized", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.ModifyCmdFunc(func(cmd *exec.Cmd) {
				cmd.Env = append(cmd.Env, "DISPLAY=:"+frameBuffer.Display)
				cmd.Env = append(cmd.Env, "XAUTHORITY="+frameBuffer.AuthPath)

				// Default modify command per chromedp
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

	return opts, cleanup, nil
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
