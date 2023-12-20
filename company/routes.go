package company

import (
	"api/auth"
	"api/server"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service Service) *echo.Group {
	group := e.Group("/companies")

	group.GET("/:id", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		company, err := service.GetById(c.Request().Context(), companyId)
		if err != nil {
			return err
		}

		if company == nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		return c.JSON(http.StatusOK, company)
	})

	group.POST("/register", func(c echo.Context) error {
		registration := new(Registration)
		if err := c.Bind(registration); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		if err := c.Validate(registration); err != nil {
			return err
		}

		company, err := service.Register(c.Request().Context(), registration)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, company)
	})

	group.POST("/login", func(c echo.Context) error {
		var credentials Credentials

		if err := c.Bind(&credentials); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		if err := c.Validate(&credentials); err != nil {
			return err
		}

		token, err := service.Login(c.Request().Context(), credentials)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, server.ValidationErrors{
				Errors: map[string]string{"email": err.Error()},
			})
		}

		return c.JSON(http.StatusOK, map[string]string{"token": token})
	})

	group.POST("/terrains/:position", func(c echo.Context) error {
		position, err := strconv.ParseInt(c.Param("position"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		err = service.PurchaseTerrain(c.Request().Context(), companyId, int(position))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusNoContent, nil)
	})

	return group
}
