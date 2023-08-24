package chromedpundetected

import (
	"context"
	"time"

	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// IdleEvent is sent through the channel returned by NetworkIdleListener when the network is considered idle.
// This event can be used to determine when a page has finished loading.
type IdleEvent struct {
	IsIdle bool
}

// NetworkIdlePermanentListener sets up a listener to monitor for network idle events.
//
// This can be used as a proxy to know for e.g. when a page has fully loaded, assuming
// that the page doesn't send any network requests within the networkIdleTimeout period
// after is has finished loading.
//
// This function creates a listener that monitors the target specified by the given context
// for network activity. It triggers a "NETWORK IDLE" event if no network request is sent
// within the provided `networkIdleTimeout` duration after a network idle lifecycle event.
//
// After the network is considered idle, the listener will terminate and the channel will close.
//
// Parameters:
//   - ctx: The context for which the listener is set up. Usually, this context is tied to a specific browser tab or page.
//   - networkIdleTimeout: The duration to wait for no network activity after a network idle lifecycle event before considering the network to be truly idle.
//   - totalTimeout: The duration to wait for the network to be idle before terminating the listener.
//
// Returns:
// 1. A channel of type IdleEvent. When the network is considered idle, an IdleEvent with IsIdle set to true is sent through this channel.
func NetworkIdleListener(ctx context.Context, networkIdleTimeout, totalTimeout time.Duration) chan IdleEvent {
	ctx, cancel := context.WithCancel(ctx)

	ch := make(chan IdleEvent, 1) // buffer to prevent blocking

	var idleTimer *time.Timer

	go func() {
		<-time.After(totalTimeout)
		ch <- IdleEvent{IsIdle: false}

		cancel()
		close(ch)
	}()

	listener := newNetworkIdleListener(ch, networkIdleTimeout, idleTimer)

	chromedp.ListenTarget(ctx, listener)

	return ch
}

// This can be used as a proxy to know for e.g. when a page has fully loaded, assuming
// that the page doesn't send any network requests within the networkIdleTimeout period
// after is has finished loading.
//
// This function creates a listener that monitors the target specified by the given context
// for network activity. It triggers a "NETWORK IDLE" event if no network request is sent
// within the provided `networkIdleTimeout` duration after a network idle lifecycle event.
//
// It's designed to run indefinitely, i.e., it doesn't automatically stop listening after
// an idle event or after a certain period. To manually stop listening and to free up
// associated resources, one should call the returned cancel function.
//
// Parameters:
//   - ctx: The context for which the listener is set up. Usually, this context is tied to a specific browser tab or page.
//   - networkIdleTimeout: The duration to wait for no network activity after a network idle lifecycle event before considering the network to be truly idle.
//
// Returns:
//  1. A channel of type IdleEvent. When the network is considered idle, an IdleEvent
//     with IsIdle set to true is sent through this channel.
//  2. A cancel function. This function can be called to terminate the listener and close the channel.
func NetworkIdlePermanentListener(ctx context.Context, networkIdleTimeout time.Duration) (chan IdleEvent, func()) {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan IdleEvent, 1) // buffer to prevent blocking

	var idleTimer *time.Timer

	listener := newNetworkIdleListener(ch, networkIdleTimeout, idleTimer)

	chromedp.ListenTarget(ctx, listener)

	cancelFunc := func() {
		cancel()
		close(ch)
	}

	return ch, cancelFunc
}

func newNetworkIdleListener(ch chan IdleEvent, networkIdleTimeout time.Duration, idleTimer *time.Timer) func(interface{}) {
	return func(ev interface{}) {
		// Check if the event is a standard protocol message
		if _, ok := ev.(*cdproto.Message); ok {
			return
		}

		// Reset the timer every time a request is sent
		if _, ok := ev.(*network.EventRequestWillBeSent); ok {
			if idleTimer != nil {
				idleTimer.Stop()
				idleTimer = nil
			}
		}

		// Start or check the timer when the network is idle
		if ev, ok := ev.(*page.EventLifecycleEvent); ok && ev.Name == "networkIdle" {
			if idleTimer == nil {
				idleTimer = time.AfterFunc(networkIdleTimeout, func() {
					ch <- IdleEvent{IsIdle: true}
				})
			} else {
				idleTimer.Reset(networkIdleTimeout)
			}
		}
	}
}
