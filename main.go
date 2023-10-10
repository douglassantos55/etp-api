package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Message struct {
	Token   string
	Message string
}

type Connection struct {
	Token  string
	Socket *websocket.Conn
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	events := make(chan Message)

	go func(receiver chan<- Message) {
		e := echo.New()

		e.GET("/", func(c echo.Context) error {
			events <- Message{Message: "Hello, World"}
			return c.String(http.StatusOK, "Hello, World!")
		})

		e.GET("/private", func(c echo.Context) error {
			token := c.QueryParam("token")

			events <- Message{
				Token:   token,
				Message: fmt.Sprintf("private message for %s", token),
			}

			return c.String(http.StatusOK, fmt.Sprintf("Private page for %s", token))
		})

		e.Logger.Fatal(e.Start(":1323"))
		wg.Done()
	}(events)

	go func(dispatcher <-chan Message) {
		var upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		connected := make(chan Connection)
		disconnected := make(chan string)
		sockets := make(map[string]*websocket.Conn)

		go func() {
			for {
				select {
				case message := <-dispatcher:
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
				case connection := <-connected:
					log.Printf("registering socket for user: %s\n", connection.Token)
					sockets[connection.Token] = connection.Socket
				case token := <-disconnected:
					log.Printf("socket disconnected for user: %s\n", token)
					delete(sockets, token)
				}
			}
		}()

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Print("upgrade:", err)
				return
			}

			defer c.Close()
			var userToken string

			for {
				_, token, err := c.ReadMessage()
				if err != nil {
					log.Print("read:", err)
					if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
						disconnected <- string(userToken)
						break
					}
				}
				userToken = string(token)
				connected <- Connection{Token: userToken, Socket: c}
			}
		})

		log.Fatal(http.ListenAndServe(":8080", nil))
		wg.Done()
	}(events)

	wg.Wait()
}
