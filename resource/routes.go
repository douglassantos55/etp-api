package resource

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service Service) {
	group := e.Group("/resources")

	group.GET("/", func(c echo.Context) error {
		resources, err := service.GetAll(c.Request().Context())
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, resources)
	})

	group.GET("/:id", func(c echo.Context) error {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		resource, err := service.GetById(c.Request().Context(), id)
		if err != nil {
			return err
		}
		if resource == nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return c.JSON(http.StatusOK, resource)
	})

	group.POST("/", func(c echo.Context) error {
		resource := new(Resource)
		if err := c.Bind(resource); err != nil {
			return err
		}
		if err := c.Validate(resource); err != nil {
			return err
		}
		_, err := service.CreateResource(c.Request().Context(), resource)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, resource)
	})

	group.PUT("/:id", func(c echo.Context) error {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		resource, err := service.GetById(c.Request().Context(), id)
		if err != nil {
			return err
		}
		if resource == nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		if err := c.Bind(resource); err != nil {
			return err
		}
		if err := c.Validate(resource); err != nil {
			return err
		}

		resource, err = service.UpdateResource(c.Request().Context(), resource)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, resource)
	})
}
