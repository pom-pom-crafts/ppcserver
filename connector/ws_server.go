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

func StartWSServer(addr string /* TODO, options ...Option */) {
	serveMux := http.DefaultServeMux

	var server *http.Server
	server = &http.Server{
		Addr:    addr,
		Handler: serveMux,
	}

	var upgrader *websocket.Upgrader
	// TODO, one can pass customized upgrader from options
	upgrader = &websocket.Upgrader{}

	wsServer := &WSServer{
		server:      server,
		serveMux:    serveMux,
		upgrader:    upgrader,
		exitCh:      make(chan os.Signal, 1), // Note: signal.Notify requires exitCh with buffer size of at least 1.
		serverErrCh: make(chan error, 1),
	}

	wsServer.Start()
}

type WSServer struct {
	server      *http.Server
	serveMux    *http.ServeMux
	upgrader    *websocket.Upgrader
	exitCh      chan os.Signal // For receiving SIGINT/SIGTERM signals.
	serverErrCh chan error     // For receiving http.ListenAndServe error.
}

func (s *WSServer) Start() {
	defer s.Shutdown()

	// TODO, custom pattern
	s.serveMux.Handle("/", s)

	go func() {
		s.serverErrCh <- s.server.ListenAndServe()
	}()

	s.blockUntilExitSignalOrServerError()
}

// blockUntilExitSignalOrServerError will block until receive exit signal or http.ListenAndServe returns with an error.
func (s *WSServer) blockUntilExitSignalOrServerError() {
	// Listen and send to exitCh when SIGINT/SIGTERM signal is received.
	signal.Notify(s.exitCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case exitSig := <-s.exitCh:
		log.Println("WSServer.Start() exit due to the signal:", exitSig)
	case err := <-s.serverErrCh:
		log.Println("WSServer.Start() exit due to http.ListenAndServe() fail with err:", err)
	}
}

func (s *WSServer) Shutdown() {
	log.Println("WSServer.Shutdown() begin")

	// TODO, do we need to add timeout ?
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(timeoutCtx); err != nil {
		log.Println("WSServer.server.Shutdown() fail with err:", err)
	}

	// TODO, should we call Close() after Shutdown() ?
	// _ = s.server.Close()

	log.Println("WSServer.Shutdown() complete")
}

func (s *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)

	// Log then return when Upgrade failed.
	if err != nil {
		log.Println("WSServer.upgrader.Upgrade() fail", err)
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
