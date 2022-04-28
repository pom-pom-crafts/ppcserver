package connector

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
)

// WebsocketConnector accepts WebSocket client connections,
// responsible for sending and receiving data with a WebSocket client.
type WebsocketConnector struct {
	opts     *WebsocketOptions
	serveMux *http.ServeMux
	server   *http.Server
	upgrader *websocket.Upgrader
}

// NewWebsocketConnector creates a new WebsocketConnector.
func NewWebsocketConnector(addr string, opts ...WebsocketOption) *WebsocketConnector {
	c := &WebsocketConnector{
		opts: defaultWebsocketOptions(),
	}

	// Apply opts to customize WebsocketConnector.
	for _, opt := range opts {
		opt(c)
	}

	// Initialize default values when required fields of WebsocketConnector are not set.
	// Note: must run after all opts are applied since opts may set the required fields.
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

// Start http server and block until.
func (s *WebsocketConnector) Start(ctx context.Context) error {
	// Handle registers the WebsocketConnector.ServeHTTP handler for requests at opts.Path.
	s.serveMux.Handle(s.opts.Path, s)

	// BaseContext specifies the ctx as the base context for incoming requests on this server,
	// which can be used to cancel the long-running http requests (not includes WebSocket connections).
	s.server.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	// ListenAndServe will block or immediately returns with a non-nil error.
	var err error
	if s.opts.TLSCertFile != "" || s.opts.TLSKeyFile != "" {
		err = s.server.ListenAndServeTLS(s.opts.TLSCertFile, s.opts.TLSKeyFile)
	} else {
		err = s.server.ListenAndServe()
	}
	// ErrServerClosed returns when http.Server.Shutdown() is invoked and does not mean ListenAndServe() fails.
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *WebsocketConnector) Shutdown() error {
	// TODO, do we need to add timeout to force the shutdown complete ?
	timeoutCtx, cancel := context.WithTimeout(context.Background(), s.opts.ShutdownTimeout)
	defer cancel()

	return s.server.Shutdown(timeoutCtx)
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
