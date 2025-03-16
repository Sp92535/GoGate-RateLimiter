// sliding_window.go
package limiter

import (
	"context"
	"net/http/httputil"
	"sync"
	"time"
)

type SlidingWindow struct {

	// track requests in current window
	curr int

	// track requests in previous window
	prev int

	// current window timestamp
	timeStamp time.Time

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
func NewSlidingWindow(capacity int, proxy *httputil.ReverseProxy) *SlidingWindow {
	ctx, cancel := context.WithCancel(context.Background())
	sw := &SlidingWindow{
		curr:      0,
		prev:      0,
		timeStamp: time.Now(),
		capacity:  capacity,
		ctx:       ctx,
		cancel:    cancel,
		proxy:     proxy,
	}

	// starting the resetting of window as a go routine once it is initalized
	go sw.reset()

	return sw
}

// core functionality of the algorithm the resetting of window
func (sw *SlidingWindow) reset() {

	// initialize ticker to tick every second
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {

		// refill as per rate
		case <-ticker.C:
			sw.mu.Lock()

			// set requests in previous window
			sw.prev = sw.curr

			// reset current window timestamp
			sw.timeStamp = time.Now()

			// reset current requests in window to 0
			sw.curr = 0

			sw.mu.Unlock()

		// returning from function if context is cancelled
		case <-sw.ctx.Done():
			return
		}
	}

}

// core functionality 2 of the algorithm calculation of dynamic window size
// function to increment requests in window and process the request
func (sw *SlidingWindow) AddRequest(req *Request) bool {

	sw.mu.Lock()

	// getting the elapsed time since latest window start
	elapsed := time.Since(sw.timeStamp).Milliseconds()

	// calculating dynamic weight as per second 
	weight := float64(1000-elapsed) / float64(1000)

	// calculating requests in current dynamic window
	reqsInCurrSildingWindow := float64(sw.prev)*weight + float64(sw.curr)

	if reqsInCurrSildingWindow < float64(sw.capacity) {
		// incrementing requests in current window
		sw.curr++

		sw.mu.Unlock()
		// serve request
		go ServeReq(sw.proxy, req, nil)

		return true
	} else {
		sw.mu.Unlock()
		return false
	}

}

// function to stop the algorithm
func (sw *SlidingWindow) Stop() {
	sw.cancel()
}
