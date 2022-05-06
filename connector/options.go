package connector

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type (
	// Option is a function to apply various configurations to customize a connector Component.
	Option func(o *Options)

	// Options hold the configurable parts of a connector Component.
	Options struct {
		// WriteTimeout is the maximum time of write message operation.
		// Slow client will be disconnected.
		// Default is 1 second if not set via WithWriteTimeout.
		WriteTimeout time.Duration

		// MaxMessageSize is the maximum allowed message size in bytes received from the client.
		// Default is 4096 bytes (4KB) if not set via WithMaxMessageSize.
		MaxMessageSize int64

		// Addr optionally specifies the TCP address for the server to listen on,
		// in the form "host:port". If empty, ":http" (port 80) is used.
		// See net.Dial for details of the address format.
		Addr string

		// WebsocketPath is the URL path to accept WebSocket connections.
		// This option only applies to WebsocketConnector.
		// Default is "/" if not set via WithWebsocketPath.
		WebsocketPath string

		// TLSCertFile is the path to TLS cert file.
		// This option only applies to WebsocketConnector.
		TLSCertFile string

		// TLSKeyFile is the path to TLS key file.
		// This option only applies to WebsocketConnector.
		TLSKeyFile string

		ServeMux *http.ServeMux

		Server *http.Server

		Upgrader *websocket.Upgrader
	}
)

func defaultOptions() *Options {
	return &Options{
		WebsocketPath:  "/",
		WriteTimeout:   1 * time.Second,
		MaxMessageSize: 4096,
		ServeMux:       http.DefaultServeMux,
		Server:         &http.Server{},
		Upgrader:       &websocket.Upgrader{},
	}
}

// WithAddr is an Option to set the TCP address for the server to listen on.
func WithAddr(a string) Option {
	return func(o *Options) {
		o.Addr = a
		if o.Server != nil {
			o.Server.Addr = o.Addr
		}
	}
}

// WithWebsocketPath is an Option to set the URL path for accepting WebSocket connections.
func WithWebsocketPath(p string) Option {
	return func(o *Options) {
		o.WebsocketPath = p
	}
}

// WithWriteTimeout is an Option to set the maximum time of one write message operation to complete.
func WithWriteTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = d
	}
}

// WithMaxMessageSize is an Option to set maximum message size in bytes allowed from client.
func WithMaxMessageSize(s int64) Option {
	return func(o *Options) {
		o.MaxMessageSize = s
	}
}

// WithTLSCertAndKey is an Option to set the path to TLS certificate file with its matching private key.
// WebsocketConnector will start the http.Server with ListenAndServeTLS that expects HTTPS connections,
// when either certFile or keyFile is not an empty string.
func WithTLSCertAndKey(certFile, keyFile string) Option {
	return func(o *Options) {
		o.TLSCertFile = certFile
		o.TLSKeyFile = keyFile
	}
}

// WithHTTPServeMux is an Option to set a custom http.ServeMux,
// will also update Server.Handler to mux if Options.Server is not nil.
func WithHTTPServeMux(mux *http.ServeMux) Option {
	return func(o *Options) {
		o.ServeMux = mux
		if o.Server != nil {
			o.Server.Handler = mux
		}
	}
}

// WithHTTPServer is an Option to set a custom http.Server,
// will also force overwrite Server.Addr to Options.Addr if Options.Addr is not nil,
// will also force overwrite Server.Handler to Options.ServeMux if Options.ServeMux is not nil.
func WithHTTPServer(s *http.Server) Option {
	return func(o *Options) {
		o.Server = s
		if o.Addr != "" {
			o.Server.Addr = o.Addr
		}
		if o.ServeMux != nil {
			o.Server.Handler = o.ServeMux
		}
	}
}

// WithWebsocketUpgrader is an Option to set a custom websocket.Upgrader.
func WithWebsocketUpgrader(upgrader *websocket.Upgrader) Option {
	return func(o *Options) {
		o.Upgrader = upgrader
	}
}
