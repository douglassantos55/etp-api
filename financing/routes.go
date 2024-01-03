package financing

import (
	"api/financing/bonds"
	"api/financing/loans"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, financingSvc Service, loansSvc loans.Service, bondsSvc bonds.Service) {
	group := e.Group("/financing")

	group.GET("/rates", func(c echo.Context) error {
		rates, err := financingSvc.GetEffectiveRates(c.Request().Context())
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, rates)
	})

	loans.CreateEndpoints(group, loansSvc)
	bonds.CreateEndpoints(group, bondsSvc)
}
