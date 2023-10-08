package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	events := make(chan string)

	go func(receiver chan<- string) {
		e := echo.New()

		e.GET("/", func(c echo.Context) error {
			receiver <- "Hello, World!"
			return c.String(http.StatusOK, "Hello, World!")
		})

		e.Logger.Fatal(e.Start(":1323"))
		wg.Done()
	}(events)

	go func(dispatcher <-chan string) {
		var upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Print("upgrade:", err)
				return
			}

			defer c.Close()

			for message := range dispatcher {
				go func(message string) {
					time.Sleep(4 * time.Second)
					log.Printf("recv: %s", message)
					err = c.WriteMessage(websocket.TextMessage, []byte(message))
					if err != nil {
						log.Println("write:", err)
					}
				}(message)
			}
		})

		log.Fatal(http.ListenAndServe(":8080", nil))
		wg.Done()
	}(events)

	wg.Wait()
}
