package market

import (
	"api/auth"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service Service) {
	group := e.Group("/market")

	group.GET("/orders", func(c echo.Context) error {
		resourceId, err := strconv.ParseUint(c.QueryParam("resource"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		quality, err := strconv.ParseUint(c.QueryParam("quality"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		orders, err := service.GetByResource(c.Request().Context(), resourceId, quality)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, orders)
	})

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

	group.POST("/orders/purchase", func(c echo.Context) error {
		purchase := new(Purchase)

		if err := c.Bind(purchase); err != nil {
			return err
		}

		if err := c.Validate(purchase); err != nil {
			return err
		}

		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		items, err := service.Purchase(c.Request().Context(), purchase, companyId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, items)
	})
}
