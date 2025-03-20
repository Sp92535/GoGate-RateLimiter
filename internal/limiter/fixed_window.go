// fixed_window.go
package limiter

import (
	"context"
	"log"
	"net/http/httputil"
	"time"

	"github.com/Sp92535/GoGate-RateLimiter/internal/utils"
	"github.com/google/uuid"
)

type FixedWindow struct {

	// track requests in current window
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
func NewFixedWindow(rateLimit *utils.RateLimit, proxy *httputil.ReverseProxy) Limiter {
	ctx, cancel := context.WithCancel(context.Background())
	fw := &FixedWindow{
		key:          uuid.NewString(),
		ctx:          ctx,
		cancel:       cancel,
		proxy:        proxy,
		noOfRequests: rateLimit.NoOfRequests,
		interval:     rateLimit.TimeDuration,
	}

	Rdb.Set(fw.ctx, fw.key, 0, 0)

	// starting the resetting of window as a go routine once it is initalized
	go fw.reset()

	return fw
}

// core functionality of the algorithm the resetting of window
func (fw *FixedWindow) reset() {

	// initialize ticker to tick every second
	ticker := time.NewTicker(fw.interval)
	defer ticker.Stop()

	for {
		select {

		// refill as per rate
		case <-ticker.C:
			// reset current requests in window to 0
			Scripts["FIXED-WINDOW"].Run(fw.ctx, Rdb, []string{fw.key}, "core")

		// returning from function if context is cancelled
		case <-fw.ctx.Done():
			return
		}
	}

}

// function to increment requests in window and process the request
func (fw *FixedWindow) AddRequest(req *Request) bool {

	res, err := Scripts["FIXED-WINDOW"].Run(fw.ctx, Rdb, []string{fw.key}, "take", fw.noOfRequests).Int()
	if err != nil {
		log.Println("Error:", err)
		return false
	}
	if res == 1 {
		go ServeReq(fw.proxy, req, nil)
		return true
	} else {
		return false
	}
}

// function to stop the algorithm
func (fw *FixedWindow) Stop() {
	fw.cancel()
}
