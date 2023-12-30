package loans

import (
	"api/auth"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(group *echo.Group, service Service) {
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
