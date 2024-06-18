package server

import (
	"context"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type GRPCServerOption interface {
	apply(*GRPCServer)
}

type fnGRPCServerOption struct {
	f func(*GRPCServer)
}

// newFnGRPCServerOption
func newFnGRPCServerOption(f func(*GRPCServer)) *fnGRPCServerOption {
	return &fnGRPCServerOption{f: f}
}

// apply
func (opt *fnGRPCServerOption) apply(gs *GRPCServer) {
	opt.f(gs)
}

// WithServiceRegister
func WithServiceRegister(fn func(s *grpc.Server)) GRPCServerOption {
	return newFnGRPCServerOption(func(gs *GRPCServer) { gs.register = fn })
}

type GRPCServer struct {
	s        *grpc.Server
	ln       net.Listener
	addr     string
	register func(s *grpc.Server)
}

// NewGRPCServer
func NewGRPCServer(opts ...GRPCServerOption) (gs *GRPCServer) {
	gs = &GRPCServer{
		s: grpc.NewServer(),
	}
	// Apply options
	for _, opt := range opts {
		opt.apply(gs)
	}

	// Option: Register
	if gs.register != nil {
		gs.register(gs.s)
	}

	return gs
}

// Start
func (gs *GRPCServer) Start(ctx context.Context, addr string) (err error) {
	return gs.StartTCP(ctx, addr)
}

// StartTCP
func (gs *GRPCServer) StartTCP(ctx context.Context, addr string) (err error) {
	gs.addr = addr

	gs.ln, err = net.Listen("tcp", gs.addr)
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	var shutdown = make(chan int)

	wg.Add(1)
	go func() {
		wg.Done()
		err = gs.s.Serve(gs.ln)
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

// TODO StartHTTP
func (gs *GRPCServer) StartHTTP(ctx context.Context, addr string) (err error) {
	return
}

// Ready
func (gs *GRPCServer) Ready() (ok bool, err error) {
	conn, err := net.Dial("tcp", gs.addr)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	return true, nil
}

// Stop
func (gs *GRPCServer) Stop(ctx context.Context) (err error) {
	gs.s.Stop()
	return gs.ln.Close()
}
