package connector

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"sync"
)

// WebsocketConnector accepts WebSocket client connections,
// responsible for sending and receiving data with a WebSocket client.
type WebsocketConnector struct {
	opts      *WebsocketOptions
	serveMux  *http.ServeMux
	server    *http.Server
	upgrader  *websocket.Upgrader
	clientsWg sync.WaitGroup
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

// Start starts an HTTP server for serving the WebSocket connection requests
// and block until the server is closed.
// A ctx (which will cancel when the server is shutting down) is required
// for gracefully shutting down the HTTP server and actively closing all the WebSocket connections.
func (c *WebsocketConnector) Start(ctx context.Context) error {
	// BaseContext specifies the ctx as the base context for incoming requests on this server,
	// which can be used to cancel the long-running HTTP requests and also the WebSocket connections.
	c.server.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	// HandleFunc registers the handler for processing WebSocket connection requests at opts.Path.
	c.serveMux.HandleFunc(
		c.opts.Path, func(w http.ResponseWriter, r *http.Request) {
			conn, err := c.upgrader.Upgrade(w, r, nil)

			// Note: upgrader.Upgrade will reply to the client with an HTTP error when it returns an error.
			if err != nil {
				log.Println("ppcserver: WebsocketConnector.upgrader.Upgrade() error:", err)
				return
			}

			// Ensure the connection is closed when the current function exits.
			defer conn.Close()

			if c.opts.MaxMessageSize > 0 {
				conn.SetReadLimit(c.opts.MaxMessageSize)
			}

			// TODO, wait pong
			// conn.SetReadDeadline(time.Now().Add(0))
			// c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

			c.clientsWg.Add(1)
			defer c.clientsWg.Done()

			if err := StartClient(
				// Note: ctx passes in for closing the connection gracefully when the server is shutting down.
				ctx, newWebsocketTransport(
					conn, &websocketTransportOptions{
						encodingType:   EncodingTypeJSON, // TODO, encodingType depends
						writeTimeout:   c.opts.WriteTimeout,
						maxMessageSize: c.opts.MaxMessageSize,
					},
				),
			); err != nil {
				log.Println("ppcserver: StartClient() error:", err)
			}
		},
	)

	// ListenAndServe will block until the server is closed for various reasons,
	// such as when WebsocketConnector.Shutdown() is invoked,
	// or when PORT is already in-used.
	var err error
	if c.opts.TLSCertFile != "" || c.opts.TLSKeyFile != "" {
		err = c.server.ListenAndServeTLS(c.opts.TLSCertFile, c.opts.TLSKeyFile)
	} else {
		err = c.server.ListenAndServe()
	}
	// ErrServerClosed returns on calling http.Server.Shutdown() and does not mean ListenAndServe() fails,
	// so we return a nil error; for the other errors we return as is.
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (c *WebsocketConnector) Shutdown() error {
	// TODO, do we need to add timeout to force the shutdown complete ?
	timeoutCtx, cancel := context.WithTimeout(context.Background(), c.opts.ShutdownTimeout)
	defer cancel()

	if err := c.server.Shutdown(timeoutCtx); err != nil {
		return err
	}

	// TODO, does it really need to wait for all the clients Close complete.
	c.clientsWg.Wait()
	return nil
}
