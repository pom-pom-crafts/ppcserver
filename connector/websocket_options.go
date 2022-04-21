package connector

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type websocketOption func(s *WebSocketServer)

// WithWebSocketPath is a websocketOption to set the URL path for accepting WebSocket connections.
func WithWebSocketPath(path string) websocketOption {
	return func(s *WebSocketServer) {
		s.path = path
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
