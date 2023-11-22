package market

import (
	"api/auth"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service Service) {
	group := e.Group("/market")

	group.POST("/orders", func(c echo.Context) error {
		order := new(Order)

		if err := c.Bind(order); err != nil {
			return err
		}

		if err := c.Validate(order); err != nil {
			return err
		}

		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		order.CompanyId = companyId

		order, err = service.PlaceOrder(c.Request().Context(), order)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, order)
	})
}
