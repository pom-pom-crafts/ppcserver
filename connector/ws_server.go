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
	var upgrader *websocket.Upgrader
	// TODO, one can pass customized upgrader from options
	upgrader = &websocket.Upgrader{}

	wsServer := &WSServer{
		addr:     addr,
		upgrader: upgrader,
	}

	// http.ListenAndServe is blocked until error is returned, so we must invoke wsServer.Start() in goroutine.
	wsServer.Start()
}

type WSServer struct {
	addr     string
	upgrader *websocket.Upgrader
}

func (s *WSServer) Start() {
	defer s.Shutdown()

	http.Handle("/", s)

	go func() {
		// TODO, notify Shutdown() when err is returned
		if err := http.ListenAndServe(s.addr, nil); err != nil {
			log.Fatal("http.ListenAndServe() fail, err=", err)
		}
	}()

	exitSig := s.blockUntilExitSignal()
	log.Println("WSServer.Start() exit due to the signal:", exitSig)
}

func (s *WSServer) blockUntilExitSignal() os.Signal {
	// Note: signal.Notify requires exitCh with buffer size of at least 1.
	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return <-exitCh
}

func (s *WSServer) Shutdown() {
	// TODO, graceful shutdown logic
	log.Println("WSServer.Shutdown() to be implemented")
}

func (s *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO, handle websocket upgrade
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
