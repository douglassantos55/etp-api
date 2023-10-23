package server

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/pt_BR"
	uni "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	br_translations "github.com/go-playground/validator/v10/translations/pt_BR"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	Validator struct{}

	}

	ValidationErrors struct {
		Errors map[string]string `json:"errors"`
	}

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
	uni := uni.New(en.New(), pt_BR.New())
	trans, _ := uni.GetTranslator("en")
	validate := validator.New(validator.WithRequiredStructEnabled())

	en_translations.RegisterDefaultTranslations(validate, trans)
	br_translations.RegisterDefaultTranslations(validate, trans)

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		// skip if tag key says it should be ignored
		if name == "-" {
			return ""
		}
		return name
	})

	if err := validate.Struct(i); err != nil {
		errors := make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			errors[e.Field()] = e.Translate(trans)
		}
		return echo.NewHTTPError(
			http.StatusBadRequest,
			ValidationErrors{Errors: errors},
		)
	}
	return nil
}

func NewServer() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{os.Getenv("WEBAPP_ORIGIN")},
	}))

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
