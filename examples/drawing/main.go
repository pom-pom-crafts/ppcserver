package main

import (
	"github.com/pom-pom-crafts/ppcserver/connector"
	"log"
)

func main() {
	log.Println("before start")
	connector.StartWSServer(":8080")
	// if err != nil {
	// 	log.Fatalln("connector.StartWSServer() fail", err)
	// }

	// defer wsServer.Shutdown()
	// wsServer.Start()

	// var exit chan int
	// <-exit
}
