package connector

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func StartWSServer(url string /* TODO, options ...Option */) (*WSServer, error) {
	var upgrader *websocket.Upgrader
	// TODO, one can pass customized upgrader from options
	upgrader = &websocket.Upgrader{}

	wsServer := &WSServer{
		upgrader,
	}
	wsServer.Start()

	return wsServer, nil
}

type WSServer struct {
	upgrader *websocket.Upgrader
}

func (s *WSServer) Start() {
	defer s.Shutdown()

	http.Handle("/", s)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("http.ListenAndServe() fail", err)
	}
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
