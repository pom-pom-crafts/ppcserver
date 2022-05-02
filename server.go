package ppcserver

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"os/signal"
	"syscall"
	"time"
)

type (
	// ServerOption is a function to apply various configurations to customize a Server.
	ServerOption func(s *Server)

	// ServerOptions defines the configurable opts of the Server.
	ServerOptions struct {
		// ShutdownTimeout is the maximum time for Component.Shutdown() to complete.
		// Defaults to 1 minute if not set through WithShutdownTimeout.
		ShutdownTimeout time.Duration
	}

	Component interface {
		Start(ctx context.Context) error
		Shutdown(ctx context.Context) error
	}

	Server struct {
		opts       *ServerOptions
		components []Component
	}
)

func defaultServerOptions() *ServerOptions {
	return &ServerOptions{
		ShutdownTimeout: 1 * time.Minute,
	}
}

func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		opts: defaultServerOptions(),
	}

	// Apply opts to customize Server.
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Start() {
	// The ctx.Done channel returns from signal.NotifyContext() will be closed when SIGINT/SIGTERM signal is received.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// The ctx.Done channel returns from errgroup.WithContext() will be closed when SIGINT/SIGTERM signal is received,
	// or the first time any Component.Start() method which passed to g.Go() returns a non-nil error,
	// or g.Wait() returns, whichever occurs first.
	g, ctx := errgroup.WithContext(ctx)
	for _, c := range s.components {
		// g.Go(f func() error) runs each f in a goroutine.
		g.Go(
			func() error {
				// TODO, should we recover panic and log with error here?

				// Component.Start() may block here, and its implementation should return when ctx.Done is closed.
				log.Printf("ppcserver: starting component: %T", c)
				return c.Start(ctx)
			},
		)
		g.Go(
			func() error {
				// Component.Shutdown() will not be invoked until ctx.Done is closed.
				<-ctx.Done()
				log.Printf("ppcserver: shutting down component: %T", c)

				// This goroutine returns when either Component.Shutdown() is complete before ShutdownTimeout,
				// or the ShutdownTimeout has passed.
				timeoutCtx, cancel := context.WithTimeout(context.Background(), s.opts.ShutdownTimeout)
				defer cancel()

				// Caution: buffer size should not be zero or c.Shutdown() would block forever.
				shutdownErrCh := make(chan error, 1)
				select {
				case <-timeoutCtx.Done():
					shutdownErrCh <- timeoutCtx.Err()
				case shutdownErrCh <- c.Shutdown(timeoutCtx):
				}
				return fmt.Errorf("ppcserver: %T.Shutdown() error: %w", c, <-shutdownErrCh)
			},
		)
	}
	// g.Wait() waits until all the blocking functions in g.Go() returns.
	if err := g.Wait(); err != nil {
		log.Println("ppcserver: server shutdown complete with error:", err)
	} else {
		log.Println("ppcserver: server shutdown complete")
	}
}

// WithComponent is a ServerOption to register a Component to Server.components.
func WithComponent(c Component) ServerOption {
	return func(s *Server) {
		s.components = append(s.components, c)
	}
}

// WithShutdownTimeout is a ServerOption to set the maximum time for each Component.Shutdown() to complete.
func WithShutdownTimeout(d time.Duration) ServerOption {
	return func(s *Server) {
		s.opts.ShutdownTimeout = d
	}
}
