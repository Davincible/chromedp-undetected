package chromedpundetected

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	// DefaultNoSandbox enables the 'no-sandbox' flag by default.
	DefaultNoSandbox = true
)

// Option is a functional option.
type Option func(*Config)

// Config is a undetected Chrome config.
type Config struct {
	// Ctx is the base context to use. By default, a background context will be used.
	Ctx context.Context `json:"-" yaml:"-"`

	// ContextOptions are chromedp context option.
	ContextOptions []chromedp.ContextOption `json:"-" yaml:"-"`

	// ChromeFlags are additional Chrome flags to pass to the browser.
	//
	// NOTE: adding additional flags can make the detection unstable, so test,
	// and be careful of what flags you add. Mostly intended to configure things
	// like a proxy. Also check if the flags you want to set are not already set
	// by this library.
	ChromeFlags []chromedp.ExecAllocatorOption

	// UserDataDir is the path to the directory where Chrome user data is stored.
	//
	// By default a temporary directory will be used.
	UserDataDir string `json:"userDataDir" yaml:"userDataDir"`

	// LogLevel is the Chrome log level, 0 by default.
	LogLevel int `json:"logLevel" yaml:"logLevel"`

	// NoSandbox dictates whether the no-sanbox flag is added. Defaults to true.
	NoSandbox bool `json:"noSandbox" yaml:"noSandbox"`

	// ChromePath is a specific binary path for Chrome.
	//
	// By default the chrome or chromium on your PATH will be used.
	ChromePath string `json:"chromePath" yaml:"chromePath"`

	// Port is the Chrome debugger port. By default a random port will be used.
	Port int `json:"port" yaml:"port"`

	// Timeout is the context timeout.
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// Headless dicates whether Chrome will start headless (without a visible window)
	//
	// It will NOT use the '--headless' option, rather it will use a virtual display.
	// Requires Xvfb to be installed, only available on Linux.
	Headless bool `json:"headless" yaml:"headless"`
}

// NewConfig creates a new config object with defaults.
func NewConfig(opts ...Option) Config {
	c := Config{
		NoSandbox: DefaultNoSandbox,
	}

	for _, o := range opts {
		o(&c)
	}

	return c
}

// WithContext adds a base context.
func WithContext(ctx context.Context) Option {
	return func(c *Config) {
		c.Ctx = ctx
	}
}

// WithUserDataDir sets the user data directory to a custom path.
func WithUserDataDir(dir string) Option {
	return func(c *Config) {
		c.UserDataDir = dir
	}
}

// WithChromeBinary sets the chrome binary path.
func WithChromeBinary(path string) Option {
	return func(c *Config) {
		c.ChromePath = path
	}
}

// WithTimeout sets the context timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithHeadless creates a headless chrome instance.
func WithHeadless() Option {
	return func(c *Config) {
		c.Headless = true
	}
}

// WithNoSandbox enable/disable sandbox. Disabled by default.
func WithNoSandbox(b bool) Option {
	return func(c *Config) {
		c.NoSandbox = b
	}
}

// WithPort sets the chrome debugger port.
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithLogLevel sets the chrome log level.
func WithLogLevel(level int) Option {
	return func(c *Config) {
		c.LogLevel = level
	}
}

// WithChromeFlags add chrome flags.
func WithChromeFlags(opts ...chromedp.ExecAllocatorOption) Option {
	return func(c *Config) {
		c.ChromeFlags = append(c.ChromeFlags, opts...)
	}
}
