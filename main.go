package main

import (
	"api/database"
	"api/resource"
	"api/server"
	"api/warehouse"
	"log"

	"github.com/gorilla/websocket"
)

func main() {
	conn, err := database.GetConnection(database.SQLITE, "development.db")
	if err != nil {
		log.Fatalf("could not connect to database: %s", err)
	}

	events := make(chan server.Message)
	connections := make(chan server.Connection)
	disconnections := make(chan string)

	svr := server.NewServer()

	resource.CreateEndpoints(svr, conn)
	warehouse.CreateEndpoints(svr, conn)

	svr.GET("/", server.Greeting(events))
	svr.GET("/private", server.Private(events))
	svr.GET("/ws", server.WS(connections, disconnections))

	go processEvents(events, connections, disconnections)

	svr.Start(":1323")
}

func processEvents(events <-chan server.Message, connections <-chan server.Connection, disconnections <-chan string) {
	sockets := make(map[string]*websocket.Conn)

	for {
		select {
		case message := <-events:
			log.Printf("received event: %v\n", message)

			if message.Token != "" {
				socket, ok := sockets[message.Token]
				if ok {
					if err := socket.WriteMessage(websocket.TextMessage, []byte(message.Message)); err != nil {
						log.Println("write:", err)
					}
				} else {
					log.Printf("socket not found: %s\n", message.Token)
				}
			} else {
				for _, socket := range sockets {
					if err := socket.WriteMessage(websocket.TextMessage, []byte(message.Message)); err != nil {
						log.Println("write:", err)
					}
				}
			}
		case connection := <-connections:
			log.Printf("registering socket for user: %s\n", connection.Token)
			sockets[connection.Token] = connection.Socket
		case token := <-disconnections:
			log.Printf("socket disconnected for user: %s\n", token)
			delete(sockets, token)
		}
	}
}
