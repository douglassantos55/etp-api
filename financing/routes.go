package financing

import (
	"api/financing/bonds"
	"api/financing/loans"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, financingSvc Service, loansSvc loans.Service, bondsSvc bonds.Service) {
	group := e.Group("/financing")

	group.GET("/inflation", func(c echo.Context) error {
		start, err := time.Parse(time.DateOnly, c.QueryParam("start"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		end, err := time.Parse(time.DateOnly, c.QueryParam("end"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		inflation, _, err := financingSvc.GetInflation(c.Request().Context(), start, end)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]float64{"inflation": inflation})
	})

	loans.CreateEndpoints(group, loansSvc)
	bonds.CreateEndpoints(group, bondsSvc)
}
