package connector

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartWebSocketServer(addr string, options ...websocketOption) {
	websocketServer := &WebSocketServer{
		path:        "/",                     // Defaults to "/" if not set through WithWebSocketPath.
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

	// Start http server and block until exit signal or server error is received.
	websocketServer.Start()
}

type WebSocketServer struct {
	// path is the URL to accept WebSocket connections.
	// Defaults to "/" if not set through WithWebSocketPath.
	path        string
	serveMux    *http.ServeMux
	server      *http.Server
	upgrader    *websocket.Upgrader
	exitCh      chan os.Signal // For receiving SIGINT/SIGTERM signals.
	serverErrCh chan error     // For receiving http.ListenAndServe error.
}

func (s *WebSocketServer) Start() {
	defer s.Shutdown()

	s.serveMux.Handle(s.path, s)

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
		log.Println("WebSocketServer.Start() exit due to the signal:", exitSig)
	case err := <-s.serverErrCh:
		log.Println("WebSocketServer.Start() exit due to http.ListenAndServe() fail with err:", err)
	}
}

func (s *WebSocketServer) Shutdown() {
	log.Println("WebSocketServer.Shutdown() begin")

	// TODO, do we need to add timeout ?
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(timeoutCtx); err != nil {
		log.Println("WebSocketServer.server.Shutdown() fail with err:", err)
	}

	// TODO, should we call Close() after Shutdown() ?
	// _ = s.server.Close()

	log.Println("WebSocketServer.Shutdown() complete")
}

func (s *WebSocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("WebSocketServer.upgrader.Upgrade() fail", err)
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
