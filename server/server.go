package server

import (
	"net/http"

	"github.com/gookit/validate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Validator struct{}

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
