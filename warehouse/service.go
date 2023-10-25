package warehouse

import (
	"api/auth"
	"api/database"
	"api/resource"
	"net/http"

	"github.com/labstack/echo/v4"
)

type StockItem struct {
	Qty      uint64             `db:"quantity" json:"quantity"`
	Quality  uint8              `db:"quality" json:"quality"`
	Cost     uint64             `db:"sourcing_cost" json:"cost"`
	Resource *resource.Resource `db:"resource" json:"resource"`
}

func CreateEndpoints(e *echo.Echo, conn *database.Connection) {
	group := e.Group("/warehouse")
	repository := NewRepository(conn)

	group.GET("", func(c echo.Context) error {
		companyId, err := auth.ParseToken(c.Get("user"))
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
