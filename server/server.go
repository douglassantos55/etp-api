package server

import (
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
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	JWT_SECRET_KEY    = "JWT_SECRET"
	CLIENT_ORIGIN_KEY = "CLIENT_ORIGIN"
)

type (
	Validator struct{}

	ValidationErrors struct {
		Errors map[string]string `json:"errors"`
	}

	BusinessRuleError struct {
		Message string
	}
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func GetJwtSecret() string {
	return os.Getenv(JWT_SECRET_KEY)
}

func NewBusinessRuleError(message string) BusinessRuleError {
	return BusinessRuleError{message}
}

func (e BusinessRuleError) Error() string {
	return e.Message
}

func NewServer() *echo.Echo {
	e := echo.New()

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		he := err
		if be, ok := err.(BusinessRuleError); ok {
			he = echo.NewHTTPError(http.StatusUnprocessableEntity, be.Message)
		}
		e.DefaultHTTPErrorHandler(he, c)
	}

	// e.Use(middleware.Logger())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{os.Getenv(CLIENT_ORIGIN_KEY)},
	}))

	e.Use(echojwt.WithConfig(echojwt.Config{
		Skipper: func(c echo.Context) bool {
			isLogin := c.Request().URL.Path == "/companies/login"
			isRegister := c.Request().URL.Path == "/companies/register"
			isWebsocket := c.Request().URL.Path == "/notifications/ws"

			return isLogin || isRegister || isWebsocket
		},
		SigningKey: []byte(GetJwtSecret()),
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwt.RegisteredClaims)
		},
	}))

	e.Validator = new(Validator)

	return e
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
