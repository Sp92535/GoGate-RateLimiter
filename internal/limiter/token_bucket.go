// token_bucket.go
package limiter

import (
	"context"
	"log"
	"net/http/httputil"
	"time"

	"github.com/Sp92535/GoGate-RateLimiter/internal/utils"
	"github.com/google/uuid"
)

type TokenBucket struct {

	// track current tokens in bucket this is a unique key for redis
	key string

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
}

// constructor to initialize token bucket
func NewTokenBucket(rateLimit *utils.RateLimit, proxy *httputil.ReverseProxy) Limiter {
	ctx, cancel := context.WithCancel(context.Background())
	tb := &TokenBucket{
		key:          uuid.NewString(),
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
			// update to whatever is minimum

			Scripts["TOKEN-BUCKET"].Run(tb.ctx, Rdb, []string{tb.key}, "core", tb.capacity, tb.noOfRequests).Int()

		// returning from function if context is cancelled
		case <-tb.ctx.Done():
			return
		}
	}

}

// function to take token and process the request
func (tb *TokenBucket) AddRequest(req *Request) bool {

	res, err := Scripts["TOKEN-BUCKET"].Run(tb.ctx, Rdb, []string{tb.key}, "take").Int()
	if err != nil {
		log.Println("Error:", err)
		return false
	}
	if res == 1 {
		go ServeReq(tb.proxy, req, nil)
		return true
	} else {
		return false
	}
}

// function to stop the algorithm
func (tb *TokenBucket) Stop() {
	tb.cancel()
}
