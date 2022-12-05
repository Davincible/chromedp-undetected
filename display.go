package chromedpundetected

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FrameBuffer controls an X virtual frame buffer running as a background
// process.
type FrameBuffer struct {
	// Display is the X11 display number that the Xvfb process is hosting
	// (without the preceding colon).
	Display string

	// AuthPath is the path to the X11 authorization file that permits X clients
	// to use the X server. This is typically provided to the client via the
	// XAUTHORITY environment variable.
	AuthPath string

	cmd *exec.Cmd
}

// FrameBufferOptions describes the options that can be used to create a frame buffer.
type FrameBufferOptions struct {
	// ScreenSize is the option for the frame buffer screen size.
	// This is of the form "{width}x{height}[x{depth}]".  For example: "1024x768x24"
	ScreenSize string
}

// NewFrameBufferWithOptions starts an X virtual frame buffer running in the background.
// FrameBufferOptions may be populated to change the behavior of the frame buffer.
func NewFrameBuffer(screenSize string) (*FrameBuffer, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	auth, err := ioutil.TempFile("", "selenium-xvfb")
	if err != nil {
		return nil, err
	}

	authPath := auth.Name()
	if err := auth.Close(); err != nil {
		return nil, err
	}

	// Xvfb will print the display on which it is listening to file descriptor 3,
	// for which we provide a pipe.
	arguments := []string{"-displayfd", "3", "-nolisten", "tcp"}

	if screenSize != "" {
		screenSizeExpression := regexp.MustCompile(`^\d+x\d+(?:x\d+)$`)
		if !screenSizeExpression.MatchString(screenSize) {
			return nil, fmt.Errorf("invalid screen size: expected 'WxH[xD]', got %q", screenSize)
		}

		arguments = append(arguments, "-screen", "0", screenSize)
	}

	xvfb := exec.Command("Xvfb", arguments...)
	xvfb.ExtraFiles = []*os.File{w}

	// TODO(minusnine): plumb a way to set xvfb.Std{err,out} conditionally.
	// TODO(minusnine): Pdeathsig is only supported on Linux. Somehow, make sure
	// process cleanup happens as gracefully as possible.
	xvfb.Env = append(xvfb.Env, "XAUTHORITY="+authPath)
	if err := xvfb.Start(); err != nil {
		return nil, err
	}
	w.Close()

	type resp struct {
		display string
		err     error
	}

	ch := make(chan resp)

	go func() {
		bufr := bufio.NewReader(r)
		s, err := bufr.ReadString('\n')
		ch <- resp{s, err}
	}()

	var display string
	select {
	case resp := <-ch:
		if resp.err != nil {
			return nil, resp.err
		}

		display = strings.TrimSpace(resp.display)
		if _, err := strconv.Atoi(display); err != nil {
			return nil, errors.New("xvfb did not print the display number")
		}

	case <-time.After(3 * time.Second):
		return nil, errors.New("timeout waiting for Xvfb")
	}

	xauth := exec.Command("xauth", "generate", ":"+display, ".", "trusted")
	xauth.Stderr = os.Stderr
	xauth.Stdout = os.Stdout
	xauth.Env = append(xauth.Env, "XAUTHORITY="+authPath)

	if err := xauth.Run(); err != nil {
		return nil, err
	}

	return &FrameBuffer{display, authPath, xvfb}, nil
}

// Stop kills the background frame buffer process and removes the X
// authorization file.
func (f FrameBuffer) Stop() error {
	if err := f.cmd.Process.Kill(); err != nil {
		return err
	}

	os.Remove(f.AuthPath) // best effort removal; ignore error

	if err := f.cmd.Wait(); err != nil && err.Error() != "signal: killed" {
		return err
	}

	return nil
}
