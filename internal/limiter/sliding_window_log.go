// sliding_window_log.go
package limiter

import (
	"context"
	"log"
	"net/http/httputil"
	"time"

	"github.com/Sp92535/GoGate-RateLimiter/internal/utils"
	"github.com/google/uuid"
)

type SlidingWindowLog struct {

	// key to track current sliding window
	key string

	// no of request in current window 
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
func NewSlidingWindowLog(rateLimit *utils.RateLimit, proxy *httputil.ReverseProxy) Limiter {
	ctx, cancel := context.WithCancel(context.Background())
	swl := &SlidingWindowLog{
		key:          uuid.NewString(),
		ctx:          ctx,
		cancel:       cancel,
		proxy:        proxy,
		noOfRequests: rateLimit.NoOfRequests,
		interval:     rateLimit.TimeDuration,
	}

	// starting the removal of logs from window as a go routine once it is initalized
	go swl.removeLogs()

	return swl
}

// core functionality of the algorithm the removal of logs from window
func (swl *SlidingWindowLog) removeLogs() {

	for {
		select {
		// returning from function if context is cancelled
		case <-swl.ctx.Done():
			return
		// removing the expired log
		default:
			// get front of the queue
			front, err := Scripts["SLIDING-WINDOW-LOG"].Run(swl.ctx, Rdb, []string{swl.key}, "core").Int()
			if err != nil {
				log.Printf("Error :%v", err)
			}
			
			// convert to unix milli
			t := time.UnixMilli(int64(front))
			
			// check the target time
			targetTime := t.Add(swl.interval)

			// sleep untill target time
			time.Sleep(time.Until(targetTime))
		}
	}

}

// function to increment requests in window and process the request
func (swl *SlidingWindowLog) AddRequest(req *Request) bool {
	// chek if request is permitted
	res, err := Scripts["SLIDING-WINDOW-LOG"].Run(swl.ctx, Rdb, []string{swl.key}, "take", swl.noOfRequests).Int()
	if err != nil {
		log.Printf("Error :%v", err)
	}

	if res == 1 {
		// serve request
		go ServeReq(swl.proxy, req, nil)
		return true
	} else {
		return false

	}

}

// function to stop the algorithm
func (swl *SlidingWindowLog) Stop() {
	swl.cancel()
}
