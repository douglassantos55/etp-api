package main

import (
	"api/repository"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Message struct {
	Token   string
	Message string
}

type Connection struct {
	Token  string
	Socket *websocket.Conn
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	events := make(chan Message)
	connections := make(chan Connection)
	disconnections := make(chan string)

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		events <- Message{Message: "Hello, World"}
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/warehouse", func(c echo.Context) error {
		items, err := repository.GetInventory(1)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, items)
	})

	e.GET("/private", func(c echo.Context) error {
		token := c.QueryParam("token")

		events <- Message{
			Token:   token,
			Message: fmt.Sprintf("private message for %s", token),
		}

		return c.String(http.StatusOK, fmt.Sprintf("Private page for %s", token))
	})

	e.GET("/ws", func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Print("upgrade: ", err)
			return err
		}

		defer ws.Close()
		var userToken string

		for {
			_, token, err := ws.ReadMessage()
			if err != nil {
				log.Print("read: ", err)
				disconnections <- string(userToken)
				break
			}
			userToken = string(token)
			connections <- Connection{Token: userToken, Socket: ws}
		}

		return nil
	})

	go processEvents(events, connections, disconnections)

	e.Logger.Fatal(e.Start(":1323"))
}

func processEvents(events <-chan Message, connections <-chan Connection, disconnections <-chan string) {
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
