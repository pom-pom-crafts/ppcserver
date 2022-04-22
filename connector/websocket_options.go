package connector

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type (
	// WebSocketOption is a function to apply various configurations to customize a WebSocketServer.
	WebSocketOption func(s *WebSocketServer)

	// WebSocketOptions defines the configurable options of the WebSocketServer.
	WebSocketOptions struct {
		// Path is the URL to accept WebSocket connections.
		// Defaults to "/" if not set through WithWebSocketPath.
		Path string

		// ShutdownTimeout sets the maximum time for WebSocketServer.Shutdown() to complete.
		// Defaults to 10 seconds if not set through WithShutdownTimeout.
		ShutdownTimeout time.Duration
	}
)

func defaultWebSocketOptions() *WebSocketOptions {
	return &WebSocketOptions{
		Path:            "/",
		ShutdownTimeout: 10 * time.Second,
	}
}

// WithWebSocketPath is a WebSocketOption to set the URL Path for accepting WebSocket connections.
func WithWebSocketPath(path string) WebSocketOption {
	return func(s *WebSocketServer) {
		s.options.Path = path
	}
}

// WithHTTPServeMux is a WebSocketOption to set a custom http.ServeMux of WebSocketServer.
func WithHTTPServeMux(serveMux *http.ServeMux) WebSocketOption {
	return func(s *WebSocketServer) {
		s.serveMux = serveMux
	}
}

// WithHTTPServer is a WebSocketOption to set a custom http.Server of WebSocketServer.
func WithHTTPServer(server *http.Server) WebSocketOption {
	return func(s *WebSocketServer) {
		s.server = server
	}
}

// WithWebSocketUpgrader is a WebSocketOption to set a custom websocket.Upgrader of WebSocketServer.
func WithWebSocketUpgrader(upgrader *websocket.Upgrader) WebSocketOption {
	return func(s *WebSocketServer) {
		s.upgrader = upgrader
	}
}

// WithShutdownTimeout is a WebSocketOption to set the maximum time for WebSocketServer.Shutdown() to complete.
func WithShutdownTimeout(shutdownTimeout time.Duration) WebSocketOption {
	return func(s *WebSocketServer) {
		s.options.ShutdownTimeout = shutdownTimeout
	}
}
