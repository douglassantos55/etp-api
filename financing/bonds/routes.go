package bonds

import (
	"api/auth"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(group *echo.Group, service Service) {
	group.GET("/bonds", func(c echo.Context) error {
		companyId, err := strconv.ParseInt(c.QueryParam("company"), 10, 64)
		if err == nil {
			bonds, err := service.GetCompanyBonds(c.Request().Context(), companyId)
			if err != nil {
				return err
			}
			return c.JSON(http.StatusOK, bonds)
		}

		page, err := strconv.ParseUint(c.QueryParam("page"), 10, 64)
		if err != nil {
			page = 1
		}

		limit, err := strconv.ParseUint(c.QueryParam("limit"), 10, 64)
		if err != nil {
			limit = 50
		}

		bonds, err := service.GetBonds(c.Request().Context(), uint(page-1), uint(limit))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, bonds)
	})

	group.POST("/bonds", func(c echo.Context) error {
		request := struct {
			Rate   float64 `json:"rate" validate:"required"`
			Amount int64   `json:"amount" validate:"required"`
		}{}

		if err := c.Bind(&request); err != nil {
			return err
		}

		if err := c.Validate(&request); err != nil {
			return err
		}

		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		bond, err := service.EmitBond(c.Request().Context(), request.Rate, request.Amount, int64(companyId))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, bond)
	})

	group.POST("/bonds/:bondId", func(c echo.Context) error {
		bondId, err := strconv.ParseInt(c.Param("bondId"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		request := struct {
			Amount     int64 `json:"amount" validate:"required"`
			CreditorId int64 `json:"creditor_id" validate:"required"`
		}{}

		if err := c.Bind(&request); err != nil {
			return err
		}

		if err := c.Validate(&request); err != nil {
			return err
		}

		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		creditor, err := service.BuyBackBond(c.Request().Context(), request.Amount, bondId, request.CreditorId, int64(companyId))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, creditor)
	})
}
