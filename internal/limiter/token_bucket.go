// token_bucket.go
package limiter

import (
	"context"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/Sp92535/GoGate-RateLimiter/internal/utils"
)

type TokenBucket struct {

	// track current tokens in bucket
	tokens int

	// bucket capacity
	capacity int

	// bucket refill requests
	noOfRequests int

	// bucket refill duration
	interval time.Duration

	// context for closure
	ctx    context.Context
	cancel context.CancelFunc

	// corresponding proxy
	proxy *httputil.ReverseProxy

	// mutex to avoid race conditions of tokens
	mu sync.Mutex
}

// constructor to initialize token bucket
func NewTokenBucket(rateLimit *utils.RateLimit, proxy *httputil.ReverseProxy) Limiter {
	ctx, cancel := context.WithCancel(context.Background())
	tb := &TokenBucket{
		tokens:       rateLimit.Capacity, // starting with full capacity
		capacity:     rateLimit.Capacity,
		ctx:          ctx,
		cancel:       cancel,
		proxy:        proxy,
		noOfRequests: rateLimit.NoOfRequests,
		interval:     rateLimit.TimeDuration,
	}

	// starting the refilling of bucket as a go routine once it is initalized
	go tb.refill()

	return tb
}

// core functionality of the algorithm the refilling of bucket
func (tb *TokenBucket) refill() {

	// initialize ticker to tick every second
	ticker := time.NewTicker(tb.interval)
	defer ticker.Stop()

	for {
		select {

		// refill as per rate
		case <-ticker.C:
			tb.mu.Lock()
			// update to whatever is minimum
			tb.tokens = min(tb.tokens+tb.noOfRequests, tb.capacity)
			tb.mu.Unlock()

		// returning from function if context is cancelled
		case <-tb.ctx.Done():
			return
		}
	}

}

// function to take token and process the request
func (tb *TokenBucket) AddRequest(req *Request) bool {

	tb.mu.Lock()

	if tb.tokens > 0 {
		// decrementing token
		tb.tokens--

		tb.mu.Unlock()
		// serve request
		go ServeReq(tb.proxy, req, nil)

		return true
	} else {
		tb.mu.Unlock()
		return false
	}

}

// function to stop the algorithm
func (tb *TokenBucket) Stop() {
	tb.cancel()
}
