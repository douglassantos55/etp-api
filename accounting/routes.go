package accounting

import (
	"api/auth"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service Service) {
	group := e.Group("/accounting")

	group.POST("/taxes", func(c echo.Context) error {
		id, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		if id != 0 {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		start, end := GetCurrentPeriod()
		return service.PayTaxes(c.Request().Context(), start, end)
	})
}
