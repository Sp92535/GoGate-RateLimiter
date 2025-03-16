// leaky_bucket.go
package limiter

import (
	"context"
	"net/http/httputil"
	"time"
)

type LeakyBucket struct {

	// request queue
	// use of channels is much better one of the reason is thread safety so there is no need of mutex
	queue chan *Request

	// queue capacity
	capacity int

	// number of requests served per second
	emptyRatePerSecond int

	// context for closure
	ctx    context.Context
	cancel context.CancelFunc

	// corresponding proxy
	proxy *httputil.ReverseProxy
}

// constructor to initialize leaky bucket
func NewLeakyBucket(capacity int, emptyRatePerSecond int, proxy *httputil.ReverseProxy) *LeakyBucket {
	ctx, cancel := context.WithCancel(context.Background())
	lb := &LeakyBucket{
		queue:              make(chan *Request, capacity),
		capacity:           capacity,
		emptyRatePerSecond: emptyRatePerSecond,
		ctx:                ctx,
		cancel:             cancel,
		proxy:              proxy,
	}

	// starting the dripping of bucket as a go routine once it is initalized
	go lb.drip()

	return lb
}

// core functionality of the algorithm the dripping of bucket
func (lb *LeakyBucket) drip() {

	// initialize ticker to tick every second
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// following worker pool pattern
	worker := make(chan struct{}, lb.emptyRatePerSecond)

	for {
		select {

		// dripping as per rate
		case <-ticker.C:

			for range lb.emptyRatePerSecond {

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
