package warehouse

import "database/sql"

type Resource struct {
	Id      uint64         `db:"id"`
	Name    string         `db:"name"`
	Qty     uint64         `db:"quantity"`
	Quality uint8          `db:"quality"`
	Cost    float64        `db:"sourcing_cost"`
	Image   sql.NullString `db:"image"`
}
