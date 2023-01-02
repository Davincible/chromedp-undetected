package chromedpundetected

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Davincible/chromedp-undetected/util/easyjson"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// Cookie is used to set browser cookies.
type Cookie struct {
	Name     string  `json:"name" yaml:"name"`
	Value    string  `json:"value" yaml:"value"`
	Domain   string  `json:"domain" yaml:"domain"`
	Path     string  `json:"path" yaml:"path"`
	Expires  float64 `json:"expires" yaml:"expires"`
	HTTPOnly bool    `json:"httpOnly" yaml:"httpOnly"`
	Secure   bool    `json:"secure" yaml:"secure"`
}

// UserAgentOverride overwrites the Chrome user agent.
//
// It's better to use this method than emulation.UserAgentOverride.
func UserAgentOverride(userAgent string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		return cdp.Execute(ctx, "Network.setUserAgentOverride",
			emulation.SetUserAgentOverride(userAgent), nil)
	}
}

// LoadCookiesFromFile takes a file path to a json file containing cookies, and
// loads in the cookies into the browser.
func LoadCookiesFromFile(path string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		f, err := os.Open(path) //nolint:gosec
		if err != nil {
			return fmt.Errorf("failed to open file '%s': %w", path, err)
		}

		data, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}

		var cookies []Cookie
		if err := json.Unmarshal(data, &cookies); err != nil {
			return fmt.Errorf("unmarshal cookies from json: %w", err)
		}

		return LoadCookies(cookies)(ctx)
	})
}

// LoadCookies will load a set of cookies into the browser.
func LoadCookies(cookies []Cookie) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for _, cookie := range cookies {
			expiry := cdp.TimeSinceEpoch(time.Unix(int64(cookie.Expires), 0))
			if err := network.SetCookie(cookie.Name, cookie.Value).
				WithHTTPOnly(cookie.HTTPOnly).
				WithSecure(cookie.Secure).
				WithDomain(cookie.Domain).
				WithPath(cookie.Path).
				WithExpires(&expiry).
				Do(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// SaveCookies extracts the cookies from the current URL and appends them to
// provided array.
func SaveCookies(cookies *[]Cookie) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		c, err := network.GetCookies().Do(ctx)
		if err != nil {
			return err
		}

		for _, cookie := range c {
			*cookies = append(*cookies, Cookie{
				Name:     cookie.Name,
				Value:    cookie.Value,
				Domain:   cookie.Domain,
				Path:     cookie.Path,
				Expires:  cookie.Expires,
				HTTPOnly: cookie.HTTPOnly,
				Secure:   cookie.HTTPOnly,
			})
		}

		return nil
	})
}

// SaveCookiesTo extracts the cookies from the current page and saves them
// as JSON to the provided path.
func SaveCookiesTo(path string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var c []Cookie

		if err := SaveCookies(&c).Do(ctx); err != nil {
			return err
		}

		b, err := json.Marshal(c)
		if err != nil {
			return err
		}

		if err := os.WriteFile(path, b, 0644); err != nil { //nolint:gosec
			return err
		}

		return nil
	})
}

// RunCommandWithRes runs any Chrome Dev Tools command, with any params and
// sets the result to the res parameter. Make sure it is a pointer.
//
// In contrast to the native method of chromedp, with this method you can directly
// pass in a map with the data passed to the command.
func RunCommandWithRes(method string, params, res any) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		i := easyjson.New(params)
		o := easyjson.New(res)

		return cdp.Execute(ctx, method, i, o)
	})
}

// RunCommand runs any Chrome Dev Tools command, with any params.
//
// In contrast to the native method of chromedp, with this method you can directly
// pass in a map with the data passed to the command.
func RunCommand(method string, params any) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		i := easyjson.New(params)

		return cdp.Execute(ctx, method, i, nil)
	})
}

// BlockURLs blocks a set of URLs in Chrome.
func BlockURLs(url ...string) chromedp.ActionFunc {
	return RunCommand("Network.setBlockedURLs", map[string][]string{"urls": url})
}
