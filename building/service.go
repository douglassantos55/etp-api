package building

import (
	"api/database"
	"api/resource"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Building struct {
	Id              uint64               `db:"id" json:"id"`
	Name            string               `db:"name" json:"name"`
	WagesHour       uint64               `db:"wages_per_hour" json:"wages_per_hour"`
	AdminHour       uint64               `db:"admin_per_hour" json:"admin_per_hour"`
	MaintenanceHour uint64               `db:"maintenance_per_hour" json:"maintenance_per_hour"`
	Resources       []*resource.Resource `json:"resources"`
}

func CreateEndpoints(e *echo.Echo, conn *database.Connection) {
	group := e.Group("/buildings")
	repository := NewRepository(conn)

	group.GET("", func(c echo.Context) error {
		buildings, err := repository.GetAll()
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

		buildings, err := repository.GetById(id)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, buildings)
	})
}
