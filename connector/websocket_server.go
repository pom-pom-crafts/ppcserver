package connector

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type (
	// WebsocketOption is a function to apply various configurations to customize a WebsocketServer.
	WebsocketOption func(s *WebsocketServer)

	// WebsocketOptions defines the configurable opts of the WebsocketServer.
	WebsocketOptions struct {
		// Path is the URL to accept WebSocket connections.
		// Defaults to "/" if not set through WithWebsocketPath.
		Path string

		// ShutdownTimeout is the maximum time for WebsocketServer.Shutdown() to complete.
		// Defaults to 10 seconds if not set through WithShutdownTimeout.
		ShutdownTimeout time.Duration

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

	// WebsocketServer accepts WebSocket client connections,
	// responsible for sending and receiving data with a WebSocket client.
	WebsocketServer struct {
		opts        *WebsocketOptions
		serveMux    *http.ServeMux
		server      *http.Server
		upgrader    *websocket.Upgrader
		exitCh      chan os.Signal // exitCh is for receiving SIGINT/SIGTERM signals.
		serverErrCh chan error     // serverErrCh is for receiving http.ListenAndServe error.
		mu          sync.RWMutex
		clients     map[*Client]struct{} // clients hold Client in connecting.
	}
)

// NewWebsocketServer creates a new WebsocketServer.
func NewWebsocketServer(addr string, opts ...WebsocketOption) *WebsocketServer {
	websocketServer := &WebsocketServer{
		opts:        defaultWebsocketOptions(),
		exitCh:      make(chan os.Signal, 1), // Note: signal.Notify requires exitCh with buffer size of at least 1.
		serverErrCh: make(chan error, 1),
		clients:     make(map[*Client]struct{}),
	}

	// Apply opts to customize WebsocketServer.
	for _, opt := range opts {
		opt(websocketServer)
	}

	// Initialize default values when required fields of WebsocketServer are not set.
	if websocketServer.serveMux == nil {
		websocketServer.serveMux = http.DefaultServeMux
	}
	if websocketServer.server == nil {
		websocketServer.server = &http.Server{
			Addr:    addr,
			Handler: websocketServer.serveMux,
		}
	}
	if websocketServer.upgrader == nil {
		websocketServer.upgrader = &websocket.Upgrader{}
	}

	return websocketServer
}

func defaultWebsocketOptions() *WebsocketOptions {
	return &WebsocketOptions{
		Path:            "/",
		ShutdownTimeout: 10 * time.Second,
		WriteTimeout:    1 * time.Second,
		MaxMessageSize:  65536,
	}
}

// WithWebsocketPath is a WebsocketOption to set the URL Path for accepting WebSocket connections.
func WithWebsocketPath(path string) WebsocketOption {
	return func(s *WebsocketServer) {
		s.opts.Path = path
	}
}

// WithShutdownTimeout is a WebsocketOption to set the maximum time for WebsocketServer.Shutdown() to complete.
func WithShutdownTimeout(shutdownTimeout time.Duration) WebsocketOption {
	return func(s *WebsocketServer) {
		s.opts.ShutdownTimeout = shutdownTimeout
	}
}

// WithWriteTimeout is a WebsocketOption to set the maximum time of one write message operation to complete.
func WithWriteTimeout(writeTimeout time.Duration) WebsocketOption {
	return func(s *WebsocketServer) {
		s.opts.WriteTimeout = writeTimeout
	}
}

// WithMaxMessageSize is a WebsocketOption to set maximum message size in bytes allowed from client.
func WithMaxMessageSize(maxMessageSize int64) WebsocketOption {
	return func(s *WebsocketServer) {
		s.opts.MaxMessageSize = maxMessageSize
	}
}

// WithTLSCertAndKey is a WebsocketOption to set the path to TLS certificate file with its matching private key.
// WebsocketServer will start the http.Server with ListenAndServeTLS that expects HTTPS connections,
// when either certFile or keyFile is not an empty string.
func WithTLSCertAndKey(certFile, keyFile string) WebsocketOption {
	return func(s *WebsocketServer) {
		s.opts.TLSCertFile = certFile
		s.opts.TLSKeyFile = keyFile
	}
}

// WithHTTPServeMux is a WebsocketOption to set a custom http.ServeMux of WebsocketServer.
func WithHTTPServeMux(serveMux *http.ServeMux) WebsocketOption {
	return func(s *WebsocketServer) {
		s.serveMux = serveMux
	}
}

// WithHTTPServer is a WebsocketOption to set a custom http.Server of WebsocketServer.
func WithHTTPServer(server *http.Server) WebsocketOption {
	return func(s *WebsocketServer) {
		s.server = server
	}
}

// WithWebsocketUpgrader is a WebsocketOption to set a custom websocket.Upgrader of WebsocketServer.
func WithWebsocketUpgrader(upgrader *websocket.Upgrader) WebsocketOption {
	return func(s *WebsocketServer) {
		s.upgrader = upgrader
	}
}

// Start http server and block until exit signal or server error is received.
func (s *WebsocketServer) Start() {
	defer s.Shutdown()

	s.serveMux.Handle(s.opts.Path, s)

	go func() {
		if s.opts.TLSCertFile != "" || s.opts.TLSKeyFile != "" {
			s.serverErrCh <- s.server.ListenAndServeTLS(s.opts.TLSCertFile, s.opts.TLSKeyFile)
		} else {
			s.serverErrCh <- s.server.ListenAndServe()
		}
	}()

	s.blockUntilExitSignalOrServerError()
}

// blockUntilExitSignalOrServerError will block until receive exit signal or http.ListenAndServe returns with an error.
func (s *WebsocketServer) blockUntilExitSignalOrServerError() {
	// Listen and send to exitCh when SIGINT/SIGTERM signal is received.
	signal.Notify(s.exitCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case exitSig := <-s.exitCh:
		log.Println("ppcserver: WebsocketServer.Start() exit due to the signal:", exitSig)
	case err := <-s.serverErrCh:
		log.Println("ppcserver: WebsocketServer.Start() exit due to http.ListenAndServe() error:", err)
	}
}

func (s *WebsocketServer) Shutdown() {
	log.Println("ppcserver: WebsocketServer.Shutdown() begin")

	// TODO, do we need to add timeout ?
	timeoutCtx, cancel := context.WithTimeout(context.Background(), s.opts.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(timeoutCtx); err != nil {
		log.Println("ppcserver: WebsocketServer.server.Shutdown() error:", err)
	}

	// TODO, should we call Close() after Shutdown() ?
	// _ = s.server.Close()

	log.Println("ppcserver: WebsocketServer.Shutdown() complete")
}

func (s *WebsocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)

	// If the upgrade fails, then Upgrade replies to the client with an HTTP error.
	if err != nil {
		log.Println("ppcserver: WebsocketServer.upgrader.Upgrade() error:", err)
		return
	}

	transportOpts := &websocketTransportOptions{
		encodingType:   EncodingTypeJSON,
		writeTimeout:   s.opts.WriteTimeout,
		maxMessageSize: s.opts.MaxMessageSize,
	}
	transport := newWebsocketTransport(conn, transportOpts)
	client := NewClient(transport)
	s.addClient(client)

	// Since per connection support only one concurrent reader and one concurrent writer,
	// we execute all writes from the `writeLoop` goroutine and all reads from the `readLoop` goroutine.
	// Reference https://pkg.go.dev/github.com/gorilla/websocket#hdr-Concurrency for the concurrency usage details.
	go transport.writeLoop()
	go transport.readLoop()
}

// addClient add Client into connections registry clients.
func (s *WebsocketServer) addClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[client] = struct{}{}
}

func (s *WebsocketServer) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, client)
}
