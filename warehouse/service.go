package warehouse

import (
	"api/database"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Resource struct {
	Id      uint64         `db:"id"`
	Name    string         `db:"name"`
	Qty     uint64         `db:"quantity"`
	Quality uint8          `db:"quality"`
	Cost    float64        `db:"sourcing_cost"`
	Image   sql.NullString `db:"image"`
}

func CreateEndpoints(e *echo.Echo, conn *database.Connection) {
	group := e.Group("/warehouse")
	repository := NewRepository(conn)

	group.GET("/:company", func(c echo.Context) error {
		companyId, err := strconv.ParseUint(c.Param("company"), 10, 64)
		if err != nil {
			return err
		}
		resources, err := repository.FetchInventory(companyId)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, resources)
	})
}
