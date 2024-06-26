package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

var (
	ErrInvalidCertOrKey = errors.New("invalid cert or key")
)

type HTTPServerOption interface {
	apply(*HTTPServer)
}

type fnHTTPServerOption struct {
	f func(*HTTPServer)
}

// newFnHTTPServerOption
func newFnHTTPServerOption(f func(*HTTPServer)) *fnHTTPServerOption {
	return &fnHTTPServerOption{f: f}
}

// apply
func (opt *fnHTTPServerOption) apply(hs *HTTPServer) {
	opt.f(hs)
}

// WithHandler
func WithHandler(handler http.Handler) HTTPServerOption {
	return newFnHTTPServerOption(func(hs *HTTPServer) { hs.s.Handler = handler })
}

// WithTLS
func WithTLS(certFile, keyFile string) HTTPServerOption {
	return newFnHTTPServerOption(func(hs *HTTPServer) { hs.cert, hs.key = certFile, keyFile })
}

type HTTPServer struct {
	s    *http.Server
	s2   *http2.Server
	cert string
	key  string
}

// NewHTTPServer
func NewHTTPServer(opts ...HTTPServerOption) (hs *HTTPServer) {
	hs = &HTTPServer{
		s: &http.Server{},
	}
	// Apply option
	for _, opt := range opts {
		opt.apply(hs)
	}

	return hs
}

// Start
func (hs *HTTPServer) Start(ctx context.Context, addr string) (err error) {
	hs.s.Addr = addr

	// Option: Addr
	if hs.s.Addr == "" {
		hs.s.Addr = ":80"
	}

	// Option: TLS
	if (hs.cert != "" && hs.key == "") ||
		(hs.cert == "" && hs.key != "") {
		return ErrInvalidCertOrKey
	}
	if hs.cert != "" {
		hs.s2 = &http2.Server{}
	}

	var wg sync.WaitGroup
	var shutdown = make(chan int)

	wg.Add(1)
	go func() {
		wg.Done()
		if hs.s2 != nil {
			err = http2.ConfigureServer(hs.s, hs.s2)
		}
		if err == nil {
			if hs.cert == "" {
				err = hs.s.ListenAndServe()
			} else {
				err = hs.s.ListenAndServeTLS(hs.cert, hs.key)
			}
		}
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

// Ready
func (hs *HTTPServer) Ready() (ok bool, err error) {
	conn, err := net.Dial("tcp", hs.s.Addr)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	return true, nil
}

// Stop
func (hs *HTTPServer) Stop(ctx context.Context) (err error) {
	defer func() {
		hs.s.Close()
	}()
	return hs.s.Shutdown(ctx)
}
