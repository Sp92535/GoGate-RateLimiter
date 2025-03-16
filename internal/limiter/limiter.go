// limiter.go
package limiter

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
)

// limiter interface to support all common functions of a rate limiter
type Limiter interface {
	AddRequest(*Request) bool
	Stop()
}

// request blueprint
type Request struct {
	// request id
	ID int

	// actual http request
	r *http.Request

	// http response writer
	w http.ResponseWriter

	// context for closure
	Ctx    context.Context
	cancel context.CancelFunc
}

// constructor to initialize request
func NewRequest(id int, w http.ResponseWriter, r *http.Request) *Request {
	ctx, cancel := context.WithCancel(context.Background())
	return &Request{
		ID:     id,
		r:      r,
		w:      w,
		Ctx:    ctx,
		cancel: cancel,
	}
}

func ServeReq(proxy *httputil.ReverseProxy, req *Request, worker chan struct{}) {

	// releasing the worker if present
	defer func(){ 
		if worker!=nil{
			<-worker
		}
	}()

	// skipping if client disconnects
	if req.r.Context().Err() != nil {
		log.Println("Skipping request: client disconnected")
		req.cancel()
		return
	}

	log.Printf("Redirecting to %s\n", req.r.URL)

	// serving the request through proxy
	proxy.ServeHTTP(req.w, req.r)

	// closing the request
	req.cancel()
}
