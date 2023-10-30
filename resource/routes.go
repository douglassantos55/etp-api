package resource

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service Service) {
	group := e.Group("/resources")

	group.GET("/", func(c echo.Context) error {
		resources, err := service.GetAll()
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
		resource, err := service.GetById(id)
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
		_, err := service.CreateResource(resource)
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

		resource, err := service.GetById(id)
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

		resource, err = service.UpdateResource(resource)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, resource)
	})
}
