package connector

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
)

// WebsocketConnector accepts WebSocket client connections,
// responsible for sending and receiving data with a WebSocket client.
type WebsocketConnector struct {
	opts      *Options
	clientsWg sync.WaitGroup
}

// NewWebsocketConnector creates a new WebsocketConnector.
func NewWebsocketConnector(opts ...Option) *WebsocketConnector {
	c := &WebsocketConnector{
		opts: defaultOptions(),
	}

	// Apply opts to customize WebsocketConnector.
	for _, opt := range opts {
		opt(c.opts)
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
	c.opts.Server.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	// HandleFunc registers the handler for processing WebSocket connection requests at opts.WebsocketPath.
	c.opts.ServeMux.HandleFunc(
		c.opts.WebsocketPath, func(w http.ResponseWriter, r *http.Request) {
			// Note: upgrader.Upgrade will reply to the client with an HTTP error when it returns an error.
			conn, err := c.opts.Upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println("ppcserver: WebsocketConnector.upgrader.Upgrade() error:", err)
				return
			}
			defer conn.Close() // Ensure the connection is closed when the current function exits.

			c.clientsWg.Add(1)
			defer c.clientsWg.Done()

			// go func() {
			// 	time.Sleep(time.Second * 3)
			//
			// 	if err := conn.WriteControl(
			// 		websocket.CloseMessage, websocket.FormatCloseMessage(3000, "hello"),
			// 		time.Now().Add(3*time.Second),
			// 	); err != nil {
			// 		log.Println("ppcserver: WriteControl() error:", err)
			// 	}
			// }()

			// SetReadLimit will close the connection when a client sends bytes larger than MaxMessageSize
			// and returns ErrReadLimit from Client.transport.Read().
			if c.opts.MaxMessageSize > 0 {
				conn.SetReadLimit(c.opts.MaxMessageSize)
			}

			// TODO, wait pong
			// conn.SetReadDeadline(time.Now().Add(0))
			// c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

			if err := StartClient(
				// Note: ctx passes in for closing the connection gracefully when the server is shutting down.
				ctx, newWebsocketTransport(
					conn,
					EncodingTypeJSON, // TODO, encodingType depends
					c.opts,
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
		err = c.opts.Server.ListenAndServeTLS(c.opts.TLSCertFile, c.opts.TLSKeyFile)
	} else {
		err = c.opts.Server.ListenAndServe()
	}
	// ErrServerClosed returns on calling http.Server.Shutdown() and does not mean ListenAndServe() fails,
	// so we return a nil error; for the other errors we return as is.
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (c *WebsocketConnector) Shutdown(ctx context.Context) error {
	if err := c.opts.Server.Shutdown(ctx); err != nil {
		return err
	}

	// Wait for all the clients' Close complete.
	c.clientsWg.Wait()
	return nil
}
