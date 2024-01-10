package notification

import (
	"api/auth"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service Service, notifier Notifier) {
	group := e.Group("/notifications")

	group.GET("", func(c echo.Context) error {
		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		notifications, err := service.GetNotifications(c.Request().Context(), int64(companyId))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, notifications)
	})

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	group.GET("/ws", func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		var clientId int64
		client := &Client{Conn: ws}

		defer ws.Close()
		defer client.Close()

		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				println(err.Error())
				notifier.Disconnect(clientId)
				break
			}

			clientId, err = strconv.ParseInt(string(message), 10, 64)
			if err != nil {
				println(err.Error())
				continue
			}

			notifier.Connect(clientId, client)
		}

		return nil
	})
}
