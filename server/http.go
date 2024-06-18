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

type HttpServer struct {
	s     *http.Server
	ready func() (bool, error)
}

// NewHttpServer
func NewHttpServer() (hs *HttpServer) {
	return &HttpServer{s: &http.Server{}}
}

// Start
func (hs *HttpServer) Start(ctx context.Context,
	addr string, handler http.Handler) (err error) {
	hs.s.Addr = addr
	hs.s.Handler = hs.WithHealthChecker(handler)

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
func (hs *HttpServer) WithHealthChecker(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			fmt.Fprint(w, "OK")
		} else {
			handler.ServeHTTP(w, r)
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
