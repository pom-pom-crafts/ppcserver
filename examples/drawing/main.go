package main

import (
	"github.com/pom-pom-crafts/ppcserver/connector"
	"log"
)

func main() {
	log.Println("before start")

	connector.StartWebSocketServer(
		":8080",
		connector.WithWebSocketPath("/"),
	)
	// if err != nil {
	// 	log.Fatalln("connector.StartWebSocketServer() fail", err)
	// }

	// defer wsServer.Shutdown()
	// wsServer.Start()

	// var exit chan int
	// <-exit
}
