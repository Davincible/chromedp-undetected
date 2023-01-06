//go:build unix

package chromedpundetected

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/exp/slog"
)

// Errors.
var (
	ErrXvfbNotFound = errors.New("xvfb not found. Please install (Linux only)")
)

// frameBuffer controls an X virtual frame buffer running as a background
// process.
type frameBuffer struct {
	// Display is the X11 display number that the Xvfb process is hosting
	// (without the preceding colon).
	Display string

	// AuthPath is the path to the X11 authorization file that permits X clients
	// to use the X server. This is typically provided to the client via the
	// XAUTHORITY environment variable.
	AuthPath string

	cmd *exec.Cmd
}

// newFrameBuffer starts an X virtual frame buffer running in the background.
// FrameBufferOptions may be populated to change the behavior of the frame buffer.
func newFrameBuffer(screenSize string) (*frameBuffer, error) { //nolint:funlen
	if err := exec.Command("which", "Xvfb").Run(); err != nil {
		return nil, ErrXvfbNotFound
	}

	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = pipeReader.Close(); err != nil {
			slog.Error("failed to close pipe reader", err)
		}
	}()

	authPath, err := tempFile("chromedp-xvfb")
	if err != nil {
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
	xvfb.ExtraFiles = []*os.File{pipeWriter}
	xvfb.Env = append(xvfb.Env, "XAUTHORITY="+authPath)

	if xvfb.SysProcAttr == nil {
		xvfb.SysProcAttr = new(syscall.SysProcAttr)
	}

	xvfb.SysProcAttr.Pdeathsig = syscall.SIGKILL

	if err := xvfb.Start(); err != nil {
		return nil, err
	}

	if err := pipeWriter.Close(); err != nil {
		return nil, err
	}

	type resp struct {
		display string
		err     error
	}

	ch := make(chan resp)

	go func() {
		bufr := bufio.NewReader(pipeReader)
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

	case <-time.After(10 * time.Second):
		return nil, errors.New("timeout waiting for Xvfb")
	}

	xauth := exec.Command("xauth", "generate", ":"+display, ".", "trusted") //nolint:gosec
	xauth.Env = append(xauth.Env, "XAUTHORITY="+authPath)
	// Make this conditional?
	xauth.Stderr = os.Stderr
	xauth.Stdout = os.Stdout

	if err := xauth.Run(); err != nil {
		return nil, err
	}

	return &frameBuffer{display, authPath, xvfb}, nil
}

// Stop kills the background frame buffer process and removes the X
// authorization file.
func (f frameBuffer) Stop() error {
	if err := f.cmd.Process.Kill(); err != nil {
		return err
	}

	_ = os.Remove(f.AuthPath) //nolint:errcheck

	if err := f.cmd.Wait(); err != nil && err.Error() != "signal: killed" {
		return err
	}

	return nil
}

func tempFile(pattern string) (string, error) {
	tempFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}

	fileName := tempFile.Name()

	if err := tempFile.Close(); err != nil {
		return "", err
	}

	return fileName, nil
}
