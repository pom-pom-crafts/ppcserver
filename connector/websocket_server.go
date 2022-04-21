package connector

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// NewWebSocketServer creates a new WebSocketServer.
func NewWebSocketServer(addr string, options ...WebSocketOption) *WebSocketServer {
	websocketServer := &WebSocketServer{
		options:     defaultWebSocketOptions(),
		exitCh:      make(chan os.Signal, 1), // Note: signal.Notify requires exitCh with buffer size of at least 1.
		serverErrCh: make(chan error, 1),
	}

	// Apply options to customize WebSocketServer.
	for _, opt := range options {
		opt(websocketServer)
	}

	// Initialize default values when required fields of WebSocketServer are not set.
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

// WebSocketServer accepts WebSocket client connections,
// responsible for sending and receiving data with a WebSocket client.
type WebSocketServer struct {
	options     *WebSocketOptions
	serveMux    *http.ServeMux
	server      *http.Server
	upgrader    *websocket.Upgrader
	exitCh      chan os.Signal // For receiving SIGINT/SIGTERM signals.
	serverErrCh chan error     // For receiving http.ListenAndServe error.
}

// Start http server and block until exit signal or server error is received.
func (s *WebSocketServer) Start() {
	defer s.Shutdown()

	s.serveMux.Handle(s.options.path, s)

	go func() {
		s.serverErrCh <- s.server.ListenAndServe()
	}()

	s.blockUntilExitSignalOrServerError()
}

// blockUntilExitSignalOrServerError will block until receive exit signal or http.ListenAndServe returns with an error.
func (s *WebSocketServer) blockUntilExitSignalOrServerError() {
	// Listen and send to exitCh when SIGINT/SIGTERM signal is received.
	signal.Notify(s.exitCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case exitSig := <-s.exitCh:
		log.Println("ppcserver: WebSocketServer.Start() exit due to the signal:", exitSig)
	case err := <-s.serverErrCh:
		log.Println("ppcserver: WebSocketServer.Start() exit due to http.ListenAndServe() error:", err)
	}
}

func (s *WebSocketServer) Shutdown() {
	log.Println("ppcserver: WebSocketServer.Shutdown() begin")

	// TODO, do we need to add timeout ?
	timeoutCtx, cancel := context.WithTimeout(context.Background(), s.options.shutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(timeoutCtx); err != nil {
		log.Println("ppcserver: WebSocketServer.server.Shutdown() error:", err)
	}

	// TODO, should we call Close() after Shutdown() ?
	// _ = s.server.Close()

	log.Println("ppcserver: WebSocketServer.Shutdown() complete")
}

func (s *WebSocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)

	// If the upgrade fails, then Upgrade replies to the client with an HTTP error.
	if err != nil {
		log.Println("ppcserver: WebSocketServer.upgrader.Upgrade() error:", err)
		return
	}

	// TODO, wrap read and write in Client.
	for {
		msgType, msgInBytes, err := conn.ReadMessage()

		if err != nil {
			log.Println("conn.ReadMessage() fail", err)
			break
		}

		log.Printf("recv: %s", msgInBytes)

		err = conn.WriteMessage(msgType, msgInBytes)
		if err != nil {
			log.Println("conn.WriteMessage() fail", err)
			break
		}
	}
}
