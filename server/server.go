package server

import (
	"fmt"
	"net/http"

	"github.com/gookit/validate"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	Validator struct{}

	Message struct {
		Token   string
		Message string
	}

	Connection struct {
		Token  string
		Socket *websocket.Conn
	}
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (v *Validator) Validate(i any) error {
	validation := validate.Struct(i)
	if !validation.Validate() {
		return echo.NewHTTPError(http.StatusBadRequest, string(validation.Errors.JSON()))
	}
	return nil
}

func NewServer() *echo.Echo {
	e := echo.New()

	e.Use(middleware.CORS())
	e.Use(middleware.Logger())

	e.Validator = &Validator{}

	return e
}

func Greeting(events chan<- Message) echo.HandlerFunc {
	return func(c echo.Context) error {
		events <- Message{Message: "Hello, World"}
		return c.String(http.StatusOK, "Hello, World!")
	}
}

func Private(events chan<- Message) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.QueryParam("token")
		events <- Message{
			Token:   token,
			Message: fmt.Sprintf("private message for %s", token),
		}
		return c.String(http.StatusOK, fmt.Sprintf("Private page for %s", token))
	}
}

func WS(connections chan<- Connection, disconnections chan<- string) echo.HandlerFunc {
	return func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		defer ws.Close()
		var userToken string

		for {
			_, token, err := ws.ReadMessage()
			if err != nil {
				disconnections <- string(userToken)
				break
			}
			userToken = string(token)
			connections <- Connection{Token: userToken, Socket: ws}
		}
		return nil
	}
}
