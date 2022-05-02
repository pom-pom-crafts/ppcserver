package connector

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type (
	// WebsocketOption is a function to apply various configurations to customize a WebsocketConnector.
	WebsocketOption func(s *WebsocketConnector)

	// WebsocketOptions defines the configurable opts of the WebsocketConnector.
	WebsocketOptions struct {
		// Path is the URL to accept WebSocket connections.
		// Defaults to "/" if not set through WithWebsocketPath.
		Path string

		// WriteTimeout is the maximum time of write message operation.
		// Slow client will be disconnected.
		// Defaults to 1 second if not set through WithWriteTimeout.
		WriteTimeout time.Duration

		// MaxMessageSize is the maximum message size in bytes allowed from client.
		// Defaults to 65536 bytes (64KB) if not set through WithMaxMessageSize.
		MaxMessageSize int64

		// TLSCertFile is the path to TLS cert file.
		TLSCertFile string

		// TLSKeyFile is the path to TLS key file.
		TLSKeyFile string
	}
)

func defaultWebsocketOptions() *WebsocketOptions {
	return &WebsocketOptions{
		Path:           "/",
		WriteTimeout:   1 * time.Second,
		MaxMessageSize: 65536,
	}
}

// WithWebsocketPath is a WebsocketOption to set the URL Path for accepting WebSocket connections.
func WithWebsocketPath(path string) WebsocketOption {
	return func(s *WebsocketConnector) {
		s.opts.Path = path
	}
}

// WithWriteTimeout is a WebsocketOption to set the maximum time of one write message operation to complete.
func WithWriteTimeout(writeTimeout time.Duration) WebsocketOption {
	return func(s *WebsocketConnector) {
		s.opts.WriteTimeout = writeTimeout
	}
}

// WithMaxMessageSize is a WebsocketOption to set maximum message size in bytes allowed from client.
func WithMaxMessageSize(maxMessageSize int64) WebsocketOption {
	return func(s *WebsocketConnector) {
		s.opts.MaxMessageSize = maxMessageSize
	}
}

// WithTLSCertAndKey is a WebsocketOption to set the path to TLS certificate file with its matching private key.
// WebsocketConnector will start the http.Server with ListenAndServeTLS that expects HTTPS connections,
// when either certFile or keyFile is not an empty string.
func WithTLSCertAndKey(certFile, keyFile string) WebsocketOption {
	return func(s *WebsocketConnector) {
		s.opts.TLSCertFile = certFile
		s.opts.TLSKeyFile = keyFile
	}
}

// WithHTTPServeMux is a WebsocketOption to set a custom http.ServeMux of WebsocketConnector.
func WithHTTPServeMux(serveMux *http.ServeMux) WebsocketOption {
	return func(s *WebsocketConnector) {
		s.serveMux = serveMux
	}
}

// WithHTTPServer is a WebsocketOption to set a custom http.Server of WebsocketConnector.
func WithHTTPServer(server *http.Server) WebsocketOption {
	return func(s *WebsocketConnector) {
		s.server = server
	}
}

// WithWebsocketUpgrader is a WebsocketOption to set a custom websocket.Upgrader of WebsocketConnector.
func WithWebsocketUpgrader(upgrader *websocket.Upgrader) WebsocketOption {
	return func(s *WebsocketConnector) {
		s.upgrader = upgrader
	}
}
