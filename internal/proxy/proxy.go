// proxy.go
package proxy

import (
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Sp92535/GoGate-RateLimiter/internal/limiter"
	"github.com/Sp92535/GoGate-RateLimiter/internal/utils"
)

// function to initialize new reverse proxy for a target url
func NewReverseProxy(target *url.URL) *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(target)
}

// function to handle proxy request
func ProxyRequestHandler(proxy *httputil.ReverseProxy, url *url.URL, endpoint string, limiters map[string]limiter.Limiter) func(http.ResponseWriter, *http.Request) {

	// return function expected by http handler
	return func(w http.ResponseWriter, r *http.Request) {

		algo, exists := limiters[r.Method]
		if !exists {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		log.Printf("Request recieved at %s\n", endpoint)

		// update headers to insure proper routing to the desired url
		r.URL.Host = url.Host
		r.URL.Scheme = url.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = url.Host

		// trimming the redundant endpoint
		path := r.URL.Path
		r.URL.Path = strings.TrimPrefix(path, endpoint)

		// initializing new request
		req := limiter.NewRequest(rand.Intn(50), w, r)

		// attempting to add new request in queue
		if !algo.AddRequest(req) {
			log.Printf("Request Throttled")
			// returning error due to too may requests
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// waiting for closure of connection
		select {

		// connection closed by server
		case <-req.Ctx.Done():
			return

		// connection closed by client
		case <-r.Context().Done():
			log.Println("Client disconnected while waiting in queue")
			return

		}
	}
}

// function to initialize and run all proxies
func Run() {

	// load config from yaml
	config := utils.NewConfiguration("config/config.yaml")

	// initializing a new router
	mux := http.NewServeMux()

	// struturing the server address
	address := config.Server.Host + ":" + config.Server.Port

	// initializing server
	srv := http.Server{
		Addr:    address,
		Handler: mux,
	}

	// looping through all the endpoints to set proxies
	for _, resource := range config.Resources {

		// parsing the target url
		url, err := url.Parse(resource.DestinationURL)
		if err != nil {
			log.Fatalf("Invalid URL: %v", err)
		}

		// creating a new reverse proxy
		proxy := NewReverseProxy(url)

		// initializing limiters
		limiters := make(map[string]limiter.Limiter)
		for method, rateLimit := range resource.RateLimits {
			limiters[method] = limiter.NewTokenBucket(rateLimit.Capacity, rateLimit.RefillRatePerSecond, proxy)
		}

		// handling the proxy
		mux.HandleFunc(resource.Endpoint, ProxyRequestHandler(proxy, url, resource.Endpoint, limiters))
	}

	log.Printf("Server started at %s", address)

	// starting the server
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("unable to start server %v", err)
	}
}
