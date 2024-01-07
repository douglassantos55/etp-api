package financing

import (
	"api/auth"
	"api/company"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, financingSvc Service, companySvc company.Service) *echo.Group {
	group := e.Group("/financing")

	group.GET("/rates", func(c echo.Context) error {
		rates, err := financingSvc.GetEffectiveRates(c.Request().Context())
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, rates)
	})

	group.POST("/rates", func(c echo.Context) error {
		id, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		if id != 0 {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		ctx := c.Request().Context()
		rates, err := financingSvc.CalculateRates(ctx)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, rates)
	})

	return group
}
