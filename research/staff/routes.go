package staff

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

		search, err := service.FindGraduate(c.Request().Context(), companyId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, search)
	})

	group.POST("/staff/experienced", func(c echo.Context) error {
		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		search, err := service.FindExperienced(c.Request().Context(), companyId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, search)
	})

	group.POST("/staff/:staff/hire", func(c echo.Context) error {
		staffId, err := strconv.ParseUint(c.Param("staff"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		staff, err := service.HireStaff(c.Request().Context(), staffId, companyId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, staff)
	})

	group.DELETE("/staff/searches/:search", func(c echo.Context) error {
		searchId, err := strconv.ParseUint(c.Param("search"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		return service.CancelSearch(c.Request().Context(), searchId, companyId)
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

		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		staff, err := service.MakeOffer(c.Request().Context(), content.Offer, staffId, companyId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, staff)
	})

	group.POST("/staff/:staff/raise", func(c echo.Context) error {
		staffId, err := strconv.ParseUint(c.Param("staff"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		var content struct {
			Salary uint64 `json:"salary"`
		}

		if err := c.Bind(&content); err != nil {
			return err
		}

		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		staff, err := service.IncreaseSalary(c.Request().Context(), content.Salary, staffId, companyId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, staff)
	})

	group.POST("/staff/:staff/train", func(c echo.Context) error {
		staffId, err := strconv.ParseUint(c.Param("staff"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		companyId, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		training, err := service.Train(c.Request().Context(), staffId, companyId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, training)
	})
}
