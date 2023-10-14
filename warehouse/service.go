package warehouse

import (
	"api/database"
	"api/resource"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

	Cost    float64 `db:"sourcing_cost" json:"cost"`
type StockItem struct {
	Qty      uint64             `db:"quantity" json:"quantity"`
	Quality  uint8              `db:"quality" json:"quality"`
	Resource *resource.Resource `db:"resource" json:"resource"`
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
