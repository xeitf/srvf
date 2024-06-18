package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type HttpServerOption interface {
	apply(*HttpServer)
}

type fnHttpServerOption struct {
	f func(*HttpServer)
}

// newFnHttpServerOption
func newFnHttpServerOption(f func(*HttpServer)) *fnHttpServerOption {
	return &fnHttpServerOption{f: f}
}

// apply
func (opt *fnHttpServerOption) apply(hs *HttpServer) {
	opt.f(hs)
}

// WithHandler
func WithHandler(handler http.Handler) HttpServerOption {
	return newFnHttpServerOption(func(hs *HttpServer) { hs.s.Handler = handler })
}

type HttpServer struct {
	s *http.Server
}

// NewHttpServer
func NewHttpServer(opts ...HttpServerOption) (hs *HttpServer) {
	hs = &HttpServer{
		s: &http.Server{},
	}
	// Apply option
	for _, opt := range opts {
		opt.apply(hs)
	}

	// Option: Handler
	hs.s.Handler = hs.WithHealthChecker(hs.s.Handler)

	return hs
}

// Start
func (hs *HttpServer) Start(ctx context.Context, addr string) (err error) {
	hs.s.Addr = addr

	// Option: Addr
	if hs.s.Addr == "" {
		hs.s.Addr = ":80"
	}

	var wg sync.WaitGroup
	var shutdown = make(chan int)

	wg.Add(1)
	go func() {
		wg.Done()
		err = hs.s.ListenAndServe()
		shutdown <- 1
	}()

	wg.Wait()

	select {
	// Failed to start
	case <-shutdown:
		return err
	case <-time.After(1 * time.Second):
		return nil
	}
}

// withHealth
func (hs *HttpServer) WithHealthChecker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			fmt.Fprint(w, "OK")
		} else if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}

// Ready
func (hs *HttpServer) Ready() (ok bool, err error) {
	host, port, err := net.SplitHostPort(hs.s.Addr)
	if err != nil {
		return false, err
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:%s/health", host, port))
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	return string(content) == "OK", nil
}

// Stop
func (hs *HttpServer) Stop(ctx context.Context) (err error) {
	defer func() {
		hs.s.Close()
	}()
	return hs.s.Shutdown(ctx)
}
