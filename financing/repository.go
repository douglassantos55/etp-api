package financing

import "api/database"

type (
	Repository interface {
	}

	goquRepository struct {
	}
)

func NewRepository(conn *database.Connection) Repository {
	return &goquRepository{}
}
