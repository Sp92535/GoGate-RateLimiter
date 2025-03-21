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

	//
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
			// wait if logQueue is empty
			front, err := Scripts["SLIDING-WINDOW-LOG"].Run(swl.ctx, Rdb, []string{swl.key}, "core").Int()
			if err != nil {
				log.Printf("Error :%v", err)
			}
			t := time.UnixMilli(int64(front))
			// check the target time
			targetTime := t.Add(swl.interval)
			if t.Compare(time.UnixMilli(0)) != 0 {
				log.Println(t.Local(), targetTime.Local())
			}
			// sleep untill target time
			time.Sleep(time.Until(targetTime))
		}
	}

}

// function to increment requests in window and process the request
func (swl *SlidingWindowLog) AddRequest(req *Request) bool {

	res, err := Scripts["SLIDING-WINDOW-LOG"].Run(swl.ctx, Rdb, []string{swl.key}, "take", swl.noOfRequests).Int()
	if err != nil {
		log.Printf("Error :%v", err)
	}
	log.Println("RESS", res)
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
