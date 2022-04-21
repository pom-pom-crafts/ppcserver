package connector

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type (
	websocketOption func(s *WebSocketServer)

	// websocketOptions defines the configurable options for starting a WebSocketServer.
	websocketOptions struct {
		// path is the URL to accept WebSocket connections.
		// Defaults to "/" if not set through WithWebSocketPath.
		path string

		// shutdownTimeout sets the maximum time for WebSocketServer.Shutdown() to complete.
		// Defaults to 10 seconds if not set through WithShutdownTimeout.
		shutdownTimeout time.Duration
	}
)

func defaultWebSocketOptions() *websocketOptions {
	return &websocketOptions{
		path:            "/",
		shutdownTimeout: 10 * time.Second,
	}
}

// WithWebSocketPath is a websocketOption to set the URL path for accepting WebSocket connections.
func WithWebSocketPath(path string) websocketOption {
	return func(s *WebSocketServer) {
		s.options.path = path
	}
}

// WithHTTPServeMux is a websocketOption to set a custom http.ServeMux of WebSocketServer.
func WithHTTPServeMux(serveMux *http.ServeMux) websocketOption {
	return func(s *WebSocketServer) {
		s.serveMux = serveMux
	}
}

// WithHTTPServer is a websocketOption to set a custom http.Server of WebSocketServer.
func WithHTTPServer(server *http.Server) websocketOption {
	return func(s *WebSocketServer) {
		s.server = server
	}
}

// WithWebSocketUpgrader is a websocketOption to set a custom websocket.Upgrader of WebSocketServer.
func WithWebSocketUpgrader(upgrader *websocket.Upgrader) websocketOption {
	return func(s *WebSocketServer) {
		s.upgrader = upgrader
	}
}

// WithShutdownTimeout is a websocketOption to set the maximum time for WebSocketServer.Shutdown() to complete.
func WithShutdownTimeout(shutdownTimeout time.Duration) websocketOption {
	return func(s *WebSocketServer) {
		s.options.shutdownTimeout = shutdownTimeout
	}
}
