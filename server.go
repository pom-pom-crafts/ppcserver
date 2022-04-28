package ppcserver

import (
	"context"
	"golang.org/x/sync/errgroup"
	"log"
	"os/signal"
	"syscall"
)

type (
	// ServerOption is a function to apply various configurations to customize a Server.
	ServerOption func(s *Server)

	Component interface {
		Start(ctx context.Context) error
		Shutdown() error
	}

	Server struct {
		components []Component
	}
)

func NewServer(opts ...ServerOption) *Server {
	s := &Server{}

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
				return c.Shutdown()
			},
		)
	}
	// g.Wait() waits until all the blocking functions in g.Go() returns.
	if err := g.Wait(); err != nil {
		log.Println("ppcserver: server shutdown with error:", err)
	} else {
		log.Println("ppcserver: server shutdown successfully")
	}
}

// WithComponent is a ServerOption to register a Component to Server.components.
func WithComponent(c Component) ServerOption {
	return func(s *Server) {
		s.components = append(s.components, c)
	}
}
