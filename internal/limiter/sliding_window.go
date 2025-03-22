// sliding_window.go
package limiter

import (
	"context"
	"log"
	"net/http/httputil"
	"time"

	"github.com/Sp92535/GoGate-RateLimiter/internal/utils"
	"github.com/google/uuid"
)

type SlidingWindow struct {
	// key to track requests
	key string

	// window noOfRequests
	noOfRequests int

	// window duration
	interval time.Duration

	// context for closure
	ctx    context.Context
	cancel context.CancelFunc

	// corresponding proxy
	proxy *httputil.ReverseProxy
}

// constructor to initialize window
func NewSlidingWindow(rateLimit *utils.RateLimit, proxy *httputil.ReverseProxy) Limiter {
	ctx, cancel := context.WithCancel(context.Background())
	sw := &SlidingWindow{
		key:          uuid.NewString(),
		ctx:          ctx,
		cancel:       cancel,
		proxy:        proxy,
		noOfRequests: rateLimit.NoOfRequests,
		interval:     rateLimit.TimeDuration,
	}

	// starting the resetting of window as a go routine once it is initalized
	go sw.reset()

	return sw
}

// core functionality of the algorithm the resetting of window
func (sw *SlidingWindow) reset() {

	// initialize ticker to tick every second
	ticker := time.NewTicker(sw.interval)
	defer ticker.Stop()

	for {
		select {

		// refill as per rate
		case <-ticker.C:
			Scripts["SLIDING-WINDOW"].Run(sw.ctx, Rdb, []string{sw.key}, "core")

		// returning from function if context is cancelled
		case <-sw.ctx.Done():
			return
		}
	}

}

// core functionality 2 of the algorithm calculation of dynamic window size
// function to increment requests in window and process the request
func (sw *SlidingWindow) AddRequest(req *Request) bool {
	// check if request can be permitted
	res, err := Scripts["SLIDING-WINDOW"].Run(sw.ctx, Rdb, []string{sw.key}, "take", sw.noOfRequests, sw.interval).Int()

	if err != nil {
		log.Println("Error:", err)
		return false
	}
	if res == 1 {

		// serve request
		go ServeReq(sw.proxy, req, nil)

		return true
	} else {
		return false
	}

}

// function to stop the algorithm
func (sw *SlidingWindow) Stop() {
	sw.cancel()
}
