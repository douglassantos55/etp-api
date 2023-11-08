package company

import (
	"api/auth"
	"api/resource"

	"api/server"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service Service) {
	group := e.Group("/companies")

	CompanyBuildings(group, service)

	group.GET("/:id", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		company, err := service.GetById(c.Request().Context(), companyId)
		if err != nil {
			return err
		}

		if company == nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		return c.JSON(http.StatusOK, company)
	})

	group.POST("/register", func(c echo.Context) error {
		registration := new(Registration)
		if err := c.Bind(registration); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		if err := c.Validate(registration); err != nil {
			return err
		}

		company, err := service.Register(c.Request().Context(), registration)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, company)
	})

	group.POST("/login", func(c echo.Context) error {
		var credentials Credentials

		if err := c.Bind(&credentials); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		if err := c.Validate(&credentials); err != nil {
			return err
		}

		token, err := service.Login(c.Request().Context(), credentials)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, server.ValidationErrors{
				Errors: map[string]string{"email": err.Error()},
			})
		}

		return c.JSON(http.StatusOK, map[string]string{"token": token})
	})
}

func CompanyBuildings(g *echo.Group, service Service) {
	group := g.Group("/:id/buildings")

	BuildingProductions(group, service)

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
}

func BuildingProductions(g *echo.Group, service Service) {
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
		return nil
	})
}
