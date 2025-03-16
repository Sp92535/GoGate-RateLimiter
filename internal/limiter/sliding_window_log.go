// sliding_window_log.go
package limiter

import (
	"context"
	"net/http/httputil"
	"time"
)

type SlidingWindowLog struct {

	// track timeStamp of first request in current window
	front time.Time

	// logs of all request
	logQueue chan time.Time

	// window capacity
	capacity int

	// context for closure
	ctx    context.Context
	cancel context.CancelFunc

	// corresponding proxy
	proxy *httputil.ReverseProxy
}

// constructor to initialize window
func NewSlidingWindowLog(capacity int, proxy *httputil.ReverseProxy) *SlidingWindowLog {
	ctx, cancel := context.WithCancel(context.Background())
	swl := &SlidingWindowLog{
		front:    time.Now(),
		logQueue: make(chan time.Time, capacity),
		capacity: capacity,
		ctx:      ctx,
		cancel:   cancel,
		proxy:    proxy,
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
			swl.front = <-swl.logQueue
			// check the target time 
			targetTime := swl.front.Add(time.Second)
			// sleep untill target time
			time.Sleep(time.Until(targetTime))
		}
	}

}


// function to increment requests in window and process the request
func (swl *SlidingWindowLog) AddRequest(req *Request) bool {

	select {
	// check if any space for next log
	case swl.logQueue <- time.Now():
		// serve request
		go ServeReq(swl.proxy, req, nil)
		return true
	default:
		return false
	}

}

// function to stop the algorithm
func (swl *SlidingWindowLog) Stop() {
	swl.cancel()
}
