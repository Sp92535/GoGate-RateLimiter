// leaky_bucket.go
package limiter

import (
	"context"
	"log"
	"net/http/httputil"
	"time"

	"github.com/Sp92535/GoGate-RateLimiter/internal/utils"
	"github.com/google/uuid"
)

type LeakyBucket struct {
	// key to track bucket
	key string

	// temperoray mapping of id -> request
	reqs map[string]*Request

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
		key:          uuid.NewString(),
		reqs:         make(map[string]*Request),
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
			// get all dripped requests
			dripped, err := Scripts["LEAKY-BUCKET"].Run(lb.ctx, Rdb, []string{lb.key}, "core", lb.noOfRequests).StringSlice()

			if err != nil {
				log.Printf("Error :%v", err)
			}
			// start serving all dripped request
			for _, id := range dripped {

				// acquire slot
				worker <- struct{}{}
				// serve request
				req, exists := lb.reqs[id]
				if !exists {
					<-worker
					continue
				}
				go ServeReq(lb.proxy, req, worker)
				delete(lb.reqs, id)
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
	res, err := Scripts["LEAKY-BUCKET"].Run(lb.ctx, Rdb, []string{lb.key}, "take", req.ID, lb.capacity).Int()
	if err != nil {
		log.Println("Error:", err)
		return false
	}
	if res == 1 {
		lb.reqs[req.ID] = req
		return true
	} else {
		return false
	}
}

// function to stop the algorithm
func (lb *LeakyBucket) Stop() {
	lb.cancel()
}
