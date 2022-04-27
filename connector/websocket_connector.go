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
)

// WebsocketConnector accepts WebSocket client connections,
// responsible for sending and receiving data with a WebSocket client.
type WebsocketConnector struct {
	opts        *WebsocketOptions
	serveMux    *http.ServeMux
	server      *http.Server
	upgrader    *websocket.Upgrader
	exitSigCh   chan os.Signal // exitSigCh is for receiving SIGINT/SIGTERM signals.
	serverErrCh chan error     // serverErrCh is for receiving http.ListenAndServe error.
	mu          sync.RWMutex
	clients     map[*client]struct{} // clients hold client in connecting.
}

// NewWebsocketConnector creates a new WebsocketConnector.
func NewWebsocketConnector(addr string, opts ...WebsocketOption) *WebsocketConnector {
	c := &WebsocketConnector{
		opts:        defaultWebsocketOptions(),
		exitSigCh:   make(chan os.Signal, 1), // Note: signal.Notify requires exitSigCh with buffer size of at least 1.
		serverErrCh: make(chan error, 1),
		clients:     make(map[*client]struct{}),
	}

	// Apply opts to customize WebsocketConnector.
	for _, opt := range opts {
		opt(c)
	}

	// Initialize default values when required fields of WebsocketConnector are not set.
	if c.serveMux == nil {
		c.serveMux = http.DefaultServeMux
	}
	if c.server == nil {
		c.server = &http.Server{
			Addr:    addr,
			Handler: c.serveMux,
		}
	}
	if c.upgrader == nil {
		c.upgrader = &websocket.Upgrader{}
	}

	return c
}

// Start http server and block until exit signal or server error is received.
func (s *WebsocketConnector) Start(ctx context.Context) {
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
func (s *WebsocketConnector) blockUntilExitSignalOrServerError() {
	// Listen and send to exitSigCh when SIGINT/SIGTERM signal is received.
	signal.Notify(s.exitSigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case exitSig := <-s.exitSigCh:
		log.Println("ppcserver: WebsocketConnector.Start() exit due to the signal:", exitSig)
	case err := <-s.serverErrCh:
		log.Println("ppcserver: WebsocketConnector.Start() exit due to http.ListenAndServe() error:", err)
	}
}

func (s *WebsocketConnector) Shutdown() {
	log.Println("ppcserver: WebsocketConnector.Shutdown() begin")

	// TODO, do we need to add timeout ?
	timeoutCtx, cancel := context.WithTimeout(context.Background(), s.opts.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(timeoutCtx); err != nil {
		log.Println("ppcserver: WebsocketConnector.server.Shutdown() error:", err)
	}

	// TODO, should we call Close() after Shutdown() ?
	// _ = s.server.Close()

	log.Println("ppcserver: WebsocketConnector.Shutdown() complete")
}

func (s *WebsocketConnector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)

	// Note: upgrader.Upgrade will reply to the client with an HTTP error when it returns an err.
	if err != nil {
		log.Println("ppcserver: WebsocketConnector.upgrader.Upgrade() error:", err)
		return
	}

	// transportOpts := &websocketTransportOptions{
	// 	encodingType:   EncodingTypeJSON,
	// 	writeTimeout:   s.opts.WriteTimeout,
	// 	maxMessageSize: s.opts.MaxMessageSize,
	// }
	// client := NewClient(newWebsocketTransport(conn, transportOpts))
	// s.addClient(client)

	if s.opts.MaxMessageSize > 0 {
		conn.SetReadLimit(s.opts.MaxMessageSize)
	}

	// TODO, wait pong
	// conn.SetReadDeadline(time.Now().Add(0))
	// c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Since per connection support only one concurrent reader and one concurrent writer,
	// we execute all writes from the `writeLoop` goroutine and all reads from the `readLoop` goroutine.
	// Reference https://pkg.go.dev/github.com/gorilla/websocket#hdr-Concurrency for the concurrency usage details.
	// go client.writeLoop()
	// go client.readLoop()
}

// addClient add client into connections registry clients.
func (s *WebsocketConnector) addClient(client *client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[client] = struct{}{}
}

func (s *WebsocketConnector) removeClient(client *client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, client)
}
