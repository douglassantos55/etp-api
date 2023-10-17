package resource

import (
	"api/database"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Resource struct {
	Id    uint64  `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
	Name  string  `db:"name" json:"name" validate:"required"`
	Image *string `db:"image" json:"image"`
}

func CreateEndpoints(e *echo.Echo, conn *database.Connection) {
	group := e.Group("/resources")
	repository := NewRepository(conn)

	group.GET("/", func(c echo.Context) error {
		resources, err := repository.FetchResources()
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
		resource, err := repository.GetById(id)
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
		_, err := repository.SaveResource(resource)
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

		resource, err := repository.GetById(id)
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

		resource, err = repository.UpdateResource(resource)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, resource)
	})
}
