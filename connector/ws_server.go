package connector

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func StartWSServer(addr string /* TODO, options ...Option */) {
	var httpServer *http.Server
	httpServer = &http.Server{Addr: addr}

	var upgrader *websocket.Upgrader
	// TODO, one can pass customized upgrader from options
	upgrader = &websocket.Upgrader{}

	wsServer := &WSServer{
		server:      httpServer,
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

	// s.server.Handler.(*http.ServeMux).Handle("/", s)

	go func() {
		s.serverErrCh <- s.server.ListenAndServe()
	}()

	s.blockUntilExitSignalOrServerError()
}

// blockUntilExitSignalOrServerError will block until receive exit signal or http.ListenAndServe returns with an error.
func (s *WSServer) blockUntilExitSignalOrServerError() {
	// Listen to SIGINT and SIGTERM signal and send to exitCh when received.
	signal.Notify(s.exitCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case exitSig := <-s.exitCh:
		log.Println("WSServer.Start() exit due to the signal:", exitSig)
	case err := <-s.serverErrCh:
		log.Println("WSServer.Start() exit due to http.ListenAndServe() fail with err:", err)
	}
}

func (s *WSServer) Shutdown() {
	// TODO, graceful shutdown logic, such as shutdown s.server
	log.Println("WSServer.Shutdown() to be implemented")
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
