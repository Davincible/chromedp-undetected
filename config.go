package chromedpundetected

import (
	"os"
	"path"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
)

const (
	DefaultNoSandbox = true
)

var (
	DefaultUserDataDir = path.Join(os.TempDir(), "chromedp-userdata-"+"-"+uuid.NewString())
)

type Option func(*Config)

type Config struct {
	UserDataDir    string                   `json:"userDataDir" yaml:"userDataDir"`
	LogLevel       int                      `json:"logLevel" yaml:"logLevel"`
	NoSandbox      bool                     `json:"noSandbox" yaml:"noSandbox"`
	ChromePath     string                   `json:"chromePath" yaml:"chromePath"`
	Port           int                      `json:"port" yaml:"port"`
	Timeout        time.Duration            `json:"timeout" yaml:"timeout"`
	Headless       bool                     `json:"headless" yaml:"headless"`
	ContextOptions []chromedp.ContextOption `json:"-" yaml:"-"`
}

func NewConfig(opts ...Option) Config {
	c := Config{
		NoSandbox:   DefaultNoSandbox,
		UserDataDir: DefaultUserDataDir,
	}

	for _, o := range opts {
		o(&c)
	}

	return c
}

func WithUserDataDir(dir string) Option {
	return func(c *Config) {
		c.UserDataDir = dir
	}
}

func WithChromeBinary(path string) Option {
	return func(c *Config) {
		c.ChromePath = path
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

func WithHeadless() Option {
	return func(c *Config) {
		c.Headless = true
	}
}
