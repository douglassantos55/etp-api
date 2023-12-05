package research

import (
	"api/auth"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service Service) {
	group := e.Group("/research")

	group.POST("/staff/graduate", func(c echo.Context) error {
		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		duration, err := service.FindGraduate(c.Request().Context(), companyId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, duration)
	})

	group.POST("/staff/experienced", func(c echo.Context) error {
		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		duration, err := service.FindExperienced(c.Request().Context(), companyId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, duration)
	})

	group.POST("/staff/:staff/hire", func(c echo.Context) error {
		staffId, err := strconv.ParseUint(c.Param("staff"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		staff, err := service.HireStaff(c.Request().Context(), staffId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, staff)
	})

	group.POST("/staff/:staff/offer", func(c echo.Context) error {
		staffId, err := strconv.ParseUint(c.Param("staff"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		var content struct {
			Offer uint64 `json:"offer"`
		}

		if err := c.Bind(&content); err != nil {
			return err
		}

		staff, err := service.MakeOffer(c.Request().Context(), content.Offer, staffId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, staff)
	})
}
