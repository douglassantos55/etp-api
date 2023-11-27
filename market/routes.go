package market

import (
	"api/auth"
	"net/http"
	"strconv"

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

	group.DELETE("/orders/:id", func(c echo.Context) error {
		orderId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		ctx := c.Request().Context()

		order, err := service.GetById(ctx, orderId)
		if err != nil {
			return err
		}

		if order == nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		if companyId != order.Company.Id {
			return echo.NewHTTPError(http.StatusForbidden)
		}

		return service.CancelOrder(ctx, order)
	})
}
