package building

import (
	"api/auth"
	"api/company"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service BuildingService, companySvc company.Service) *echo.Group {
	g := company.CreateEndpoints(e, companySvc)

	group := g.Group("/:id/buildings")

	group.GET("", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		buildings, err := service.GetBuildings(c.Request().Context(), companyId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, buildings)
	})

	group.POST("", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		data := new(Building)
		if err := c.Bind(data); err != nil {
			return err
		}

		if err := c.Validate(data); err != nil {
			return err
		}

		authenticated, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		if companyId != authenticated {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		companyBuilding, err := service.AddBuilding(c.Request().Context(), companyId, data.BuildingId, data.Position)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, companyBuilding)
	})

	group.DELETE("/:buildingId", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		buildingId, err := strconv.ParseUint(c.Param("buildingId"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		authenticated, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		if companyId != authenticated {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		err = service.Demolish(c.Request().Context(), companyId, buildingId)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	})

	return group
}
