// fixed_window.go
package limiter

import (
	"context"
	"net/http/httputil"
	"sync"
	"time"
)

type FixedWindow struct {

	// track requests in current window
	curr int

	// window capacity
	capacity int

	// context for closure
	ctx    context.Context
	cancel context.CancelFunc

	// corresponding proxy
	proxy *httputil.ReverseProxy

	// mutex to avoid race conditions of tokens
	mu sync.Mutex
}

// constructor to initialize window
func NewFixedWindow(capacity int, proxy *httputil.ReverseProxy) *FixedWindow {
	ctx, cancel := context.WithCancel(context.Background())
	fw := &FixedWindow{
		curr:     0,
		capacity: capacity,
		ctx:      ctx,
		cancel:   cancel,
		proxy:    proxy,
	}

	// starting the resetting of window as a go routine once it is initalized
	go fw.reset()

	return fw
}

// core functionality of the algorithm the resetting of window
func (fw *FixedWindow) reset() {

	// initialize ticker to tick every second
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {

		// refill as per rate
		case <-ticker.C:
			fw.mu.Lock()
			// reset current requests in window to 0
			fw.curr = 0
			fw.mu.Unlock()

		// returning from function if context is cancelled
		case <-fw.ctx.Done():
			return
		}
	}

}

// function to increment requests in window and process the request
func (fw *FixedWindow) AddRequest(req *Request) bool {

	fw.mu.Lock()

	if fw.curr < fw.capacity {
		// incrementing requests in current window
		fw.curr++

		fw.mu.Unlock()
		// serve request
		go ServeReq(fw.proxy, req, nil)

		return true
	} else {
		fw.mu.Unlock()
		return false
	}

}

// function to stop the algorithm
func (fw *FixedWindow) Stop() {
	fw.cancel()
}
