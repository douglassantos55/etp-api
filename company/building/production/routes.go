package production

import (
	"api/auth"
	"api/company"
	"api/company/building"
	"api/resource"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service ProductionService, buildingSvc building.BuildingService, companySvc company.Service) {
	g := building.CreateEndpoints(e, buildingSvc, companySvc)

	group := g.Group("/:building/productions")

	group.POST("", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		buildingId, err := strconv.ParseUint(c.Param("building"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		item := new(resource.Item)
		if err := c.Bind(item); err != nil {
			return err
		}

		if err := c.Validate(item); err != nil {
			return err
		}

		authenticated, err := auth.ParseToken(c.Get("user"))
		if err != nil {
			return err
		}

		if companyId != authenticated {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		production, err := service.Produce(c.Request().Context(), companyId, buildingId, item)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, production)
	})

	group.DELETE("/:production", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		buildingId, err := strconv.ParseUint(c.Param("building"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		productionId, err := strconv.ParseUint(c.Param("production"), 10, 64)
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

		err = service.CancelProduction(c.Request().Context(), companyId, buildingId, productionId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusNoContent, nil)
	})

	group.POST("/:production/collect", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		buildingId, err := strconv.ParseUint(c.Param("building"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		productionId, err := strconv.ParseUint(c.Param("production"), 10, 64)
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

		collected, err := service.CollectResource(
			c.Request().Context(),
			productionId,
			buildingId,
			companyId,
		)

		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, collected)
	})
}
