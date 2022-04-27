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
		ShouldStartInGoroutine() bool
		Start(ctx context.Context) error
		Shutdown()
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
	// close ctx.Done() channel when SIGINT/SIGTERM signal is received.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)
	for _, c := range s.components {
		if c.ShouldStartInGoroutine() {
			g.Go(
				func() error {
					return c.Start(ctx)
				},
			)
		}
	}

	if err := g.Wait(); err != nil {
		log.Println("ppcserver: Server.Start() returns with error:", err)
	} else {
		log.Println("ppcserver: Server.Start() returns")
	}
}

// WithComponent is a ServerOption to register a Component to Server.
func WithComponent(c Component) ServerOption {
	return func(s *Server) {
		s.components = append(s.components, c)
	}
}
