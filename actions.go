package chromedpundetected

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"github.com/Davincible/chromedp-undetected/util/easyjson"
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

		b, err := json.MarshalIndent(c, "", "  ")
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

// SendKeys does the same as chromedp.SendKeys excepts it randomly waits 100-500ms
// between sending key presses.
func SendKeys(sel any, v string, opts ...chromedp.QueryOption) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for _, key := range v {
			if err := chromedp.SendKeys(sel, string(key), opts...).Do(ctx); err != nil {
				return err
			}
			s := rand.Int63n(100) + 100 //nolint:gosec
			time.Sleep(time.Duration(s) * time.Millisecond)
		}

		return nil
	})
}

var (
	removeMouseVisualsJS = `window.removeEventListener('mousemove', drawDotAtCoords);`

	// addMouseVisualsJS adds a mouse event listener to the page that draws a
	// red dot at the current mouse position to visualize the mouse path.
	addMouseVisualsJS = `
  function drawDotAtCoords(event) {
      var x = event.clientX;
      var y = event.clientY;
  
      // Create a dot
      var dot = document.createElement("div");
      var dotSize = 8;  // Set to 2px to make a small dot
      dot.style.width = dotSize + "px";
      dot.style.height = dotSize + "px";
      dot.style.backgroundColor = "red";
      dot.style.position = "absolute";
      dot.style.top = (y - dotSize/2) + "px";  // Adjusting by half the size to center it
      dot.style.left = (x - dotSize/2) + "px";  // Adjusting by half the size to center it
      dot.style.borderRadius = "50%";
      dot.style.pointerEvents = "none"; // So it doesn't interfere with other mouse events
      dot.style.padding = "0";  // Setting padding to zero
      dot.style.margin = "0";  // Setting margin to zero
      dot.style.transition = "opacity 1s";  // Setting transition for fading effect
  
      document.body.appendChild(dot);
  
      // Fade out the dot after a delay
      setTimeout(function() {
          dot.style.opacity = "0";
          
          // Remove the dot from the DOM after it's fully faded
          setTimeout(function() {
              dot.remove();
          }, 10000);
  
      }, 3000);
  }

  window.addEventListener('mousemove', drawDotAtCoords);
  
	`

	// mouseTrackingJS adds a global state of the last mouse position so we can
	// start moving from the current mouse position instead of 0,0.
	mouseTrackingJS = `
  // Global storage on window object for mouse position
  window.globalMousePos = { x: 0, y: 0 };
  
  window.addEventListener('mousemove', (event) => {
      const x = event.x;
      const y = event.y;
  
  		if (x > 0 || y > 0) {
        window.globalMousePos = { x, y };
  		}
  
  		console.log(x, y, event);
  });
  
  // Function to get the current mouse position or default to zero
  function getCurrentMousePosition() {
      return window.globalMousePos || { x: 0, y: 0 };
  }
  `
)

// MouseMoveOptions contains options for mouse movement.
type MouseMoveOptions struct {
	steps          int
	delayMin       time.Duration
	delayMax       time.Duration
	randomJitter   float64
	visualizeMouse bool
}

// Default values for mouse movement.
var defaultMouseMoveOptions = MouseMoveOptions{
	steps:          20,
	delayMin:       5 * time.Millisecond,
	delayMax:       50 * time.Millisecond,
	randomJitter:   3,
	visualizeMouse: false,
}

// MoveOptionSetter defines a function type to set mouse move options.
type MoveOptionSetter func(*MouseMoveOptions)

// WithSteps returns a MoveOptionSetter that sets the number of steps for the mouse movement.
func WithSteps(s int) MoveOptionSetter {
	return func(opt *MouseMoveOptions) {
		opt.steps = s
	}
}

// WithDelayRange returns a MoveOptionSetter that sets the delay range between steps.
func WithDelayRange(min, max time.Duration) MoveOptionSetter {
	return func(opt *MouseMoveOptions) {
		opt.delayMin = min
		opt.delayMax = max
	}
}

// WithRandomJitter returns a MoveOptionSetter that sets the random jitter to introduce in mouse movement.
func WithRandomJitter(jitter float64) MoveOptionSetter {
	return func(opt *MouseMoveOptions) {
		opt.randomJitter = jitter
	}
}

// WithVisualizeMouse returns a MoveOptionSetter that enables mouse movement visualization.
func WithVisualizeMouse() MoveOptionSetter {
	return func(opt *MouseMoveOptions) {
		opt.visualizeMouse = true
	}
}

// MoveMouseToPosition moves the mouse to the given position, mimic random human mouse movements.
//
// If desired you can tweak the mouse movement behavior, defaults are set to mimic human mouse movements.
func MoveMouseToPosition(x, y float64, setters ...MoveOptionSetter) chromedp.ActionFunc { //nolint:varnamelen
	options := defaultMouseMoveOptions

	for _, setter := range setters {
		setter(&options)
	}

	return func(ctx context.Context) error {
		var pos struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		}

		if err := chromedp.Evaluate(`getCurrentMousePosition()`, &pos).Do(ctx); err != nil {
			if err := chromedp.Evaluate(mouseTrackingJS, nil).Do(ctx); err != nil {
				return fmt.Errorf("inject mouse position tracing js: %w", err)
			}
		}

		// Add mouse visualization event listener if enabled.
		if options.visualizeMouse {
			if err := chromedp.Evaluate(addMouseVisualsJS, nil).Do(ctx); err != nil {
				return fmt.Errorf("inject mouse visualization js: %w", err)
			}

			// Remove mouse visualization event listere after mouse movemvent is complete.
			defer func() {
				chromedp.Evaluate(removeMouseVisualsJS, nil).Do(ctx) //nolint:errcheck,gosec
			}()
		}

		// Generate control points for Bezier curve.
		control1 := point{
			x: pos.X + rand.Float64()*math.Max(x-pos.X, (y-pos.Y)/2),
			y: pos.Y + rand.Float64()*math.Max(y-pos.Y, (x-pos.X)/2),
		}

		control2 := point{
			x: x - rand.Float64()*(x-pos.X),
			y: y - rand.Float64()*(y-pos.Y),
		}

		start := point{x: pos.X, y: pos.Y}
		end := point{x: x, y: y}

		for i := 0; i <= options.steps; i++ {
			t := float64(i) / float64(options.steps)
			point := bezierCubic(start, control1, control2, end, t)

			targetX := point.x + rand.Float64()*options.randomJitter - options.randomJitter
			targetY := point.y + rand.Float64()*options.randomJitter - options.randomJitter

			p := &input.DispatchMouseEventParams{
				Type:   input.MouseMoved,
				X:      targetX,
				Y:      targetY,
				Button: input.None,
			}

			if err := p.Do(ctx); err != nil {
				return err
			}

			sleepDuration := options.delayMin + time.Duration(rand.Int63n(int64(options.delayMax-options.delayMin)))
			time.Sleep(sleepDuration)
		}

		return nil
	}
}

type point struct {
	x, y float64
}

// Returns a point along a cubic Bézier curve.
// t is the "progress" along the curve, should be between 0 and 1.
func bezierCubic(p0, p1, p2, p3 point, t float64) point {
	mt := 1 - t
	mt2 := mt * mt
	t2 := t * t

	return point{
		x: mt2*mt*p0.x + 3*mt2*t*p1.x + 3*mt*t2*p2.x + t2*t*p3.x,
		y: mt2*mt*p0.y + 3*mt2*t*p1.y + 3*mt*t2*p2.y + t2*t*p3.y,
	}
}
