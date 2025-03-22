// proxy.go
package proxy

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Sp92535/GoGate-RateLimiter/internal/limiter"
	"github.com/Sp92535/GoGate-RateLimiter/internal/utils"
	"github.com/google/uuid"
)

// function to initialize new reverse proxy for a target url
func NewReverseProxy(target *url.URL) *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(target)
}

// function to handle proxy request
func ProxyRequestHandler(proxy *httputil.ReverseProxy, url *url.URL, endpoint string, limiters map[string]limiter.Limiter) func(http.ResponseWriter, *http.Request) {

	// return function expected by http handler
	return func(w http.ResponseWriter, r *http.Request) {
		// getting algo asper request method
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
		req := limiter.NewRequest(uuid.NewString(), w, r)

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

	// stop funcs to stop every thing at end
	var stopFunc []func()

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
			algo, exists := limiter.Limiters[rateLimit.Strategy]
			if !exists {
				log.Fatalf("no such strategy %s", rateLimit.Strategy)
			}
			limiters[method] = algo(rateLimit, proxy)
			// exiting all algos at end of Run
			stopFunc = append(stopFunc, limiters[method].Stop)
		}

		// handling the proxy
		mux.HandleFunc(resource.Endpoint, ProxyRequestHandler(proxy, url, resource.Endpoint, limiters))
	}

	log.Printf("Server started at %s", address)

	// starting the server as a separate go routine
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("unable to start server %v", err)
		}
	}()

	// graceful shutdown
	// initializing an buffered channel to listen for shutdown signal CTRL+C
	sigChan := make(chan os.Signal, 1)
	// notify the channel in specified signals
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Stop all limiters
	for _, stop := range stopFunc {
		stop()
	}

	// initializing context for timeout
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}

	log.Println("Graceful shutdown complete.")

}
