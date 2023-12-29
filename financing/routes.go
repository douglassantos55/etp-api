package financing

import (
	"api/auth"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service Service) {
	group := e.Group("/financing")

	createLoanEndpoints(group, service)
	createBondEndpoints(group, service)
}

func createLoanEndpoints(group *echo.Group, service Service) {
	group.GET("/loans", func(c echo.Context) error {
		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		loans, err := service.GetLoans(c.Request().Context(), int64(companyId))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, loans)
	})

	group.POST("/loans", func(c echo.Context) error {
		request := struct {
			Amount int64 `json:"amount" validate:"required"`
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

		loan, err := service.TakeLoan(c.Request().Context(), request.Amount, int64(companyId))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, loan)
	})

	group.POST("/loans/:loanId", func(c echo.Context) error {
		loanId, err := strconv.ParseInt(c.Param("loanId"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		request := struct {
			Amount int64 `json:"amount" validate:"required"`
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

		loan, err := service.BuyBackLoan(c.Request().Context(), request.Amount, loanId, int64(companyId))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, loan)
	})
}

func createBondEndpoints(group *echo.Group, service Service) {
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
}
