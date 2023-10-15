package resource

import (
	"api/database"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Resource struct {
	Id    uint64  `db:"id" json:"id" goqu:"skipinsert,skipupdate"`
	Name  string  `db:"name" json:"name"`
	Image *string `db:"image" json:"image"`
}

func CreateEndpoints(e *echo.Echo, conn *database.Connection) {
	group := e.Group("/resources")
	repository := NewRepository(conn)

	group.POST("/", func(c echo.Context) error {
		var resource *Resource
		if err := c.Bind(&resource); err != nil {
			return err
		}
		resource, err := repository.SaveResource(resource)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, resource)
	})
}
