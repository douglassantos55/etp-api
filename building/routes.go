package building

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CreateEndpoints(e *echo.Echo, service Service) {
	group := e.Group("/buildings")

	group.GET("", func(c echo.Context) error {
		buildings, err := service.GetAll()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, buildings)
	})

	group.GET("/:id", func(c echo.Context) error {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		buildings, err := service.GetById(id)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, buildings)
	})
}
