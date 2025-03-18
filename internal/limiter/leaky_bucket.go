// leaky_bucket.go
package limiter

import (
	"context"
	"net/http/httputil"
	"time"

	"github.com/Sp92535/GoGate-RateLimiter/internal/utils"
)

type LeakyBucket struct {

	// request queue
	// use of channels is much better one of the reason is thread safety so there is no need of mutex
	queue chan *Request

	// queue capacity
	capacity int

	// number of requests served per unit time
	noOfRequests int

	// time duration unit
	interval time.Duration

	// context for closure
	ctx    context.Context
	cancel context.CancelFunc

	// corresponding proxy
	proxy *httputil.ReverseProxy
}

// constructor to initialize leaky bucket
func NewLeakyBucket(rateLimit *utils.RateLimit, proxy *httputil.ReverseProxy) Limiter {
	ctx, cancel := context.WithCancel(context.Background())
	lb := &LeakyBucket{
		queue:        make(chan *Request, rateLimit.Capacity),
		capacity:     rateLimit.Capacity,
		ctx:          ctx,
		cancel:       cancel,
		proxy:        proxy,
		noOfRequests: rateLimit.NoOfRequests,
		interval:     rateLimit.TimeDuration,
	}

	// starting the dripping of bucket as a go routine once it is initalized
	go lb.drip()

	return lb
}

// core functionality of the algorithm the dripping of bucket
func (lb *LeakyBucket) drip() {

	// initialize ticker to tick every second
	ticker := time.NewTicker(lb.interval)
	defer ticker.Stop()

	// following worker pool pattern
	limit := min(100, lb.noOfRequests)
	worker := make(chan struct{}, limit)

	for {
		select {

		// dripping as per rate
		case <-ticker.C:

			for range lb.noOfRequests {

				select {
				// getting the first request in queue
				case req := <-lb.queue:

					// acquire slot
					worker <- struct{}{}
					// serve request
					go ServeReq(lb.proxy, req, worker)

				default:
					continue
				}
			}

		// returning from function if context is cancelled
		case <-lb.ctx.Done():
			return
		}
	}

}

// function to add request to queue
func (lb *LeakyBucket) AddRequest(req *Request) bool {

	// adding the request to queue if space available
	select {
	case lb.queue <- req:
		return true
	default:
		return false
	}
}

// function to stop the algorithm
func (lb *LeakyBucket) Stop() {
	lb.cancel()
}
