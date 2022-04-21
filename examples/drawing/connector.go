package main

import (
	"github.com/pom-pom-crafts/ppcserver/connector"
	"log"
	"net/http"
)

func main() {
	// Serve client.html in root path.
	http.HandleFunc(
		"/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "client.html")
		},
	)
	log.Println("Open client.html through: http://localhost:8080")

	connector.NewWebSocketServer(
		":8080",
		connector.WithWebSocketPath("/ws"),
	).Start()
}
